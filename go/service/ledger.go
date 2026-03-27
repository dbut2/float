package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

type ledgerEntry struct {
	TransactionID       uuid.UUID
	BucketID            uuid.UUID
	Description         string
	Message             string
	AmountCents         int64
	ForeignCurrencyCode sql.NullString
	ForeignAmountCents  sql.NullInt64
	CreatedAt           time.Time
	IsTransaction       bool
	CoversTransactionID uuid.NullUUID
}

type LedgerService struct {
	q database.Querier
}

func NewLedgerService(q database.Querier) *LedgerService {
	return &LedgerService{q: q}
}

func (s *LedgerService) ForUser(ctx context.Context, userID uuid.UUID) ([]ledgerEntry, error) {
	txns, err := s.q.ListUpTransactionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	transfers, err := s.q.ListTransfers(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ledgerEntry, 0, len(txns)+len(transfers)*2)
	for _, t := range txns {
		out = append(out, entryFromUpTransaction(t))
	}
	for _, t := range transfers {
		out = append(out, entryFromTransfer(t, t.ToBucketID))
		out = append(out, entryFromTransfer(t, t.FromBucketID))
	}
	return out, nil
}

func (s *LedgerService) ForBucket(ctx context.Context, bucketID uuid.UUID, userID uuid.UUID) ([]ledgerEntry, error) {
	txns, err := s.q.ListUpTransactionsByBucket(ctx, bucketID, userID)
	if err != nil {
		return nil, err
	}
	transfers, err := s.q.ListTransfersByBucket(ctx, bucketID)
	if err != nil {
		return nil, err
	}
	out := make([]ledgerEntry, 0, len(txns)+len(transfers))
	for _, t := range txns {
		out = append(out, entryFromUpTransaction(t))
	}
	for _, t := range transfers {
		row := database.ListTransfersRow{
			FromBucketID:         t.FromBucketID,
			FromBucketName:       t.FromBucketName,
			ToBucketID:           t.ToBucketID,
			ToBucketName:         t.ToBucketName,
			AmountCents:          t.AmountCents,
			Note:                 t.Note,
			CreatedAt:            t.CreatedAt,
			CoversTransactionID:  t.CoversTransactionID,
			CoveredTxDescription: t.CoveredTxDescription,
		}
		if t.ToBucketID == bucketID {
			out = append(out, entryFromTransfer(row, bucketID))
		}
		if t.FromBucketID == bucketID {
			out = append(out, entryFromTransfer(row, bucketID))
		}
	}
	return out, nil
}

func entryFromUpTransaction(t database.FloatUpTransaction) ledgerEntry {
	return ledgerEntry{
		TransactionID:       t.TransactionID,
		BucketID:            t.BucketID,
		Description:         t.Description,
		Message:             t.Message,
		AmountCents:         t.AmountCents,
		ForeignCurrencyCode: t.ForeignCurrencyCode,
		ForeignAmountCents:  t.ForeignAmountCents,
		CreatedAt:           t.CreatedAt,
		IsTransaction:       true,
		CoversTransactionID: uuid.NullUUID{},
	}
}

func entryFromTransfer(t database.ListTransfersRow, bucketID uuid.UUID) ledgerEntry {
	var desc string
	if t.CoversTransactionID.Valid {
		desc = "Cover for " + t.CoveredTxDescription.String
	} else if bucketID == t.ToBucketID {
		desc = "Transfer from " + t.FromBucketName
	} else {
		desc = "Transfer to " + t.ToBucketName
	}

	amount := t.AmountCents
	if bucketID == t.FromBucketID {
		amount = -t.AmountCents
	}

	return ledgerEntry{
		BucketID:            bucketID,
		Description:         desc,
		Message:             t.Note,
		AmountCents:         amount,
		CreatedAt:           t.CreatedAt,
		IsTransaction:       false,
		CoversTransactionID: t.CoversTransactionID,
	}
}

func sumLedgerBalance(rows []ledgerEntry) int64 {
	var total int64
	for _, r := range rows {
		total += r.AmountCents
	}
	return total
}

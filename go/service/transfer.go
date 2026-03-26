package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/utils"
)

type Transfer struct {
	TransferID     uuid.UUID `json:"transfer_id"`
	FromBucketID   uuid.UUID `json:"from_bucket_id"`
	FromBucketName string    `json:"from_bucket_name"`
	ToBucketID     uuid.UUID `json:"to_bucket_id"`
	ToBucketName   string    `json:"to_bucket_name"`
	AmountCents    int64     `json:"amount_cents"`
	DisplayAmount  string    `json:"display_amount"`
	Description    string    `json:"description"`
	Note           string    `json:"note"`
	CreatedAt      time.Time `json:"created_at"`
	DisplayDate    string    `json:"display_date"`
}

type TransferService struct {
	q database.Querier
}

func NewTransferService(q database.Querier) *TransferService {
	return &TransferService{q: q}
}

func (s *TransferService) ListTransfers(ctx context.Context, userID uuid.UUID) ([]Transfer, error) {
	rows, err := s.q.ListTransfers(ctx, userID)
	if err != nil {
		return nil, err
	}
	transfers := make([]Transfer, len(rows))
	for i, r := range rows {
		description := "Transfer to " + r.ToBucketName
		if r.CoversTransactionID.Valid && r.CoveredTxDescription.Valid {
			description = "Cover for " + r.CoveredTxDescription.String
		}
		transfers[i] = Transfer{
			TransferID:     r.TransferID,
			FromBucketID:   r.FromBucketID,
			FromBucketName: r.FromBucketName,
			ToBucketID:     r.ToBucketID,
			ToBucketName:   r.ToBucketName,
			AmountCents:    r.AmountCents,
			DisplayAmount:  FormatCurrencyAmount(r.AmountCents, "AUD"),
			Description:    description,
			Note:           r.Note,
			CreatedAt:      r.CreatedAt,
			DisplayDate:    utils.FormatDate(r.CreatedAt),
		}
	}
	return transfers, nil
}

func (s *TransferService) CreateTransfer(ctx context.Context, transfer Transfer) (Transfer, error) {
	t, err := s.q.CreateTransfer(ctx, database.CreateTransferParams{
		FromBucketID: transfer.FromBucketID,
		ToBucketID:   transfer.ToBucketID,
		AmountCents:  transfer.AmountCents,
		Note:         transfer.Note,
	})
	if err != nil {
		return Transfer{}, err
	}
	return Transfer{
		TransferID:    t.TransferID,
		FromBucketID:  t.FromBucketID,
		ToBucketID:    t.ToBucketID,
		AmountCents:   t.AmountCents,
		DisplayAmount: FormatCurrencyAmount(t.AmountCents, "AUD"),
		Note:          t.Note,
		CreatedAt:     t.CreatedAt,
		DisplayDate:   utils.FormatDate(t.CreatedAt),
	}, nil
}

func (s *TransferService) DeleteTransfer(ctx context.Context, transferID, userID uuid.UUID) error {
	rows, err := s.q.DeleteTransfer(ctx, transferID, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/utils"
)

type Cover struct {
	CoverID       uuid.UUID `json:"cover_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	AmountCents   int64     `json:"amount_cents"`
	DisplayAmount string    `json:"display_amount"`
	Note          string    `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
	DisplayDate   string    `json:"display_date"`
}

type CoverService struct {
	q database.Querier
}

func NewCoverService(q database.Querier) *CoverService {
	return &CoverService{q: q}
}

func (s *CoverService) CreateCover(ctx context.Context, userID, transactionID uuid.UUID, amountCents int64, note string) (Cover, error) {
	tx, err := s.q.GetTransactionOwner(ctx, transactionID, userID)
	if err != nil {
		return Cover{}, ErrNotFound
	}

	if tx.AmountCents >= 0 {
		return Cover{}, errors.New("only debit transactions can be covered")
	}

	existingCovers, err := s.q.ListCoversByTransaction(ctx, uuid.NullUUID{UUID: transactionID, Valid: true})
	if err != nil {
		return Cover{}, err
	}
	var totalCovered int64
	for _, c := range existingCovers {
		totalCovered += c.AmountCents
	}
	maxCoverable := -tx.AmountCents
	if totalCovered+amountCents > maxCoverable {
		return Cover{}, errors.New("cover amount exceeds transaction amount")
	}

	generalBucket, err := s.q.GetGeneralBucket(ctx, userID)
	if err != nil {
		return Cover{}, err
	}

	transfer, err := s.q.CreateCover(ctx, database.CreateCoverParams{
		FromBucketID:        generalBucket.BucketID,
		ToBucketID:          tx.BucketID,
		AmountCents:         amountCents,
		Note:                note,
		CoversTransactionID: uuid.NullUUID{UUID: transactionID, Valid: true},
	})
	if err != nil {
		return Cover{}, err
	}

	return Cover{
		CoverID:       transfer.TransferID,
		TransactionID: transactionID,
		AmountCents:   transfer.AmountCents,
		DisplayAmount: utils.FormatSignedAmount(transfer.AmountCents, "AUD"),
		Note:          transfer.Note,
		CreatedAt:     transfer.CreatedAt,
		DisplayDate:   utils.FormatDate(transfer.CreatedAt),
	}, nil
}

func (s *CoverService) ListByBucket(ctx context.Context, bucketID uuid.UUID) (map[uuid.UUID][]Cover, error) {
	rows, err := s.q.ListCoversByBucket(ctx, bucketID)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID][]Cover)
	for _, r := range rows {
		if !r.CoversTransactionID.Valid {
			continue
		}
		txID := r.CoversTransactionID.UUID
		result[txID] = append(result[txID], Cover{
			CoverID:       r.TransferID,
			TransactionID: txID,
			AmountCents:   r.AmountCents,
			DisplayAmount: utils.FormatSignedAmount(r.AmountCents, "AUD"),
			Note:          r.Note,
			CreatedAt:     r.CreatedAt,
			DisplayDate:   utils.FormatDate(r.CreatedAt),
		})
	}
	return result, nil
}

func (s *CoverService) DeleteCover(ctx context.Context, coverID, userID uuid.UUID) error {
	rows, err := s.q.DeleteCover(ctx, coverID, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

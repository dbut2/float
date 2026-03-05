package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

type Transfer struct {
	TransferID     uuid.UUID `json:"transfer_id"`
	FromBucketID   uuid.UUID `json:"from_bucket_id"`
	FromBucketName string    `json:"from_bucket_name"`
	ToBucketID     uuid.UUID `json:"to_bucket_id"`
	ToBucketName   string    `json:"to_bucket_name"`
	AmountCents    int64     `json:"amount_cents"`
	DisplayAmount  string    `json:"display_amount"`
	Note           string    `json:"note"`
	CreatedAt      time.Time `json:"created_at"`
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
		transfers[i] = Transfer{
			TransferID:     r.TransferID,
			FromBucketID:   r.FromBucketID,
			FromBucketName: r.FromBucketName,
			ToBucketID:     r.ToBucketID,
			ToBucketName:   r.ToBucketName,
			AmountCents:    r.AmountCents,
			DisplayAmount:  FormatCurrencyAmount(r.AmountCents, "AUD"),
			Note:           r.Note,
			CreatedAt:      r.CreatedAt,
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

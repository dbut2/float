package seed

import (
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/service"
)

type TransferOption func(*service.Transfer)

func WithNote(s string) TransferOption {
	return func(t *service.Transfer) {
		t.Note = s
	}
}

func CreateTransfer(from, to uuid.UUID, amountCents int64, opts ...TransferOption) service.Transfer {
	t := service.Transfer{
		TransferID:   uuid.New(),
		FromBucketID: from,
		ToBucketID:   to,
		AmountCents:  amountCents,
		CreatedAt:    time.Now(),
	}
	for _, opt := range opts {
		opt(&t)
	}
	t.DisplayAmount = service.FormatCurrencyAmount(t.AmountCents, "AUD")
	return t
}

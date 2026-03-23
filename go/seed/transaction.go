package seed

import (
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/service"
)

type TransactionOption func(*service.Transaction)

func WithTransactionID(id uuid.UUID) TransactionOption {
	return func(t *service.Transaction) {
		t.TransactionID = id
	}
}

func WithDescription(s string) TransactionOption {
	return func(t *service.Transaction) {
		t.Description = s
	}
}

func WithMessage(s string) TransactionOption {
	return func(t *service.Transaction) {
		t.Message = s
	}
}

func WithForeign(code string, foreignCents int64) TransactionOption {
	return func(t *service.Transaction) {
		t.ForeignCurrencyCode = &code
		t.ForeignAmountCents = &foreignCents
	}
}

func At(s string) TransactionOption {
	return func(t *service.Transaction) {
		parsed, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic("seed.At: invalid time string: " + s)
		}
		t.CreatedAt = parsed
	}
}

func CreateDeposit(bucketID uuid.UUID, amountCents int64, opts ...TransactionOption) service.Transaction {
	if amountCents < 0 {
		amountCents = -amountCents
	}
	tx := newTransaction(bucketID, amountCents, "Deposit")
	for _, opt := range opts {
		opt(&tx)
	}
	tx.DisplayAmount = service.FormatSignedAmount(tx.AmountCents, "AUD")
	return tx
}

func CreateExpense(bucketID uuid.UUID, amountCents int64, opts ...TransactionOption) service.Transaction {
	if amountCents > 0 {
		amountCents = -amountCents
	}
	tx := newTransaction(bucketID, amountCents, "Expense")
	for _, opt := range opts {
		opt(&tx)
	}
	tx.DisplayAmount = service.FormatSignedAmount(tx.AmountCents, "AUD")
	return tx
}

func CreateTransaction(bucketID uuid.UUID, amountCents int64, opts ...TransactionOption) service.Transaction {
	tx := newTransaction(bucketID, amountCents, "")
	for _, opt := range opts {
		opt(&tx)
	}
	tx.DisplayAmount = service.FormatSignedAmount(tx.AmountCents, "AUD")
	return tx
}

func newTransaction(bucketID uuid.UUID, amountCents int64, description string) service.Transaction {
	return service.Transaction{
		TransactionID: uuid.New(),
		BucketID:      bucketID,
		Description:   description,
		AmountCents:   amountCents,
		CreatedAt:     time.Now(),
		IsTransaction: true,
	}
}

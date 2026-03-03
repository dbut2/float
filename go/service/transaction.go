package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

type Transaction struct {
	TransactionID   uuid.UUID       `json:"transaction_id"`
	BucketID        uuid.UUID       `json:"bucket_id"`
	UserID          uuid.UUID       `json:"-"`
	Description     string          `json:"description,omitempty"`
	Message         string          `json:"message,omitempty"`
	AmountCents     int64           `json:"amount_cents"`
	DisplayAmount   string          `json:"display_amount,omitempty"`
	CurrencyCode    string          `json:"currency_code,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	IsTransaction   bool            `json:"is_transaction"`
	TransactionType *string         `json:"transaction_type,omitempty"`
	DeepLinkUrl     string          `json:"deep_link_url,omitempty"`
	RawJson         json.RawMessage `json:"raw_json,omitempty"`
}

type TransactionService struct {
	q database.Querier
}

func NewTransactionService(q database.Querier) *TransactionService {
	return &TransactionService{q: q}
}

func (s *TransactionService) ListTransactions(ctx context.Context, userID uuid.UUID) ([]Transaction, error) {
	rows, err := s.q.ListTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toTransactions(rows), nil
}

func (s *TransactionService) AssignToBucket(ctx context.Context, transactionID, bucketID uuid.UUID) error {
	return s.q.AssignTransactionToBucket(ctx, transactionID, bucketID)
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tx Transaction) (Transaction, error) {
	var txType sql.NullString
	if tx.TransactionType != nil {
		txType = sql.NullString{String: *tx.TransactionType, Valid: true}
	}
	_, err := s.q.UpsertUpTransaction(ctx, database.UpsertUpTransactionParams{
		TransactionID:   tx.TransactionID,
		UserID:          tx.UserID,
		Description:     tx.Description,
		Message:         tx.Message,
		AmountCents:     tx.AmountCents,
		DisplayAmount:   tx.DisplayAmount,
		CurrencyCode:    tx.CurrencyCode,
		CreatedAt:       tx.CreatedAt,
		TransactionType: txType,
		DeepLinkUrl:     tx.DeepLinkUrl,
		RawJson:         tx.RawJson,
	})
	if err != nil {
		return Transaction{}, err
	}
	return tx, nil
}

func toTransactions(rows []database.FloatBucketLedger) []Transaction {
	txs := make([]Transaction, len(rows))
	for i, r := range rows {
		txs[i] = Transaction{
			TransactionID: r.TransactionID,
			BucketID:      r.BucketID,
			AmountCents:   r.AmountCents,
			CreatedAt:     r.CreatedAt,
			IsTransaction: r.IsTransaction,
		}
	}
	return txs
}

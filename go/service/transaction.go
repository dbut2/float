package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/utils"
)

type Transaction struct {
	TransactionID        uuid.UUID       `json:"transaction_id"`
	BucketID             uuid.UUID       `json:"bucket_id"`
	UserID               uuid.UUID       `json:"-"`
	Description          string          `json:"description,omitempty"`
	Message              string          `json:"message,omitempty"`
	AmountCents          int64           `json:"amount_cents"`
	DisplayAmount        string          `json:"display_amount"`
	CreatedAt            time.Time       `json:"created_at"`
	DisplayDate          string          `json:"display_date"`
	IsTransaction        bool            `json:"is_transaction"`
	TransactionType      *string         `json:"transaction_type,omitempty"`
	RawJson              json.RawMessage `json:"raw_json,omitempty"`
	ForeignCurrencyCode  *string         `json:"foreign_currency_code,omitempty"`
	ForeignAmountCents   *int64          `json:"foreign_amount_cents,omitempty"`
	ForeignDisplayAmount *string         `json:"foreign_display_amount,omitempty"`
	Covers               []Cover         `json:"covers,omitempty"`
	CoversAmountCents    int64           `json:"covers_amount_cents"`
	NetAmountCents       int64           `json:"net_amount_cents"`
	NetDisplayAmount     string          `json:"net_display_amount"`
}

type TransactionService struct {
	q      database.Querier
	ledger *LedgerService
}

func NewTransactionService(q database.Querier, ledger *LedgerService) *TransactionService {
	return &TransactionService{q: q, ledger: ledger}
}

func (s *TransactionService) ListTransactions(ctx context.Context, userID uuid.UUID) ([]Transaction, error) {
	entries, err := s.ledger.ForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toTransactions(entries), nil
}

func (s *TransactionService) AssignToBucket(ctx context.Context, transactionID, bucketID, userID uuid.UUID) error {
	return s.q.AssignTransactionToBucket(ctx, transactionID, bucketID, userID)
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tx Transaction) (Transaction, error) {
	var txType sql.NullString
	if tx.TransactionType != nil {
		txType = sql.NullString{String: *tx.TransactionType, Valid: true}
	}
	var foreignCurrencyCode sql.NullString
	var foreignAmountCents sql.NullInt64
	if tx.ForeignCurrencyCode != nil {
		foreignCurrencyCode = sql.NullString{String: *tx.ForeignCurrencyCode, Valid: true}
	}
	if tx.ForeignAmountCents != nil {
		foreignAmountCents = sql.NullInt64{Int64: *tx.ForeignAmountCents, Valid: true}
	}
	_, err := s.q.UpsertUpTransaction(ctx, database.UpsertUpTransactionParams{
		TransactionID:       tx.TransactionID,
		UserID:              tx.UserID,
		Description:         tx.Description,
		Message:             tx.Message,
		AmountCents:         tx.AmountCents,
		CreatedAt:           tx.CreatedAt,
		TransactionType:     txType,
		RawJson:             tx.RawJson,
		ForeignCurrencyCode: foreignCurrencyCode,
		ForeignAmountCents:  foreignAmountCents,
	})
	if err != nil {
		return Transaction{}, err
	}
	return tx, nil
}

func toTransactions(rows []ledgerEntry) []Transaction {
	txs := make([]Transaction, len(rows))
	for i, r := range rows {
		txs[i] = Transaction{
			TransactionID: r.TransactionID,
			BucketID:      r.BucketID,
			Description:   r.Description,
			Message:       r.Message,
			AmountCents:   r.AmountCents,
			DisplayAmount: utils.FormatSignedAmount(r.AmountCents, "AUD"),
			CreatedAt:     r.CreatedAt,
			DisplayDate:   utils.FormatDate(r.CreatedAt),
			IsTransaction: r.IsTransaction,
		}
		if r.ForeignCurrencyCode.Valid && r.ForeignAmountCents.Valid {
			txs[i].ForeignCurrencyCode = &r.ForeignCurrencyCode.String
			txs[i].ForeignAmountCents = &r.ForeignAmountCents.Int64
			display := utils.FormatSignedAmount(r.ForeignAmountCents.Int64, r.ForeignCurrencyCode.String)
			txs[i].ForeignDisplayAmount = &display
		}
	}
	return txs
}

// FormatCurrencyAmount delegates to utils.FormatAmount.
func FormatCurrencyAmount(baseUnits int64, currencyCode string) string {
	return utils.FormatAmount(baseUnits, currencyCode)
}

// FormatSignedAmount delegates to utils.FormatAmount.
func FormatSignedAmount(baseUnits int64, currencyCode string) string {
	return utils.FormatAmount(baseUnits, currencyCode)
}

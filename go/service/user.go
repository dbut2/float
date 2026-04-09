package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/up"
)

type User struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	HasToken  bool      `json:"has_token"`
}

type UserService struct {
	q          database.Querier
	classifier *ClassifierService
}

func NewUserService(q database.Querier, classifier *ClassifierService) *UserService {
	return &UserService{q: q, classifier: classifier}
}

func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (User, error) {
	u, err := s.q.GetUserByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return User{
		UserID:    u.UserID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		HasToken:  u.UpToken.Valid && u.UpToken.String != "",
	}, nil
}

func (s *UserService) UpdateToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.q.SetUserToken(ctx, userID, sql.NullString{String: token, Valid: true})
}

func (s *UserService) GetTransactBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	nullToken, err := s.q.GetUserToken(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !nullToken.Valid {
		return 0, ErrTokenNotSet
	}

	client, err := up.NewUpClient(nullToken.String)
	if err != nil {
		return 0, err
	}

	acct, err := client.GetTransactAccount(ctx)
	if err != nil {
		return 0, err
	}

	return int64(acct.Attributes.Balance.ValueInBaseUnits), nil
}

func (s *UserService) SyncTransactions(ctx context.Context, userID uuid.UUID) (int, error) {
	nullToken, err := s.q.GetUserToken(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !nullToken.Valid {
		return 0, ErrTokenNotSet
	}

	client, err := up.NewUpClient(nullToken.String)
	if err != nil {
		return 0, err
	}

	accounts, err := client.ListAccounts(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, account := range accounts {
		txs, err := client.ListTransactions(ctx, account.Id)
		if err != nil {
			return 0, err
		}

		for _, tx := range txs {
			params, err := upTransactionToParams(userID, tx)
			if err != nil {
				continue
			}

			inserted, err := s.q.UpsertUpTransaction(ctx, params)
			if err != nil {
				return 0, err
			}

			if inserted {
				count++
				if s.classifier != nil {
					if err := s.classifier.ClassifyTransaction(ctx, userID, params.TransactionID, params.Description, params.AmountCents, params.TransactionType, params.CategoryID, params.CreatedAt, params.ForeignCurrencyCode); err != nil {
						log.Printf("classify tx %s: %v", params.TransactionID, err)
					}
				}
			}
		}
	}

	return count, nil
}

func upTransactionToParams(userID uuid.UUID, tx up.TransactionResource) (database.UpsertUpTransactionParams, error) {
	txID, err := uuid.Parse(tx.Id)
	if err != nil {
		return database.UpsertUpTransactionParams{}, err
	}

	msg := ""
	if tx.Attributes.Message != nil {
		msg = *tx.Attributes.Message
	}

	var txType sql.NullString
	if tx.Attributes.TransactionType != nil {
		txType = sql.NullString{String: *tx.Attributes.TransactionType, Valid: true}
	}

	var categoryID sql.NullString
	if tx.Relationships.Category.Data != nil {
		categoryID = sql.NullString{String: tx.Relationships.Category.Data.Id, Valid: true}
	}

	rawJSON, err := json.Marshal(tx)
	if err != nil {
		return database.UpsertUpTransactionParams{}, err
	}

	var foreignCurrencyCode sql.NullString
	var foreignAmountCents sql.NullInt64
	if fa := tx.Attributes.ForeignAmount; fa != nil {
		foreignCurrencyCode = sql.NullString{String: fa.CurrencyCode, Valid: true}
		foreignAmountCents = sql.NullInt64{Int64: int64(fa.ValueInBaseUnits), Valid: true}
	}

	return database.UpsertUpTransactionParams{
		TransactionID:       txID,
		UserID:              userID,
		Description:         tx.Attributes.Description,
		Message:             msg,
		AmountCents:         int64(tx.Attributes.Amount.ValueInBaseUnits),
		CreatedAt:           tx.Attributes.CreatedAt,
		TransactionType:     txType,
		RawJson:             rawJSON,
		CategoryID:          categoryID,
		ForeignCurrencyCode: foreignCurrencyCode,
		ForeignAmountCents:  foreignAmountCents,
	}, nil
}

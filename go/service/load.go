package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

func LoadData(ctx context.Context, queries database.Querier, user User, buckets []Bucket, txs []Transaction, transfers []Transfer, trickles []Trickle) (uuid.UUID, error) {
	dbUser, err := queries.SeedUser(ctx, user.UserID, user.Email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("seed user: %w", err)
	}
	userID := dbUser.UserID

	var generalBucketID uuid.UUID
	for _, b := range buckets {
		if b.IsGeneral {
			generalBucketID = b.BucketID
		}
		cc := sql.NullString{}
		if b.CurrencyCode != nil {
			cc = sql.NullString{String: *b.CurrencyCode, Valid: true}
		}
		_, err := queries.SeedBucket(ctx, database.SeedBucketParams{
			BucketID:     b.BucketID,
			UserID:       userID,
			Name:         b.Name,
			IsGeneral:    b.IsGeneral,
			CurrencyCode: cc,
			Description:  b.Description,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("seed bucket %s: %w", b.Name, err)
		}
	}

	for _, tx := range txs {
		fcc := sql.NullString{}
		if tx.ForeignCurrencyCode != nil {
			fcc = sql.NullString{String: *tx.ForeignCurrencyCode, Valid: true}
		}
		fac := sql.NullInt64{}
		if tx.ForeignAmountCents != nil {
			fac = sql.NullInt64{Int64: *tx.ForeignAmountCents, Valid: true}
		}
		_, err := queries.UpsertUpTransaction(ctx, database.UpsertUpTransactionParams{
			TransactionID:       tx.TransactionID,
			UserID:              userID,
			Description:         tx.Description,
			Message:             tx.Message,
			AmountCents:         tx.AmountCents,
			CreatedAt:           tx.CreatedAt,
			RawJson:             json.RawMessage("{}"),
			ForeignCurrencyCode: fcc,
			ForeignAmountCents:  fac,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("upsert transaction %s: %w", tx.Description, err)
		}
		if tx.BucketID != generalBucketID {
			if err := queries.AssignTransactionToBucket(ctx, tx.TransactionID, tx.BucketID, userID); err != nil {
				return uuid.Nil, fmt.Errorf("assign transaction to bucket: %w", err)
			}
		}
	}

	for _, t := range transfers {
		_, err := queries.CreateTransfer(ctx, database.CreateTransferParams{
			FromBucketID: t.FromBucketID,
			ToBucketID:   t.ToBucketID,
			AmountCents:  t.AmountCents,
			Note:         t.Note,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("create transfer: %w", err)
		}
	}

	for _, t := range trickles {
		endDate := sql.NullTime{}
		if t.EndDate != nil {
			endDate = sql.NullTime{Time: *t.EndDate, Valid: true}
		}
		_, err := queries.InsertTrickle(ctx, database.InsertTrickleParams{
			FromBucketID: t.FromBucketID,
			ToBucketID:   t.ToBucketID,
			AmountCents:  t.AmountCents,
			Description:  t.Description,
			Period:       t.Period,
			StartDate:    t.StartDate,
			EndDate:      endDate,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert trickle: %w", err)
		}
	}

	return userID, nil
}

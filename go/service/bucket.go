package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/utils"
)

type Bucket struct {
	BucketID              uuid.UUID `json:"bucket_id"`
	UserID                uuid.UUID `json:"user_id"`
	Name                  string    `json:"name"`
	IsGeneral             bool      `json:"is_general"`
	CreatedAt             time.Time `json:"created_at"`
	BalanceCents          int64     `json:"balance_cents"`
	BalanceDisplay        string    `json:"balance_display"`
	DisplayOrder          *int      `json:"display_order,omitempty"`
	CurrencyCode          *string   `json:"currency_code,omitempty"`
	FXRate                *float64  `json:"fx_rate,omitempty"`
	ForeignBalanceDisplay *string   `json:"foreign_balance_display,omitempty"`
}

type BucketService struct {
	q  database.Querier
	fx *FXService
}

func NewBucketService(q database.Querier, fx *FXService) *BucketService {
	return &BucketService{q: q, fx: fx}
}

func (s *BucketService) ListBuckets(ctx context.Context, userID uuid.UUID) ([]Bucket, error) {
	rows, err := s.q.ListBuckets(ctx, userID)
	if err != nil {
		return nil, err
	}
	buckets := make([]Bucket, len(rows))
	for i, r := range rows {
		var displayOrder *int
		if r.DisplayOrder.Valid {
			v := int(r.DisplayOrder.Int32)
			displayOrder = &v
		}
		var currencyCode *string
		if r.CurrencyCode.Valid {
			currencyCode = &r.CurrencyCode.String
		}
		buckets[i] = Bucket{
			BucketID:     r.BucketID,
			UserID:       r.UserID,
			Name:         r.Name,
			IsGeneral:    r.IsGeneral,
			CreatedAt:    r.CreatedAt,
			BalanceCents: r.BalanceCents,
			DisplayOrder: displayOrder,
			CurrencyCode: currencyCode,
		}
	}

	trickleRows, _ := s.q.ListTrickles(ctx, userID)
	byID := make(map[uuid.UUID]*Bucket, len(buckets))
	for i := range buckets {
		byID[buckets[i].BucketID] = &buckets[i]
	}
	now := time.Now()
	for _, r := range trickleRows {
		t := dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt)
		amount := trickleAmount(t, now)
		if b := byID[t.ToBucketID]; b != nil {
			b.BalanceCents += amount
		}
		if b := byID[t.FromBucketID]; b != nil {
			b.BalanceCents -= amount
		}
	}

	for i := range buckets {
		if buckets[i].CurrencyCode != nil && s.fx != nil {
			if rate, err := s.fx.GetRate(ctx, "AUD", *buckets[i].CurrencyCode); err == nil {
				buckets[i].FXRate = &rate
				display := utils.FormatForeignBalance(buckets[i].BalanceCents, rate, *buckets[i].CurrencyCode)
				buckets[i].ForeignBalanceDisplay = &display
			}
		}
		buckets[i].setDisplays()
	}

	return buckets, nil
}

func (s *BucketService) CreateBucket(ctx context.Context, bucket Bucket) (Bucket, error) {
	var currencyCode sql.NullString
	if bucket.CurrencyCode != nil {
		currencyCode = sql.NullString{String: *bucket.CurrencyCode, Valid: true}
	}
	b, err := s.q.CreateBucket(ctx, bucket.UserID, bucket.Name, currencyCode)
	if err != nil {
		return Bucket{}, err
	}
	result := Bucket{
		BucketID:  b.BucketID,
		UserID:    b.UserID,
		Name:      b.Name,
		IsGeneral: b.IsGeneral,
		CreatedAt: b.CreatedAt,
	}
	if b.CurrencyCode.Valid {
		result.CurrencyCode = &b.CurrencyCode.String
	}
	result.setDisplays()
	return result, nil
}

func (s *BucketService) GetBucket(ctx context.Context, bucketID, userID uuid.UUID) (Bucket, error) {
	b, err := s.q.GetBucket(ctx, bucketID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Bucket{}, ErrNotFound
	}
	if err != nil {
		return Bucket{}, err
	}
	bucket := Bucket{
		BucketID:     b.BucketID,
		UserID:       b.UserID,
		Name:         b.Name,
		IsGeneral:    b.IsGeneral,
		CreatedAt:    b.CreatedAt,
		BalanceCents: b.BalanceCents,
	}
	if b.CurrencyCode.Valid {
		bucket.CurrencyCode = &b.CurrencyCode.String
	}

	now := time.Now()
	if bucket.IsGeneral {
		trickleRows, _ := s.q.ListTrickles(ctx, userID)
		for _, r := range trickleRows {
			bucket.BalanceCents -= trickleAmount(dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt), now)
		}
	} else {
		r, err := s.q.GetActiveTrickleByToBucketID(ctx, bucketID, userID)
		if err == nil {
			bucket.BalanceCents += trickleAmount(dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt), now)
		}
	}

	if bucket.CurrencyCode != nil && s.fx != nil {
		if rate, err := s.fx.GetRate(ctx, "AUD", *bucket.CurrencyCode); err == nil {
			bucket.FXRate = &rate
			display := utils.FormatForeignBalance(bucket.BalanceCents, rate, *bucket.CurrencyCode)
			bucket.ForeignBalanceDisplay = &display
		}
	}
	bucket.setDisplays()

	return bucket, nil
}

func (s *BucketService) DeleteBucket(ctx context.Context, bucketID, userID uuid.UUID) error {
	if err := s.q.ReassignBucketTransactionsToGeneral(ctx, bucketID); err != nil {
		return err
	}
	return s.q.DeleteBucket(ctx, bucketID, userID)
}

func (s *BucketService) ReorderBuckets(ctx context.Context, userID uuid.UUID, bucketIDs []uuid.UUID) error {
	for i, id := range bucketIDs {
		if err := s.q.SetBucketDisplayOrder(ctx, id, sql.NullInt32{Int32: int32(i), Valid: true}, userID); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bucket) setDisplays() {
	b.BalanceDisplay = utils.FormatAmount(b.BalanceCents, "AUD")
}

func (s *BucketService) ListBucketTransactions(ctx context.Context, bucketID, userID uuid.UUID) ([]Transaction, error) {
	rows, err := s.q.ListBucketTransactions(ctx, bucketID)
	if err != nil {
		return nil, err
	}
	txns := toTransactions(rows)

	trickleRows, _ := s.q.GetTricklesByBucketID(ctx, bucketID)
	now := time.Now()
	for _, r := range trickleRows {
		t := dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt)
		var createdAt time.Time
		if t.EndDate != nil && t.EndDate.Before(now) {
			createdAt = *t.EndDate
		} else {
			createdAt = now
		}
		if t.ToBucketID == bucketID {
			amt := trickleAmount(t, now)
			txns = append(txns, Transaction{
				BucketID:      bucketID,
				Description:   "Trickle from " + t.FromBucketName,
				AmountCents:   amt,
				DisplayAmount: utils.FormatSignedAmount(amt, "AUD"),
				CreatedAt:     createdAt,
			})
		} else if t.FromBucketID == bucketID {
			amt := -trickleAmount(t, now)
			txns = append(txns, Transaction{
				BucketID:      bucketID,
				Description:   "Trickle to " + t.ToBucketName,
				AmountCents:   amt,
				DisplayAmount: utils.FormatSignedAmount(amt, "AUD"),
				CreatedAt:     createdAt,
			})
		}
	}

	sort.Slice(txns, func(i, j int) bool {
		return txns[i].CreatedAt.After(txns[j].CreatedAt)
	})

	// For travel buckets, enrich every transaction with a foreign display amount.
	// Transactions already settled in the bucket's currency keep their exact rate;
	// everything else (AUD transfers, trickles, etc.) is converted at today's rate.
	if b, err := s.q.GetBucket(ctx, bucketID, userID); err == nil && b.CurrencyCode.Valid && s.fx != nil {
		currencyCode := b.CurrencyCode.String
		if rate, err := s.fx.GetRate(ctx, "AUD", currencyCode); err == nil {
			for i := range txns {
				tx := &txns[i]
				if tx.ForeignCurrencyCode != nil && *tx.ForeignCurrencyCode == currencyCode {
					// Already has an accurate settled foreign amount — keep it.
					continue
				}
				display := utils.FormatForeignBalance(tx.AmountCents, rate, currencyCode)
				tx.ForeignCurrencyCode = &currencyCode
				tx.ForeignDisplayAmount = &display
			}
		}
	}

	return txns, nil
}

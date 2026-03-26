package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	Description           string    `json:"description"`
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
	q      database.Querier
	fx     *FXService
	covers *CoverService
}

func NewBucketService(q database.Querier, fx *FXService, covers *CoverService) *BucketService {
	return &BucketService{q: q, fx: fx, covers: covers}
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
			Description:  r.Description,
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
	now := utils.Now()
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
	b, err := s.q.CreateBucket(ctx, database.CreateBucketParams{
		UserID:       bucket.UserID,
		Name:         bucket.Name,
		CurrencyCode: currencyCode,
		Description:  bucket.Description,
	})
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
		Description:  b.Description,
		IsGeneral:    b.IsGeneral,
		CreatedAt:    b.CreatedAt,
		BalanceCents: b.BalanceCents,
	}
	if b.CurrencyCode.Valid {
		bucket.CurrencyCode = &b.CurrencyCode.String
	}

	now := utils.Now()
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

func (s *BucketService) CloseBucket(ctx context.Context, bucketID, userID uuid.UUID) error {
	bucket, err := s.GetBucket(ctx, bucketID, userID)
	if err != nil {
		return err
	}
	if bucket.IsGeneral {
		return fmt.Errorf("cannot close General bucket")
	}

	// End active trickle (reuse DeleteTrickle logic)
	existing, err := s.q.GetActiveTrickleByToBucketID(ctx, bucketID, userID)
	if err == nil {
		today := utils.Today()
		existingStart := utils.ToDate(existing.StartDate)
		if existingStart.Before(today) {
			_ = s.q.SetTrickleEndDate(ctx, existing.TrickleID, sql.NullTime{Time: today, Valid: true}, userID)
		} else {
			_, _ = s.q.DeleteTrickle(ctx, existing.TrickleID, userID)
		}
	}

	// Recalculate balance including trickle amount
	balance := bucket.BalanceCents
	if balance != 0 {
		general, err := s.q.GetGeneralBucket(ctx, userID)
		if err != nil {
			return err
		}
		var fromID, toID uuid.UUID
		var amount int64
		if balance > 0 {
			fromID = bucketID
			toID = general.BucketID
			amount = balance
		} else {
			fromID = general.BucketID
			toID = bucketID
			amount = -balance
		}
		_, err = s.q.CreateTransfer(ctx, database.CreateTransferParams{
			FromBucketID: fromID,
			ToBucketID:   toID,
			AmountCents:  amount,
			Note:         "Close bucket: " + bucket.Name,
		})
		if err != nil {
			return err
		}
	}

	return s.q.CloseBucket(ctx, bucketID, userID)
}

func (s *BucketService) ReorderBuckets(ctx context.Context, userID uuid.UUID, bucketIDs []uuid.UUID) error {
	for i, id := range bucketIDs {
		if err := s.q.SetBucketDisplayOrder(ctx, id, sql.NullInt32{Int32: int32(i), Valid: true}, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *BucketService) UpdateBucketDescription(ctx context.Context, bucketID, userID uuid.UUID, description string) error {
	return s.q.UpdateBucketDescription(ctx, bucketID, description, userID)
}

func (b *Bucket) setDisplays() {
	b.BalanceDisplay = utils.FormatAmount(b.BalanceCents, "AUD")
}

func (s *BucketService) ListBucketTransactions(ctx context.Context, bucketID, userID uuid.UUID) ([]Transaction, error) {
	rows, err := s.q.ListBucketTransactions(ctx, bucketID, userID)
	if err != nil {
		return nil, err
	}

	filtered := rows[:0]
	for _, r := range rows {
		if r.CoversTransactionID.Valid && r.AmountCents > 0 {
			continue
		}
		filtered = append(filtered, r)
	}

	txns := toTransactions(filtered)

	trickleRows, _ := s.q.GetTricklesByBucketID(ctx, bucketID)
	now := utils.Now()
	for _, r := range trickleRows {
		t := dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt)
		dates := trickleOccurrences(t, now)
		for _, d := range dates {
			if t.ToBucketID == bucketID {
				txns = append(txns, Transaction{
					BucketID:      bucketID,
					Description:   "Trickle from " + t.FromBucketName,
					AmountCents:   t.AmountCents,
					DisplayAmount: utils.FormatSignedAmount(t.AmountCents, "AUD"),
					CreatedAt:     d,
					DisplayDate:   utils.FormatDate(d),
				})
			} else if t.FromBucketID == bucketID {
				amt := -t.AmountCents
				txns = append(txns, Transaction{
					BucketID:      bucketID,
					Description:   "Trickle to " + t.ToBucketName,
					AmountCents:   amt,
					DisplayAmount: utils.FormatSignedAmount(amt, "AUD"),
					CreatedAt:     d,
					DisplayDate:   utils.FormatDate(d),
				})
			}
		}
	}

	sort.Slice(txns, func(i, j int) bool {
		return txns[i].CreatedAt.After(txns[j].CreatedAt)
	})

	if s.covers != nil {
		coversByTx, err := s.covers.ListByBucket(ctx, bucketID)
		if err == nil {
			for i := range txns {
				tx := &txns[i]
				if !tx.IsTransaction {
					continue
				}
				covers := coversByTx[tx.TransactionID]
				var total int64
				for _, c := range covers {
					total += c.AmountCents
				}
				tx.Covers = covers
				tx.CoversAmountCents = total
				tx.NetAmountCents = tx.AmountCents + total
				tx.NetDisplayAmount = utils.FormatSignedAmount(tx.NetAmountCents, "AUD")
			}
		}
	}

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

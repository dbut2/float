package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

type Bucket struct {
	BucketID     uuid.UUID `json:"bucket_id"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	IsGeneral    bool      `json:"is_general"`
	CreatedAt    time.Time `json:"created_at"`
	BalanceCents int64     `json:"balance_cents"`
	DisplayOrder *int      `json:"display_order,omitempty"`
}

type BucketService struct {
	q database.Querier
}

func NewBucketService(q database.Querier) *BucketService {
	return &BucketService{q: q}
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
		buckets[i] = Bucket{
			BucketID:     r.BucketID,
			UserID:       r.UserID,
			Name:         r.Name,
			IsGeneral:    r.IsGeneral,
			CreatedAt:    r.CreatedAt,
			BalanceCents: r.BalanceCents,
			DisplayOrder: displayOrder,
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

	return buckets, nil
}

func (s *BucketService) CreateBucket(ctx context.Context, bucket Bucket) (Bucket, error) {
	b, err := s.q.CreateBucket(ctx, bucket.UserID, bucket.Name)
	if err != nil {
		return Bucket{}, err
	}
	return Bucket{
		BucketID:  b.BucketID,
		UserID:    b.UserID,
		Name:      b.Name,
		IsGeneral: b.IsGeneral,
		CreatedAt: b.CreatedAt,
	}, nil
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

func (s *BucketService) ListBucketTransactions(ctx context.Context, bucketID uuid.UUID) ([]Transaction, error) {
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
			txns = append(txns, Transaction{
				BucketID:    bucketID,
				Description: "Trickle from " + t.FromBucketName,
				AmountCents: trickleAmount(t, now),
				CreatedAt:   createdAt,
			})
		} else if t.FromBucketID == bucketID {
			txns = append(txns, Transaction{
				BucketID:    bucketID,
				Description: "Trickle to " + t.ToBucketName,
				AmountCents: -trickleAmount(t, now),
				CreatedAt:   createdAt,
			})
		}
	}

	sort.Slice(txns, func(i, j int) bool {
		return txns[i].CreatedAt.After(txns[j].CreatedAt)
	})

	return txns, nil
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

type Trickle struct {
	TrickleID      uuid.UUID  `json:"trickle_id"`
	FromBucketID   uuid.UUID  `json:"from_bucket_id"`
	FromBucketName string     `json:"from_bucket_name"`
	ToBucketID     uuid.UUID  `json:"to_bucket_id"`
	ToBucketName   string     `json:"to_bucket_name"`
	AmountCents    int64      `json:"amount_cents"`
	Description    string     `json:"description"`
	Period         string     `json:"period"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UserID         uuid.UUID  `json:"-"`
}

var validPeriods = map[string]bool{
	"daily": true, "weekly": true, "fortnightly": true, "monthly": true,
}

type TrickleService struct {
	q database.Querier
}

func NewTrickleService(q database.Querier) *TrickleService {
	return &TrickleService{q: q}
}

func (s *TrickleService) ListTrickles(ctx context.Context, userID uuid.UUID) ([]Trickle, error) {
	rows, err := s.q.ListTrickles(ctx, userID)
	if err != nil {
		return nil, err
	}
	trickles := make([]Trickle, len(rows))
	for i, r := range rows {
		trickles[i] = dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt)
	}
	return trickles, nil
}

func (s *TrickleService) GetTrickle(ctx context.Context, toBucketID, userID uuid.UUID) (Trickle, error) {
	r, err := s.q.GetActiveTrickleByToBucketID(ctx, toBucketID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return Trickle{}, ErrNotFound
	}
	if err != nil {
		return Trickle{}, err
	}
	return dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt), nil
}

func (s *TrickleService) UpsertTrickle(ctx context.Context, trickle Trickle) (Trickle, error) {
	if trickle.AmountCents <= 0 {
		return Trickle{}, fmt.Errorf("amount_cents must be positive")
	}
	if !validPeriods[trickle.Period] {
		return Trickle{}, fmt.Errorf("invalid period: %s", trickle.Period)
	}
	tomorrow := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 1)
	if trickle.StartDate.UTC().Truncate(24 * time.Hour).Before(tomorrow) {
		return Trickle{}, fmt.Errorf("start_date must be tomorrow or later")
	}
	if trickle.EndDate != nil && trickle.EndDate.Before(trickle.StartDate) {
		return Trickle{}, fmt.Errorf("end_date must be on or after start_date")
	}

	general, err := s.q.GetGeneralBucket(ctx, trickle.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return Trickle{}, fmt.Errorf("no general bucket found for user")
	}
	if err != nil {
		return Trickle{}, err
	}

	toBucket, err := s.q.GetBucket(ctx, trickle.ToBucketID, trickle.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return Trickle{}, ErrNotFound
	}
	if err != nil {
		return Trickle{}, err
	}
	if toBucket.IsGeneral {
		return Trickle{}, fmt.Errorf("cannot set trickle to General bucket")
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	existing, err := s.q.GetActiveTrickleByToBucketID(ctx, trickle.ToBucketID, trickle.UserID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Trickle{}, err
	}
	if err == nil {
		existingStart := existing.StartDate.UTC().Truncate(24 * time.Hour)
		if existingStart.Before(today) {
			if err := s.q.SetTrickleEndDate(ctx, existing.TrickleID, sql.NullTime{Time: today, Valid: true}, trickle.UserID); err != nil {
				return Trickle{}, err
			}
		} else {
			if _, err := s.q.DeleteTrickle(ctx, existing.TrickleID, trickle.UserID); err != nil {
				return Trickle{}, err
			}
		}
	}

	var endDate sql.NullTime
	if trickle.EndDate != nil {
		endDate = sql.NullTime{Time: *trickle.EndDate, Valid: true}
	}

	t, err := s.q.InsertTrickle(ctx, database.InsertTrickleParams{
		FromBucketID: general.BucketID,
		ToBucketID:   trickle.ToBucketID,
		AmountCents:  trickle.AmountCents,
		Description:  trickle.Description,
		Period:       trickle.Period,
		StartDate:    trickle.StartDate,
		EndDate:      endDate,
	})
	if err != nil {
		return Trickle{}, err
	}
	return dbTrickleToService(t.TrickleID, t.FromBucketID, general.Name, t.ToBucketID, toBucket.Name, t.AmountCents, t.Description, t.Period, t.StartDate, t.EndDate, t.CreatedAt), nil
}

func (s *TrickleService) DeleteTrickle(ctx context.Context, toBucketID, userID uuid.UUID) error {
	existing, err := s.q.GetActiveTrickleByToBucketID(ctx, toBucketID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	existingStart := existing.StartDate.UTC().Truncate(24 * time.Hour)
	if existingStart.Before(today) {
		return s.q.SetTrickleEndDate(ctx, existing.TrickleID, sql.NullTime{Time: today, Valid: true}, userID)
	}
	_, err = s.q.DeleteTrickle(ctx, existing.TrickleID, userID)
	return err
}

func trickleAmount(t Trickle, asOf time.Time) int64 {
	asOfDate := asOf.UTC().Truncate(24 * time.Hour)
	startDate := t.StartDate.UTC().Truncate(24 * time.Hour)
	if startDate.After(asOfDate) {
		return 0
	}
	endDate := asOfDate
	if t.EndDate != nil && t.EndDate.UTC().Truncate(24*time.Hour).Before(asOfDate) {
		endDate = t.EndDate.UTC().Truncate(24 * time.Hour)
	}

	days := int64(endDate.Sub(startDate) / (24 * time.Hour))
	var count int64
	switch t.Period {
	case "daily":
		count = days + 1
	case "weekly":
		count = days/7 + 1
	case "fortnightly":
		count = days/14 + 1
	case "monthly":
		years := endDate.Year() - startDate.Year()
		months := int(endDate.Month()) - int(startDate.Month())
		count = int64(years*12+months) + 1
	}
	if count < 0 {
		count = 0
	}
	return count * t.AmountCents
}

func dbTrickleToService(trickleID, fromBucketID uuid.UUID, fromBucketName string, toBucketID uuid.UUID, toBucketName string, amountCents int64, description, period string, startDate time.Time, endDate sql.NullTime, createdAt time.Time) Trickle {
	t := Trickle{
		TrickleID:      trickleID,
		FromBucketID:   fromBucketID,
		FromBucketName: fromBucketName,
		ToBucketID:     toBucketID,
		ToBucketName:   toBucketName,
		AmountCents:    amountCents,
		Description:    description,
		Period:         period,
		StartDate:      startDate,
		CreatedAt:      createdAt,
	}
	if endDate.Valid {
		t.EndDate = &endDate.Time
	}
	return t
}

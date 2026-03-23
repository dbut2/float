package seed

import (
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/service"
)

type TrickleOption func(*service.Trickle)

func WithTrickleDescription(s string) TrickleOption {
	return func(t *service.Trickle) {
		t.Description = s
	}
}

func WithEndDate(end time.Time) TrickleOption {
	return func(t *service.Trickle) {
		t.EndDate = &end
	}
}

func CreateTrickle(fromBucketID, toBucketID uuid.UUID, amountCents int64, period string, startDate time.Time, opts ...TrickleOption) service.Trickle {
	t := service.Trickle{
		TrickleID:    uuid.New(),
		FromBucketID: fromBucketID,
		ToBucketID:   toBucketID,
		AmountCents:  amountCents,
		Period:       period,
		StartDate:    startDate,
		CreatedAt:    time.Now(),
	}
	for _, opt := range opts {
		opt(&t)
	}
	t.DisplayAmount = service.FormatCurrencyAmount(t.AmountCents, "AUD")
	return t
}

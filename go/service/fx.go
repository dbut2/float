package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/mastercard"
)

type FXService struct {
	q      database.Querier
	client *mastercard.FXClient
}

func NewFXService(q database.Querier, client *mastercard.FXClient) *FXService {
	return &FXService{q: q, client: client}
}

// GetRate returns how many units of quote currency equal 1 unit of base currency.
// e.g. GetRate(ctx, "AUD", "JPY") → 95.3 means 1 AUD = 95.3 JPY.
func (s *FXService) GetRate(ctx context.Context, base, quote string) (float64, error) {
	if s.client == nil {
		return 0, fmt.Errorf("FX service not configured")
	}

	today := time.Now().Truncate(24 * time.Hour)

	rate, err := s.q.GetFXRate(ctx, base, quote, today)
	if err == nil {
		return rate, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	rate, err = s.client.GetConversionRate(ctx, base, quote, today)
	if err != nil {
		log.Printf("FX: GetConversionRate(%s→%s): %v", base, quote, err)
		return 0, err
	}

	_ = s.q.UpsertFXRate(ctx, database.UpsertFXRateParams{
		BaseCurrency:  base,
		QuoteCurrency: quote,
		Rate:          rate,
		Date:          today,
	})
	return rate, nil
}

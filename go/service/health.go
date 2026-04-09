package service

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/utils"
)

// BucketHealth holds computed health signals for a single bucket.
type BucketHealth struct {
	BucketID         uuid.UUID  `json:"bucket_id"`
	BucketName       string     `json:"bucket_name"`
	Balance          float64    `json:"balance"`        // AUD, in dollars
	BalanceCents     int64      `json:"balance_cents"`  // raw cents
	TrickleAmount    float64    `json:"trickle_amount"` // trickle cycle amount in dollars (0 if none)
	TrickleAmountCents int64    `json:"trickle_amount_cents"`
	SpentPct         float64    `json:"spent_pct"`        // 0–1, fraction spent this cycle
	DailyAllowance   float64    `json:"daily_allowance"` // remaining / days until trickle
	DaysUntilTrickle int        `json:"days_until_trickle"`
	NextTrickleAt    *time.Time `json:"next_trickle_at,omitempty"`
	IsAtRisk         bool       `json:"is_at_risk"`
	Status           string     `json:"status"` // great/ok/warn/critical/stale
	HasTrickle       bool       `json:"has_trickle"`
	Period           string     `json:"period,omitempty"`
}

// HealthSummary is the top-level health payload returned by GET /api/health.
type HealthSummary struct {
	Buckets      []BucketHealth `json:"buckets"`
	OverallScore int            `json:"overall_score"` // 0–100
	AtRiskCount  int            `json:"at_risk_count"`
	StaleCount   int            `json:"stale_count"`
	HealthyCount int            `json:"healthy_count"`
}

// HealthService computes health signals and manages health notifications.
type HealthService struct {
	q    database.Querier
	push *PushService
}

func NewHealthService(q database.Querier, push *PushService) *HealthService {
	return &HealthService{q: q, push: push}
}

// GetHealthSummary computes health for all active non-general buckets for a user.
func (s *HealthService) GetHealthSummary(ctx context.Context, userID uuid.UUID) (HealthSummary, error) {
	// Load all active trickles for the user.
	trickleRows, err := s.q.ListTrickles(ctx, userID)
	if err != nil {
		return HealthSummary{}, err
	}

	// Build a map of trickle by ToBucketID for quick lookup.
	trickleByBucket := make(map[uuid.UUID]Trickle)
	for _, r := range trickleRows {
		t := dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt)
		trickleByBucket[t.ToBucketID] = t
	}

	// Load all buckets.
	bucketRows, err := s.q.ListBuckets(ctx, userID)
	if err != nil {
		return HealthSummary{}, err
	}

	now := utils.Now()
	today := utils.Today()

	// Pre-compute trickle amounts that affect balances (same logic as bucket service).
	bucketBalances := make(map[uuid.UUID]int64)
	for _, r := range bucketRows {
		bucketBalances[r.BucketID] = r.BalanceCents
	}
	for _, r := range trickleRows {
		t := dbTrickleToService(r.TrickleID, r.FromBucketID, r.FromBucketName, r.ToBucketID, r.ToBucketName, r.AmountCents, r.Description, r.Period, r.StartDate, r.EndDate, r.CreatedAt)
		amount := trickleAmount(t, now)
		bucketBalances[t.ToBucketID] += amount
		bucketBalances[t.FromBucketID] -= amount
	}

	var buckets []BucketHealth
	for _, r := range bucketRows {
		// Skip the general bucket — it's the funding source, not a spend bucket.
		if r.IsGeneral {
			continue
		}

		balanceCents := bucketBalances[r.BucketID]
		balanceDollars := float64(balanceCents) / 100.0

		bh := BucketHealth{
			BucketID:     r.BucketID,
			BucketName:   r.Name,
			BalanceCents: balanceCents,
			Balance:      balanceDollars,
		}

		t, hasTrickle := trickleByBucket[r.BucketID]
		bh.HasTrickle = hasTrickle

		if !hasTrickle {
			bh.Status = "stale"
			buckets = append(buckets, bh)
			continue
		}

		bh.Period = t.Period
		bh.TrickleAmountCents = t.AmountCents
		bh.TrickleAmount = float64(t.AmountCents) / 100.0

		// Find next trickle occurrence from today+1.
		tomorrow := today.AddDate(0, 0, 1)
		nextDate := nextOccurrence(t.StartDate, t.Period, tomorrow)
		bh.NextTrickleAt = &nextDate

		daysUntil := int(nextDate.Sub(today).Hours() / 24)
		if daysUntil < 0 {
			daysUntil = 0
		}
		bh.DaysUntilTrickle = daysUntil

		// SpentPct = (trickleAmount - balance) / trickleAmount, clamped 0-1.
		if t.AmountCents > 0 {
			spent := t.AmountCents - balanceCents
			pct := float64(spent) / float64(t.AmountCents)
			if pct < 0 {
				pct = 0
			}
			if pct > 1 {
				pct = 1
			}
			bh.SpentPct = pct
		}

		// DailyAllowance = balance / daysUntilTrickle (0 if no days left).
		if daysUntil > 0 {
			bh.DailyAllowance = balanceDollars / float64(daysUntil)
		}

		// IsAtRisk = balance is at or below zero before the next trickle.
		bh.IsAtRisk = balanceCents <= 0

		// Determine status.
		switch {
		case bh.IsAtRisk || bh.SpentPct > 0.8:
			bh.Status = "critical"
		case bh.SpentPct > 0.6:
			bh.Status = "warn"
		case bh.SpentPct > 0.4:
			bh.Status = "ok"
		default:
			bh.Status = "great"
		}

		buckets = append(buckets, bh)
	}

	// Compute overall score and counts.
	var scoreSum float64
	var scoreCount int
	atRisk, stale, healthy := 0, 0, 0

	for _, b := range buckets {
		if b.Status == "stale" {
			stale++
			continue
		}
		if b.IsAtRisk || b.Status == "critical" {
			atRisk++
		}
		if b.Status == "great" || b.Status == "ok" {
			healthy++
		}
		scoreSum += (1.0 - b.SpentPct) * 100.0
		scoreCount++
	}

	var overallScore int
	if scoreCount > 0 {
		overallScore = int(scoreSum / float64(scoreCount))
	}

	return HealthSummary{
		Buckets:      buckets,
		OverallScore: overallScore,
		AtRiskCount:  atRisk,
		StaleCount:   stale,
		HealthyCount: healthy,
	}, nil
}

// CheckAndNotify checks if a bucket is now critical/at-risk and, if it hasn't been
// notified today, sends a push notification and records the notification time.
// This is called asynchronously after classification.
func (s *HealthService) CheckAndNotify(ctx context.Context, userID, bucketID uuid.UUID) {
	if s.push == nil {
		return
	}

	// Recompute health for just this bucket.
	summary, err := s.GetHealthSummary(ctx, userID)
	if err != nil {
		log.Printf("health: CheckAndNotify: get summary: %v", err)
		return
	}

	var bh *BucketHealth
	for i := range summary.Buckets {
		if summary.Buckets[i].BucketID == bucketID {
			bh = &summary.Buckets[i]
			break
		}
	}
	if bh == nil || (bh.Status != "critical" && !bh.IsAtRisk) {
		return
	}

	// Check last notification time.
	lastNotified, err := s.q.GetBucketHealthNotification(ctx, bucketID, userID)
	if err == nil {
		// Already notified — check if it was today.
		today := utils.Today()
		notifiedDate := utils.ToDate(lastNotified)
		if !notifiedDate.Before(today) {
			// Already notified today — skip.
			return
		}
	} else if err != sql.ErrNoRows {
		log.Printf("health: CheckAndNotify: get notification: %v", err)
		return
	}

	// Send push notification.
	var body string
	if bh.IsAtRisk {
		body = "Balance is at $0 — no daily allowance available."
	} else {
		body = "Over 80% spent this trickle cycle."
	}
	s.push.SendNotification(ctx, userID, bh.BucketName+" is critical", body)

	// Upsert notification record.
	if err := s.q.UpsertBucketHealthNotification(ctx, bucketID, userID); err != nil {
		log.Printf("health: CheckAndNotify: upsert notification: %v", err)
	}
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

type Rule struct {
	RuleID              uuid.UUID `json:"rule_id"`
	BucketID            uuid.UUID `json:"bucket_id"`
	BucketName          string    `json:"bucket_name"`
	Name                string    `json:"name"`
	Priority            int32     `json:"priority"`
	DescriptionContains *string   `json:"description_contains"`
	MinAmountCents      *int64    `json:"min_amount_cents"`
	MaxAmountCents      *int64    `json:"max_amount_cents"`
	TransactionType     *string   `json:"transaction_type"`
	CategoryID          *string   `json:"category_id"`
	CreatedAt           time.Time `json:"created_at"`
}

type RuleService struct {
	q database.Querier
}

func NewRuleService(q database.Querier) *RuleService {
	return &RuleService{q: q}
}

func (s *RuleService) ListRules(ctx context.Context, userID uuid.UUID) ([]Rule, error) {
	rows, err := s.q.ListRulesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	rules := make([]Rule, len(rows))
	for i, r := range rows {
		rules[i] = dbRowToRule(r)
	}
	return rules, nil
}

func (s *RuleService) ListRulesByBucket(ctx context.Context, bucketID, userID uuid.UUID) ([]Rule, error) {
	rows, err := s.q.ListRulesByBucket(ctx, bucketID, userID)
	if err != nil {
		return nil, err
	}
	rules := make([]Rule, len(rows))
	for i, r := range rows {
		rules[i] = dbRowToRule(database.ListRulesForUserRow(r))
	}
	return rules, nil
}

func (s *RuleService) CreateRule(ctx context.Context, rule Rule) (Rule, error) {
	r, err := s.q.CreateRule(ctx, database.CreateRuleParams{
		BucketID:            rule.BucketID,
		Name:                rule.Name,
		Priority:            rule.Priority,
		DescriptionContains: nullableString(rule.DescriptionContains),
		MinAmountCents:      nullableInt64(rule.MinAmountCents),
		MaxAmountCents:      nullableInt64(rule.MaxAmountCents),
		TransactionType:     nullableString(rule.TransactionType),
		CategoryID:          nullableString(rule.CategoryID),
	})
	if err != nil {
		return Rule{}, err
	}
	rule.RuleID = r.RuleID
	rule.CreatedAt = r.CreatedAt
	return rule, nil
}

func (s *RuleService) UpdateRule(ctx context.Context, rule Rule, userID uuid.UUID) (Rule, error) {
	r, err := s.q.UpdateRule(ctx, database.UpdateRuleParams{
		RuleID:              rule.RuleID,
		Name:                rule.Name,
		Priority:            rule.Priority,
		DescriptionContains: nullableString(rule.DescriptionContains),
		MinAmountCents:      nullableInt64(rule.MinAmountCents),
		MaxAmountCents:      nullableInt64(rule.MaxAmountCents),
		TransactionType:     nullableString(rule.TransactionType),
		CategoryID:          nullableString(rule.CategoryID),
		UserID:              userID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Rule{}, ErrNotFound
	}
	if err != nil {
		return Rule{}, err
	}
	rule.RuleID = r.RuleID
	rule.BucketID = r.BucketID
	rule.CreatedAt = r.CreatedAt
	return rule, nil
}

func (s *RuleService) DeleteRule(ctx context.Context, ruleID, userID uuid.UUID) error {
	return s.q.DeleteRule(ctx, ruleID, userID)
}

func (s *RuleService) ApplyRulesToGeneral(ctx context.Context, userID uuid.UUID) (int, error) {
	general, err := s.q.GetGeneralBucket(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	txs, err := s.q.ListUpTransactionsByBucketID(ctx, general.BucketID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, tx := range txs {
		matched, err := applyRules(ctx, s.q, userID, tx.TransactionID, tx.Description, tx.AmountCents, tx.TransactionType, tx.CategoryID)
		if err != nil {
			log.Printf("applyRules for tx %s: %v", tx.TransactionID, err)
			continue
		}
		if matched {
			count++
		}
	}
	return count, nil
}

func applyRules(ctx context.Context, q database.Querier, userID, txID uuid.UUID, description string, amountCents int64, txType, categoryID sql.NullString) (bool, error) {
	rules, err := q.ListRulesForUser(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, r := range rules {
		if !matchesRule(r, description, amountCents, txType, categoryID) {
			continue
		}
		return true, q.AssignTransactionToBucket(ctx, txID, r.BucketID)
	}
	return false, nil
}

func matchesRule(r database.ListRulesForUserRow, description string, amountCents int64, txType, categoryID sql.NullString) bool {
	if r.DescriptionContains.Valid {
		if !strings.Contains(strings.ToLower(description), strings.ToLower(r.DescriptionContains.String)) {
			return false
		}
	}
	if r.MinAmountCents.Valid {
		if amountCents < r.MinAmountCents.Int64 {
			return false
		}
	}
	if r.MaxAmountCents.Valid {
		if amountCents > r.MaxAmountCents.Int64 {
			return false
		}
	}
	if r.TransactionType.Valid {
		if !txType.Valid || txType.String != r.TransactionType.String {
			return false
		}
	}
	if r.CategoryID.Valid {
		if !categoryID.Valid || categoryID.String != r.CategoryID.String {
			return false
		}
	}
	return true
}

func dbRowToRule(r database.ListRulesForUserRow) Rule {
	rule := Rule{
		RuleID:     r.RuleID,
		BucketID:   r.BucketID,
		BucketName: r.BucketName,
		Name:       r.Name,
		Priority:   r.Priority,
		CreatedAt:  r.CreatedAt,
	}
	if r.DescriptionContains.Valid {
		rule.DescriptionContains = &r.DescriptionContains.String
	}
	if r.MinAmountCents.Valid {
		rule.MinAmountCents = &r.MinAmountCents.Int64
	}
	if r.MaxAmountCents.Valid {
		rule.MaxAmountCents = &r.MaxAmountCents.Int64
	}
	if r.TransactionType.Valid {
		rule.TransactionType = &r.TransactionType.String
	}
	if r.CategoryID.Valid {
		rule.CategoryID = &r.CategoryID.String
	}
	return rule
}

func nullableString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullableInt64(n *int64) sql.NullInt64 {
	if n == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *n, Valid: true}
}

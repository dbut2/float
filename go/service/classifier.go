package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/utils"
)

type ReclassifyStatus struct {
	Running      bool `json:"running"`
	Total        int  `json:"total"`
	Processed    int  `json:"processed"`
	Reclassified int  `json:"reclassified"`
}

type ClassifierService struct {
	q   database.Querier
	llm LLMClient

	mu     sync.Mutex
	status map[uuid.UUID]*ReclassifyStatus // keyed by userID
}

func NewClassifierService(q database.Querier, llm LLMClient) *ClassifierService {
	return &ClassifierService{
		q:      q,
		llm:    llm,
		status: make(map[uuid.UUID]*ReclassifyStatus),
	}
}

func (s *ClassifierService) ClassifyTransaction(ctx context.Context, userID, txID uuid.UUID, description string, amountCents int64, txType, categoryID sql.NullString, createdAt time.Time, foreignCurrencyCode sql.NullString) error {
	buckets, err := s.q.ListBuckets(ctx, userID)
	if err != nil {
		return fmt.Errorf("list buckets: %w", err)
	}

	if len(buckets) == 0 {
		return nil
	}

	var generalBucketID uuid.UUID
	var nonGeneralBuckets []database.ListBucketsRow
	for _, b := range buckets {
		if b.IsGeneral {
			generalBucketID = b.BucketID
		} else {
			nonGeneralBuckets = append(nonGeneralBuckets, b)
		}
	}

	if len(nonGeneralBuckets) == 0 {
		return nil
	}

	prompt, err := s.buildPrompt(ctx, nonGeneralBuckets, description, amountCents, txType, categoryID, createdAt, foreignCurrencyCode)
	if err != nil {
		return fmt.Errorf("build prompt: %w", err)
	}

	result, err := s.llm.Classify(ctx, prompt)
	if err != nil {
		log.Printf("LLM classification failed for tx %s, leaving in General: %v", txID, err)
		return nil
	}

	bucketNameToID := make(map[string]uuid.UUID)
	for _, b := range nonGeneralBuckets {
		bucketNameToID[b.Name] = b.BucketID
	}

	chosenBucketID := generalBucketID
	if result.Confidence >= 0.6 && result.BucketName != "General" {
		if id, ok := bucketNameToID[result.BucketName]; ok {
			chosenBucketID = id
		}
	}

	if chosenBucketID != generalBucketID {
		if err := s.q.AssignTransactionToBucket(ctx, txID, chosenBucketID); err != nil {
			return fmt.Errorf("assign transaction: %w", err)
		}
	}

	_ = s.q.InsertClassificationLog(ctx, database.InsertClassificationLogParams{
		TransactionID:  txID,
		ChosenBucketID: chosenBucketID,
		Confidence:     float32(result.Confidence),
		Reasoning:      result.Reasoning,
		Model:          s.llm.Model(),
	})

	return nil
}

// ClassifyOne fetches a single transaction by ID and classifies it.
func (s *ClassifierService) ClassifyOne(ctx context.Context, userID, txID uuid.UUID) error {
	tx, err := s.q.GetTransaction(ctx, txID, userID)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}
	return s.ClassifyTransaction(ctx, userID, tx.TransactionID, tx.Description, tx.AmountCents, sql.NullString{}, sql.NullString{}, tx.CreatedAt, tx.ForeignCurrencyCode)
}

// StartReclassifyGeneral kicks off background reclassification and returns immediately.
// Returns false if a reclassification is already running for this user.
func (s *ClassifierService) StartReclassifyGeneral(userID uuid.UUID) bool {
	s.mu.Lock()
	if st, ok := s.status[userID]; ok && st.Running {
		s.mu.Unlock()
		return false
	}
	status := &ReclassifyStatus{Running: true}
	s.status[userID] = status
	s.mu.Unlock()

	go s.runReclassify(userID, status)
	return true
}

// GetReclassifyStatus returns the current reclassification status for a user.
func (s *ClassifierService) GetReclassifyStatus(userID uuid.UUID) ReclassifyStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.status[userID]; ok {
		return *st
	}
	return ReclassifyStatus{}
}

func (s *ClassifierService) runReclassify(userID uuid.UUID, status *ReclassifyStatus) {
	defer func() {
		s.mu.Lock()
		status.Running = false
		s.mu.Unlock()
	}()

	ctx := context.Background()

	general, err := s.q.GetGeneralBucket(ctx, userID)
	if err != nil {
		log.Printf("reclassify: get general bucket: %v", err)
		return
	}

	txs, err := s.q.ListBucketTransactions(ctx, general.BucketID)
	if err != nil {
		log.Printf("reclassify: list transactions: %v", err)
		return
	}

	// Filter to actual transactions only
	var toClassify []database.FloatBucketLedger
	for _, tx := range txs {
		if tx.IsTransaction {
			toClassify = append(toClassify, tx)
		}
	}

	s.mu.Lock()
	status.Total = len(toClassify)
	s.mu.Unlock()

	for _, tx := range toClassify {
		beforeBucket := general.BucketID

		if err := s.ClassifyTransaction(ctx, userID, tx.TransactionID, tx.Description, tx.AmountCents, sql.NullString{}, sql.NullString{}, tx.CreatedAt, tx.ForeignCurrencyCode); err != nil {
			log.Printf("classify tx %s: %v", tx.TransactionID, err)
		} else {
			ledgerEntry, err := s.q.GetTransaction(ctx, tx.TransactionID, userID)
			if err == nil && ledgerEntry.BucketID != beforeBucket {
				s.mu.Lock()
				status.Reclassified++
				s.mu.Unlock()
			}
		}

		s.mu.Lock()
		status.Processed++
		s.mu.Unlock()
	}
}

func (s *ClassifierService) buildPrompt(ctx context.Context, buckets []database.ListBucketsRow, description string, amountCents int64, txType, categoryID sql.NullString, createdAt time.Time, foreignCurrencyCode sql.NullString) (string, error) {
	var sb strings.Builder
	sb.WriteString("You are a transaction categorizer for a personal budget app.\n\n")
	sb.WriteString("The user has these budget buckets:\n")

	for _, b := range buckets {
		_, _ = fmt.Fprintf(&sb, "- %q", b.Name)
		if b.Description != "" {
			sb.WriteString(": " + b.Description)
		}
		sb.WriteString("\n")

		samples, err := s.q.ListBucketSampleTransactions(ctx, b.BucketID)
		if err == nil && len(samples) > 0 {
			sb.WriteString("  Example transactions: ")
			parts := make([]string, 0, len(samples))
			for _, s := range samples {
				display := utils.FormatSignedAmount(s.AmountCents, "AUD")
				extra := ""
				if s.CategoryID.Valid {
					extra = ", " + s.CategoryID.String
				}
				if s.ForeignCurrencyCode.Valid {
					extra = ", " + s.ForeignCurrencyCode.String
				}
				parts = append(parts, fmt.Sprintf("%s (%s%s)", s.Description, display, extra))
			}
			sb.WriteString(strings.Join(parts, "; "))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nClassify this transaction:\n")
	sb.WriteString("- Description: " + description + "\n")
	sb.WriteString("- Amount: " + utils.FormatSignedAmount(amountCents, "AUD") + "\n")
	if categoryID.Valid {
		sb.WriteString("- Category: " + categoryID.String + "\n")
	}
	if txType.Valid {
		sb.WriteString("- Type: " + txType.String + "\n")
	}
	currency := "AUD"
	if foreignCurrencyCode.Valid {
		currency = foreignCurrencyCode.String
	}
	sb.WriteString("- Currency: " + currency + "\n")
	sb.WriteString("- Date: " + createdAt.Format("2006-01-02") + "\n")

	sb.WriteString("\nWhich bucket does this belong to? If none are a good fit, respond with \"General\".\n")
	sb.WriteString("Respond with JSON: {\"bucket_name\": \"...\", \"confidence\": 0.0-1.0, \"reasoning\": \"...\"}\n")

	return sb.String(), nil
}

package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

//go:embed data/demo.json
var demoDataJSON []byte

type demoData struct {
	User         User              `json:"user"`
	Buckets      []demoBucketData  `json:"buckets"`
	Transactions []demoTxData      `json:"transactions"`
	Transfers    []demoTranserData `json:"transfers"`
}

type demoBucketData struct {
	BucketID  uuid.UUID `json:"bucket_id"`
	Name      string    `json:"name"`
	IsGeneral bool      `json:"is_general"`
	CreatedAt time.Time `json:"created_at"`
}

type demoTxData struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	BucketID      uuid.UUID `json:"bucket_id"`
	Description   string    `json:"description"`
	Message       string    `json:"message"`
	AmountCents   int64     `json:"amount_cents"`
	DisplayAmount string    `json:"display_amount"`
	CurrencyCode  string    `json:"currency_code"`
	CreatedAt     time.Time `json:"created_at"`
}

type demoTranserData struct {
	TransferID   uuid.UUID `json:"transfer_id"`
	FromBucketID uuid.UUID `json:"from_bucket_id"`
	ToBucketID   uuid.UUID `json:"to_bucket_id"`
	AmountCents  int64     `json:"amount_cents"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
}

type DemoService struct {
	user      User
	buckets   []Bucket
	transfers []Transfer
	ledger    []Transaction
}

func NewDemoService() *DemoService {
	var data demoData
	if err := json.Unmarshal(demoDataJSON, &data); err != nil {
		panic("failed to parse demo data: " + err.Error())
	}

	s := &DemoService{
		user: data.User,
	}

	bucketNames := make(map[uuid.UUID]string)

	for _, b := range data.Buckets {
		bucketNames[b.BucketID] = b.Name
		s.buckets = append(s.buckets, Bucket{
			BucketID:  b.BucketID,
			UserID:    data.User.UserID,
			Name:      b.Name,
			IsGeneral: b.IsGeneral,
			CreatedAt: b.CreatedAt,
		})
	}

	for _, tx := range data.Transactions {
		s.ledger = append(s.ledger, Transaction{
			TransactionID: tx.TransactionID,
			BucketID:      tx.BucketID,
			UserID:        data.User.UserID,
			Description:   tx.Description,
			Message:       tx.Message,
			AmountCents:   tx.AmountCents,
			DisplayAmount: tx.DisplayAmount,
			CurrencyCode:  tx.CurrencyCode,
			CreatedAt:     tx.CreatedAt,
			IsTransaction: true,
		})
	}

	for _, t := range data.Transfers {
		s.transfers = append(s.transfers, Transfer{
			TransferID:     t.TransferID,
			FromBucketID:   t.FromBucketID,
			FromBucketName: bucketNames[t.FromBucketID],
			ToBucketID:     t.ToBucketID,
			ToBucketName:   bucketNames[t.ToBucketID],
			AmountCents:    t.AmountCents,
			Note:           t.Note,
			CreatedAt:      t.CreatedAt,
		})
		s.ledger = append(s.ledger, Transaction{
			TransactionID: t.TransferID,
			BucketID:      t.FromBucketID,
			UserID:        data.User.UserID,
			AmountCents:   -t.AmountCents,
			CreatedAt:     t.CreatedAt,
			IsTransaction: false,
		})
		s.ledger = append(s.ledger, Transaction{
			TransactionID: t.TransferID,
			BucketID:      t.ToBucketID,
			UserID:        data.User.UserID,
			AmountCents:   t.AmountCents,
			CreatedAt:     t.CreatedAt,
			IsTransaction: false,
		})
	}

	balances := make(map[uuid.UUID]int64)
	for _, entry := range s.ledger {
		balances[entry.BucketID] += entry.AmountCents
	}
	for i := range s.buckets {
		s.buckets[i].BalanceCents = balances[s.buckets[i].BucketID]
	}

	return s
}

func (s *DemoService) UserID() uuid.UUID {
	return s.user.UserID
}

func (s *DemoService) GetUser(_ context.Context, _ uuid.UUID) (User, error) {
	return s.user, nil
}

func (s *DemoService) UpdateToken(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (s *DemoService) SyncTransactions(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (s *DemoService) ListBuckets(_ context.Context, _ uuid.UUID) ([]Bucket, error) {
	return s.buckets, nil
}

func (s *DemoService) CreateBucket(_ context.Context, bucket Bucket) (Bucket, error) {
	bucket.BucketID = uuid.New()
	bucket.CreatedAt = time.Now()
	return bucket, nil
}

func (s *DemoService) GetBucket(_ context.Context, bucketID, _ uuid.UUID) (Bucket, error) {
	for _, b := range s.buckets {
		if b.BucketID == bucketID {
			return b, nil
		}
	}
	return Bucket{}, ErrNotFound
}

func (s *DemoService) DeleteBucket(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) ListBucketTransactions(_ context.Context, bucketID uuid.UUID) ([]Transaction, error) {
	var txs []Transaction
	for _, t := range s.ledger {
		if t.BucketID == bucketID {
			txs = append(txs, t)
		}
	}
	return txs, nil
}

func (s *DemoService) ListTransactions(_ context.Context, _ uuid.UUID) ([]Transaction, error) {
	return s.ledger, nil
}

func (s *DemoService) AssignToBucket(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) ListTransfers(_ context.Context, _ uuid.UUID) ([]Transfer, error) {
	return s.transfers, nil
}

func (s *DemoService) CreateTransfer(_ context.Context, transfer Transfer) (Transfer, error) {
	transfer.TransferID = uuid.New()
	transfer.CreatedAt = time.Now()
	return transfer, nil
}

func (s *DemoService) DeleteTransfer(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) RegisterToken(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (s *DemoService) UnregisterToken(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (s *DemoService) ListTrickles(_ context.Context, _ uuid.UUID) ([]Trickle, error) {
	return []Trickle{}, nil
}

func (s *DemoService) GetTrickle(_ context.Context, _, _ uuid.UUID) (Trickle, error) {
	return Trickle{}, ErrNotFound
}

func (s *DemoService) UpsertTrickle(_ context.Context, trickle Trickle) (Trickle, error) {
	trickle.TrickleID = uuid.New()
	trickle.CreatedAt = time.Now()
	return trickle, nil
}

func (s *DemoService) DeleteTrickle(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) ListRules(_ context.Context, _ uuid.UUID) ([]Rule, error) {
	return []Rule{}, nil
}

func (s *DemoService) ListRulesByBucket(_ context.Context, _, _ uuid.UUID) ([]Rule, error) {
	return []Rule{}, nil
}

func (s *DemoService) CreateRule(_ context.Context, rule Rule) (Rule, error) {
	rule.RuleID = uuid.New()
	rule.CreatedAt = time.Now()
	return rule, nil
}

func (s *DemoService) UpdateRule(_ context.Context, rule Rule, _ uuid.UUID) (Rule, error) {
	return rule, nil
}

func (s *DemoService) DeleteRule(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) ApplyRulesToGeneral(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

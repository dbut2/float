package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/utils"
)

//go:embed data/demo.json
var demoDataJSON []byte

type demoData struct {
	User         User              `json:"user"`
	Buckets      []demoBucketData  `json:"buckets"`
	Transactions []demoTxData      `json:"transactions"`
	Transfers    []demoTranserData `json:"transfers"`
	Trickles     []demoTrickleData `json:"trickles"`
}

type demoBucketData struct {
	BucketID     uuid.UUID `json:"bucket_id"`
	Name         string    `json:"name"`
	IsGeneral    bool      `json:"is_general"`
	CreatedAt    time.Time `json:"created_at"`
	CurrencyCode *string   `json:"currency_code"`
}

type demoTxData struct {
	TransactionID       uuid.UUID `json:"transaction_id"`
	BucketID            uuid.UUID `json:"bucket_id"`
	Description         string    `json:"description"`
	Message             string    `json:"message"`
	AmountCents         int64     `json:"amount_cents"`
	CreatedAt           time.Time `json:"created_at"`
	ForeignCurrencyCode *string   `json:"foreign_currency_code"`
	ForeignAmountCents  *int64    `json:"foreign_amount_cents"`
}

type demoTranserData struct {
	TransferID   uuid.UUID `json:"transfer_id"`
	FromBucketID uuid.UUID `json:"from_bucket_id"`
	ToBucketID   uuid.UUID `json:"to_bucket_id"`
	AmountCents  int64     `json:"amount_cents"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
}

type demoTrickleData struct {
	TrickleID   uuid.UUID  `json:"trickle_id"`
	ToBucketID  uuid.UUID  `json:"to_bucket_id"`
	AmountCents int64      `json:"amount_cents"`
	Description string     `json:"description"`
	Period      string     `json:"period"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   time.Time  `json:"created_at"`
}

type DemoService struct {
	user      User
	buckets   []Bucket
	transfers []Transfer
	ledger    []Transaction
	trickles  []Trickle
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

	// Demo FX rates (static, for display only)
	demoRates := map[string]float64{
		"CNY": 4.72,
		"JPY": 95.3,
		"USD": 0.63,
		"EUR": 0.58,
	}

	for _, b := range data.Buckets {
		bucketNames[b.BucketID] = b.Name
		bucket := Bucket{
			BucketID:     b.BucketID,
			UserID:       data.User.UserID,
			Name:         b.Name,
			IsGeneral:    b.IsGeneral,
			CreatedAt:    b.CreatedAt,
			CurrencyCode: b.CurrencyCode,
		}
		if b.CurrencyCode != nil {
			if rate, ok := demoRates[*b.CurrencyCode]; ok {
				bucket.FXRate = &rate
			}
		}
		s.buckets = append(s.buckets, bucket)
	}

	for _, tx := range data.Transactions {
		t := Transaction{
			TransactionID: tx.TransactionID,
			BucketID:      tx.BucketID,
			UserID:        data.User.UserID,
			Description:   tx.Description,
			Message:       tx.Message,
			AmountCents:   tx.AmountCents,
			DisplayAmount: FormatSignedAmount(tx.AmountCents, "AUD"),
			CreatedAt:     tx.CreatedAt,
			IsTransaction: true,
		}
		t.ForeignCurrencyCode = tx.ForeignCurrencyCode
		t.ForeignAmountCents = tx.ForeignAmountCents
		if tx.ForeignCurrencyCode != nil && tx.ForeignAmountCents != nil {
			display := FormatSignedAmount(*tx.ForeignAmountCents, *tx.ForeignCurrencyCode)
			t.ForeignDisplayAmount = &display
		}
		s.ledger = append(s.ledger, t)
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
		if s.buckets[i].CurrencyCode != nil && s.buckets[i].FXRate != nil {
			display := utils.FormatForeignBalance(s.buckets[i].BalanceCents, *s.buckets[i].FXRate, *s.buckets[i].CurrencyCode)
			s.buckets[i].ForeignBalanceDisplay = &display
		}
		s.buckets[i].setDisplays()
	}

	generalBucketID := uuid.UUID{}
	for _, b := range s.buckets {
		if b.IsGeneral {
			generalBucketID = b.BucketID
			break
		}
	}

	for _, t := range data.Trickles {
		s.trickles = append(s.trickles, Trickle{
			TrickleID:      t.TrickleID,
			FromBucketID:   generalBucketID,
			FromBucketName: bucketNames[generalBucketID],
			ToBucketID:     t.ToBucketID,
			ToBucketName:   bucketNames[t.ToBucketID],
			AmountCents:    t.AmountCents,
			Description:    t.Description,
			Period:         t.Period,
			StartDate:      t.StartDate,
			EndDate:        t.EndDate,
			CreatedAt:      t.CreatedAt,
			UserID:         data.User.UserID,
		})
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

func (s *DemoService) GetTransactBalance(_ context.Context, _ uuid.UUID) (int64, error) {
	return 189234, nil
}

func (s *DemoService) ListBuckets(_ context.Context, _ uuid.UUID) ([]Bucket, error) {
	return s.buckets, nil
}

func (s *DemoService) CreateBucket(_ context.Context, bucket Bucket) (Bucket, error) {
	bucket.BucketID = uuid.New()
	bucket.CreatedAt = utils.Now()
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

func (s *DemoService) CloseBucket(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) ListBucketTransactions(_ context.Context, bucketID, _ uuid.UUID) ([]Transaction, error) {
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

func (s *DemoService) AssignToBucket(_ context.Context, _, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) ListTransfers(_ context.Context, _ uuid.UUID) ([]Transfer, error) {
	return s.transfers, nil
}

func (s *DemoService) CreateTransfer(_ context.Context, transfer Transfer) (Transfer, error) {
	transfer.TransferID = uuid.New()
	transfer.CreatedAt = utils.Now()
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
	return s.trickles, nil
}

func (s *DemoService) GetTrickle(_ context.Context, toBucketID, _ uuid.UUID) (Trickle, error) {
	for _, t := range s.trickles {
		if t.ToBucketID == toBucketID {
			return t, nil
		}
	}
	return Trickle{}, ErrNotFound
}

func (s *DemoService) UpsertTrickle(_ context.Context, trickle Trickle) (Trickle, error) {
	trickle.TrickleID = uuid.New()
	trickle.CreatedAt = utils.Now()
	return trickle, nil
}

func (s *DemoService) DeleteTrickle(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) ReorderBuckets(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return nil
}

func (s *DemoService) UpdateBucketDescription(_ context.Context, _, _ uuid.UUID, _ string) error {
	return nil
}

func (s *DemoService) ClassifyOne(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (s *DemoService) StartReclassifyGeneral(_ uuid.UUID) bool {
	return true
}

func (s *DemoService) GetReclassifyStatus(_ uuid.UUID) ReclassifyStatus {
	return ReclassifyStatus{}
}

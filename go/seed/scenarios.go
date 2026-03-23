package seed

import (
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/service"
)

func DemoScenario() (service.User, []service.Bucket, []service.Transaction, []service.Transfer, []service.Trickle) {
	user := CreateUser(
		WithUserID(uuid.MustParse("00000000-0000-4000-8000-000000000001")),
		WithEmail("demo@float-demo.dbut.dev"),
	)

	general := CreateGeneralBucket(user.UserID,
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000010")),
	)
	rent := CreateBucket(user.UserID, "Rent",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000011")),
	)
	groceries := CreateBucket(user.UserID, "Groceries",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000012")),
	)
	savings := CreateBucket(user.UserID, "Savings",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000013")),
	)
	japanTrip := CreateBucket(user.UserID, "Japan Trip",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000014")),
		WithCurrency("JPY"),
	)

	buckets := []service.Bucket{general, rent, groceries, savings, japanTrip}

	var txs []service.Transaction

	txs = append(txs, CreateDeposit(general.BucketID, 520000,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000100")),
		WithDescription("Salary"), WithMessage("February pay"),
		At("2026-02-28T09:00:00Z"),
	))
	txs = append(txs, CreateDeposit(general.BucketID, 520000,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000101")),
		WithDescription("Salary"), WithMessage("January pay"),
		At("2026-01-31T09:00:00Z"),
	))

	txs = append(txs, CreateExpense(rent.BucketID, 180000,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000102")),
		WithDescription("Rent Payment"),
		At("2026-02-01T10:00:00Z"),
	))
	txs = append(txs, CreateExpense(rent.BucketID, 180000,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000103")),
		WithDescription("Rent Payment"),
		At("2026-03-01T10:00:00Z"),
	))

	txs = append(txs, CreateExpense(groceries.BucketID, 8543,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000104")),
		WithDescription("Woolworths"), WithMessage("Weekly shop"),
		At("2026-02-10T14:30:00Z"),
	))
	txs = append(txs, CreateExpense(groceries.BucketID, 6290,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000105")),
		WithDescription("Coles"),
		At("2026-02-17T16:00:00Z"),
	))
	txs = append(txs, CreateExpense(groceries.BucketID, 9120,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000106")),
		WithDescription("Woolworths"), WithMessage("Weekly shop"),
		At("2026-02-24T13:15:00Z"),
	))

	txs = append(txs, CreateExpense(general.BucketID, 1699,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000107")),
		WithDescription("Netflix"),
		At("2026-02-05T00:00:00Z"),
	))
	txs = append(txs, CreateExpense(general.BucketID, 1299,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000108")),
		WithDescription("Spotify"),
		At("2026-02-05T00:00:00Z"),
	))
	txs = append(txs, CreateExpense(general.BucketID, 24500,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000109")),
		WithDescription("Electricity Bill"), WithMessage("Quarterly bill"),
		At("2026-02-15T09:00:00Z"),
	))

	txs = append(txs, CreateExpense(japanTrip.BucketID, 148900,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000110")),
		WithDescription("Japan Airlines"), WithMessage("Return flights MEL-TYO"),
		At("2026-02-20T11:00:00Z"),
	))
	txs = append(txs, CreateExpense(japanTrip.BucketID, 52341,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000111")),
		WithDescription("Keio Plaza Hotel"), WithMessage("Shinjuku, 5 nights"),
		WithForeign("JPY", -4987300),
		At("2026-02-18T15:30:00Z"),
	))

	txs = append(txs, CreateExpense(general.BucketID, 550,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000112")),
		WithDescription("Coffee"),
		At("2026-02-26T07:30:00Z"),
	))
	txs = append(txs, CreateDeposit(general.BucketID, 1250,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000113")),
		WithDescription("Interest"), WithMessage("Monthly interest"),
		At("2026-02-28T23:59:00Z"),
	))

	txs = append(txs, CreateExpense(groceries.BucketID, 4380,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000114")),
		WithDescription("Aldi"),
		At("2026-03-01T11:00:00Z"),
	))

	transfers := []service.Transfer{
		CreateTransfer(general.BucketID, rent.BucketID, 360000, WithNote("Rent for Jan + Feb")),
		CreateTransfer(general.BucketID, groceries.BucketID, 40000, WithNote("Grocery budget")),
		CreateTransfer(general.BucketID, savings.BucketID, 100000, WithNote("Monthly savings")),
		CreateTransfer(general.BucketID, japanTrip.BucketID, 75000, WithNote("Holiday fund")),
	}

	trickles := []service.Trickle{
		CreateTrickle(general.BucketID, rent.BucketID, 180000, "monthly",
			time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			WithTrickleDescription("Monthly rent"),
		),
		CreateTrickle(general.BucketID, savings.BucketID, 50000, "monthly",
			time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			WithTrickleDescription("Monthly savings"),
		),
	}

	return user, buckets, txs, transfers, trickles
}

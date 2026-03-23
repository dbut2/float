package seed

import (
	"time"

	"dbut.dev/float/go/service"
)

func DemoScenario() (service.User, []service.Bucket, []service.Transaction, []service.Transfer, []service.Trickle) {
	user := CreateUser(WithEmail("demo@float-demo.dbut.dev"))

	general := CreateGeneralBucket(user.UserID)
	rent := CreateBucket(user.UserID, "Rent")
	groceries := CreateBucket(user.UserID, "Groceries")
	savings := CreateBucket(user.UserID, "Savings")
	japanTrip := CreateBucket(user.UserID, "Japan Trip", WithCurrency("JPY"))

	buckets := []service.Bucket{general, rent, groceries, savings, japanTrip}

	var txs []service.Transaction

	txs = append(txs, CreateDeposit(general.BucketID, 520000,
		WithDescription("Salary"), WithMessage("January pay"),
		At("2026-01-31T09:00:00Z"),
	))
	txs = append(txs, CreateDeposit(general.BucketID, 520000,
		WithDescription("Salary"), WithMessage("February pay"),
		At("2026-02-28T09:00:00Z"),
	))

	txs = append(txs, CreateExpense(rent.BucketID, 180000,
		WithDescription("Rent Payment"),
		At("2026-02-01T10:00:00Z"),
	))
	txs = append(txs, CreateExpense(rent.BucketID, 180000,
		WithDescription("Rent Payment"),
		At("2026-03-01T10:00:00Z"),
	))

	txs = append(txs, CreateExpense(groceries.BucketID, 8543,
		WithDescription("Woolworths"), WithMessage("Weekly shop"),
		At("2026-02-10T14:30:00Z"),
	))
	txs = append(txs, CreateExpense(groceries.BucketID, 6290,
		WithDescription("Coles"),
		At("2026-02-17T16:00:00Z"),
	))
	txs = append(txs, CreateExpense(groceries.BucketID, 9120,
		WithDescription("Woolworths"), WithMessage("Weekly shop"),
		At("2026-02-24T13:15:00Z"),
	))
	txs = append(txs, CreateExpense(groceries.BucketID, 4380,
		WithDescription("Aldi"),
		At("2026-03-01T11:00:00Z"),
	))

	txs = append(txs, CreateExpense(general.BucketID, 1699,
		WithDescription("Netflix"),
		At("2026-02-05T00:00:00Z"),
	))
	txs = append(txs, CreateExpense(general.BucketID, 1299,
		WithDescription("Spotify"),
		At("2026-02-05T00:00:00Z"),
	))
	txs = append(txs, CreateExpense(general.BucketID, 24500,
		WithDescription("Electricity Bill"), WithMessage("Quarterly bill"),
		At("2026-02-15T09:00:00Z"),
	))
	txs = append(txs, CreateExpense(general.BucketID, 550,
		WithDescription("Coffee"),
		At("2026-02-26T07:30:00Z"),
	))
	txs = append(txs, CreateDeposit(general.BucketID, 1250,
		WithDescription("Interest"), WithMessage("Monthly interest"),
		At("2026-02-28T23:59:00Z"),
	))

	txs = append(txs, CreateExpense(japanTrip.BucketID, 0,
		WithDescription("Japan Airlines"), WithMessage("Return flights MEL-TYO"),
		WithForeign("JPY", -1419470),
		At("2026-02-20T11:00:00Z"),
	))
	txs = append(txs, CreateExpense(japanTrip.BucketID, 0,
		WithDescription("Keio Plaza Hotel"), WithMessage("Shinjuku, 5 nights"),
		WithForeign("JPY", -4987300),
		At("2026-02-18T15:30:00Z"),
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

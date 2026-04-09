package seed

import (
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/service"
)

// today is 9 April 2026 (Wednesday). Weekly trickles run Monday→Monday.
// This means the last trickle was Mon 7 Apr, next is Mon 14 Apr (5 days away).
// Monthly trickles started 1 Apr, next 1 May (22 days away).

func DemoScenario() (service.User, []service.Bucket, []service.Transaction, []service.Transfer, []service.Trickle) {
	user := CreateUser(
		WithUserID(uuid.MustParse("00000000-0000-4000-8000-000000000001")),
		WithEmail("demo@float-demo.dbut.dev"),
	)

	general := CreateGeneralBucket(user.UserID,
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000010")),
	)
	groceries := CreateBucket(user.UserID, "Groceries",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000011")),
	)
	eatingOut := CreateBucket(user.UserID, "Eating Out",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000012")),
	)
	transport := CreateBucket(user.UserID, "Transport",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000013")),
	)
	japanTrip := CreateBucket(user.UserID, "Japan Trip",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000014")),
		WithCurrency("JPY"),
	)
	savings := CreateBucket(user.UserID, "Savings",
		WithBucketID(uuid.MustParse("00000000-0000-4000-8000-000000000015")),
	)

	buckets := []service.Bucket{general, groceries, eatingOut, transport, japanTrip, savings}

	var txs []service.Transaction

	// --- Salary (General) ---
	txs = append(txs, CreateDeposit(general.BucketID, 520000,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000100")),
		WithDescription("Salary"), WithMessage("April pay"),
		At("2026-04-01T09:00:00Z"),
	))
	txs = append(txs, CreateDeposit(general.BucketID, 520000,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000101")),
		WithDescription("Salary"), WithMessage("March pay"),
		At("2026-03-31T09:00:00Z"),
	))

	// --- GROCERIES: CRITICAL — $190 of $200 weekly budget spent on day 2 ---
	// Weekly trickle started Mon 7 Apr. $190 already spent. Only $10 left. 5 days until refill.
	txs = append(txs, CreateExpense(groceries.BucketID, 10320,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000110")),
		WithDescription("Woolworths"), WithMessage("Weekly shop"),
		At("2026-04-07T14:00:00Z"), // Monday
	))
	txs = append(txs, CreateExpense(groceries.BucketID, 8760,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000111")),
		WithDescription("Coles"),
		At("2026-04-08T11:30:00Z"), // Tuesday
	))
	// Previous week (healthy back-reference)
	txs = append(txs, CreateExpense(groceries.BucketID, 9540,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000112")),
		WithDescription("Woolworths"), WithMessage("Weekly shop"),
		At("2026-03-31T14:00:00Z"),
	))
	txs = append(txs, CreateExpense(groceries.BucketID, 6870,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000113")),
		WithDescription("Aldi"),
		At("2026-04-03T10:00:00Z"),
	))

	// --- EATING OUT: WARNING — $140 of $200 fortnightly budget spent (7 days in, 7 days left) ---
	txs = append(txs, CreateExpense(eatingOut.BucketID, 5200,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000120")),
		WithDescription("Neighbourhood Wine"),
		At("2026-04-05T20:30:00Z"),
	))
	txs = append(txs, CreateExpense(eatingOut.BucketID, 3850,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000121")),
		WithDescription("Penny for Pound"), WithMessage("Lunch with Kim"),
		At("2026-04-07T13:00:00Z"),
	))
	txs = append(txs, CreateExpense(eatingOut.BucketID, 2490,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000122")),
		WithDescription("Hector's Deli"),
		At("2026-04-08T12:30:00Z"),
	))
	txs = append(txs, CreateExpense(eatingOut.BucketID, 4670,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000123")),
		WithDescription("Tipo 00"), WithMessage("Date night"),
		At("2026-04-09T19:00:00Z"),
	))

	// --- TRANSPORT: OK — $81 of $180 monthly budget spent (9 days in, 22 days left) ---
	txs = append(txs, CreateExpense(transport.BucketID, 1950,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000130")),
		WithDescription("Myki Top-Up"),
		At("2026-04-01T08:00:00Z"),
	))
	txs = append(txs, CreateExpense(transport.BucketID, 6500,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000131")),
		WithDescription("Uber"), WithMessage("Late night home"),
		At("2026-04-05T23:45:00Z"),
	))
	txs = append(txs, CreateExpense(transport.BucketID, 650,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000132")),
		WithDescription("Myki Top-Up"),
		At("2026-04-09T07:30:00Z"),
	))

	// --- JAPAN TRIP: GREAT — only $180 of $1000 monthly budget spent ---
	txs = append(txs, CreateExpense(japanTrip.BucketID, 9900,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000140")),
		WithDescription("Japan Airlines"), WithMessage("Change fee"),
		WithForeign("JPY", -990000),
		At("2026-04-03T10:00:00Z"),
	))
	txs = append(txs, CreateExpense(japanTrip.BucketID, 8100,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000141")),
		WithDescription("Hyperdia Pass"), WithMessage("7-day rail pass"),
		WithForeign("JPY", -810000),
		At("2026-04-07T09:00:00Z"),
	))

	// --- SAVINGS: no trickle, stale — hasn't had a transaction in 6 weeks ---
	txs = append(txs, CreateDeposit(savings.BucketID, 100000,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000150")),
		WithDescription("Transfer to Savings"), WithMessage("March savings"),
		At("2026-02-28T10:00:00Z"),
	))

	// --- General: subscriptions & bills ---
	txs = append(txs, CreateExpense(general.BucketID, 1699,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000160")),
		WithDescription("Netflix"),
		At("2026-04-05T00:00:00Z"),
	))
	txs = append(txs, CreateExpense(general.BucketID, 1299,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000161")),
		WithDescription("Spotify"),
		At("2026-04-05T00:00:00Z"),
	))
	txs = append(txs, CreateExpense(general.BucketID, 24500,
		WithTransactionID(uuid.MustParse("00000000-0000-4000-8000-000000000162")),
		WithDescription("Electricity Bill"),
		At("2026-04-02T09:00:00Z"),
	))

	transfers := []service.Transfer{
		CreateTransfer(general.BucketID, groceries.BucketID, 20000, WithNote("Grocery budget top-up")),
		CreateTransfer(general.BucketID, eatingOut.BucketID, 20000, WithNote("Eating out budget")),
		CreateTransfer(general.BucketID, transport.BucketID, 10000, WithNote("Transport top-up")),
		CreateTransfer(general.BucketID, japanTrip.BucketID, 100000, WithNote("Japan trip fund")),
		CreateTransfer(general.BucketID, savings.BucketID, 50000, WithNote("Old manual savings transfer")),
	}

	// Trickles: designed to produce varied health states as of 9 April 2026
	trickles := []service.Trickle{
		// Groceries: $200/week, started Mon 7 Apr → next Mon 14 Apr (5 days away)
		// After spending $190, only $10 left → CRITICAL
		CreateTrickle(general.BucketID, groceries.BucketID, 20000, "weekly",
			time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC),
			WithTrickleDescription("Weekly grocery budget"),
		),
		// Eating Out: $200/fortnightly, started Mon 31 Mar → next Mon 14 Apr (5 days away)
		// $140 of $200 spent → WARNING
		CreateTrickle(general.BucketID, eatingOut.BucketID, 20000, "fortnightly",
			time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
			WithTrickleDescription("Fortnightly dining budget"),
		),
		// Transport: $180/month, started 1 Apr → next 1 May (22 days away)
		// $81 of $180 spent → OK
		CreateTrickle(general.BucketID, transport.BucketID, 18000, "monthly",
			time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			WithTrickleDescription("Monthly transport"),
		),
		// Japan Trip: $1000/month, started 1 Apr → next 1 May (22 days away)
		// Only $180 spent → GREAT
		CreateTrickle(general.BucketID, japanTrip.BucketID, 100000, "monthly",
			time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			WithTrickleDescription("Japan trip savings"),
		),
		// Savings: NO trickle → STALE (intentionally omitted)
	}

	return user, buckets, txs, transfers, trickles
}

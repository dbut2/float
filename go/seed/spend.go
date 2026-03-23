package seed

import (
	"math/rand"

	"github.com/google/uuid"

	"dbut.dev/float/go/service"
)

type Merchant struct {
	Name     string
	MinCents int64
	MaxCents int64
}

var GroceryMerchants = []Merchant{
	{"Woolworths", 3000, 15000},
	{"Coles", 2500, 12000},
	{"Aldi", 2000, 8000},
	{"IGA", 1500, 6000},
}

var CoffeeMerchants = []Merchant{
	{"Coffee Shop", 450, 700},
	{"Starbucks", 500, 800},
	{"Gloria Jeans", 450, 750},
}

var SubscriptionMerchants = []Merchant{
	{"Netflix", 1699, 1699},
	{"Spotify", 1299, 1299},
	{"YouTube Premium", 1499, 1499},
}

var DiningMerchants = []Merchant{
	{"Uber Eats", 1500, 4500},
	{"Menulog", 1200, 3500},
	{"Pizza Hut", 1500, 3000},
	{"Guzman y Gomez", 1200, 2500},
}

func GenerateSpend(bucketID uuid.UUID, merchants []Merchant, opts ...TransactionOption) service.Transaction {
	m := merchants[rand.Intn(len(merchants))]
	amount := m.MinCents
	if m.MaxCents > m.MinCents {
		amount += rand.Int63n(m.MaxCents - m.MinCents)
	}
	return CreateExpense(bucketID, amount, append([]TransactionOption{WithDescription(m.Name)}, opts...)...)
}

package seed

import (
	"time"

	"github.com/google/uuid"

	"dbut.dev/float/go/service"
)

type UserOption func(*service.User)

func WithEmail(email string) UserOption {
	return func(u *service.User) {
		u.Email = email
	}
}

func CreateUser(opts ...UserOption) service.User {
	u := service.User{
		UserID:    uuid.New(),
		Email:     "seed@float.test",
		CreatedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(&u)
	}
	return u
}

type BucketOption func(*service.Bucket)

func WithCurrency(code string) BucketOption {
	return func(b *service.Bucket) {
		b.CurrencyCode = &code
	}
}

func WithBucketDescription(desc string) BucketOption {
	return func(b *service.Bucket) {
		b.Description = desc
	}
}

func CreateBucket(userID uuid.UUID, name string, opts ...BucketOption) service.Bucket {
	b := service.Bucket{
		BucketID:  uuid.New(),
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(&b)
	}
	return b
}

func CreateGeneralBucket(userID uuid.UUID) service.Bucket {
	return service.Bucket{
		BucketID:  uuid.New(),
		UserID:    userID,
		Name:      "General",
		IsGeneral: true,
		CreatedAt: time.Now(),
	}
}

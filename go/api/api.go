package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/service"
)

type UserService interface {
	GetUser(ctx context.Context, userID uuid.UUID) (service.User, error)
	UpdateToken(ctx context.Context, userID uuid.UUID, token string) error
	SyncTransactions(ctx context.Context, userID uuid.UUID) (int, error)
}

type BucketService interface {
	ListBuckets(ctx context.Context, userID uuid.UUID) ([]service.Bucket, error)
	CreateBucket(ctx context.Context, bucket service.Bucket) (service.Bucket, error)
	GetBucket(ctx context.Context, bucketID, userID uuid.UUID) (service.Bucket, error)
	DeleteBucket(ctx context.Context, bucketID, userID uuid.UUID) error
	ListBucketTransactions(ctx context.Context, bucketID uuid.UUID) ([]service.Transaction, error)
}

type TransactionService interface {
	ListTransactions(ctx context.Context, userID uuid.UUID) ([]service.Transaction, error)
	AssignToBucket(ctx context.Context, transactionID, bucketID uuid.UUID) error
}

type TransferService interface {
	ListTransfers(ctx context.Context, userID uuid.UUID) ([]service.Transfer, error)
	CreateTransfer(ctx context.Context, transfer service.Transfer) (service.Transfer, error)
	DeleteTransfer(ctx context.Context, transferID, userID uuid.UUID) error
}

type PushService interface {
	RegisterToken(ctx context.Context, userID uuid.UUID, token string) error
	UnregisterToken(ctx context.Context, userID uuid.UUID, token string) error
}

type API struct {
	users        UserService
	buckets      BucketService
	transactions TransactionService
	transfers    TransferService
	push         PushService
}

func New(q database.Querier) *API {
	return &API{
		users:        service.NewUserService(q),
		buckets:      service.NewBucketService(q),
		transactions: service.NewTransactionService(q),
		transfers:    service.NewTransferService(q),
		push:         service.NewPushService(q),
	}
}

func NewDemo() (*API, uuid.UUID) {
	demoService := service.NewDemoService()

	return &API{
		users:        demoService,
		buckets:      demoService,
		transactions: demoService,
		transfers:    demoService,
		push:         demoService,
	}, demoService.UserID()
}

func (a *API) Register(r *gin.RouterGroup) {
	r.GET("/user", a.getUser)
	r.PUT("/user/token", a.putUserToken)
	r.POST("/user/sync", a.postUserSync)

	r.GET("/buckets", a.listBuckets)
	r.POST("/buckets", a.createBucket)
	r.GET("/buckets/:bucketID", a.getBucket)
	r.DELETE("/buckets/:bucketID", a.deleteBucket)
	r.GET("/buckets/:bucketID/transactions", a.listBucketTransactions)

	r.GET("/transactions", a.listTransactions)
	r.PUT("/transactions/:transactionID/bucket", a.assignTransactionToBucket)

	r.GET("/transfers", a.listTransfers)
	r.POST("/transfers", a.createTransfer)
	r.DELETE("/transfers/:transferID", a.deleteTransfer)

	r.POST("/fcm-tokens", a.registerFCMToken)
	r.DELETE("/fcm-tokens", a.unregisterFCMToken)
}

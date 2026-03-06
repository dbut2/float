package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/frankfurter"
	"dbut.dev/float/go/service"
)

type UserService interface {
	GetUser(ctx context.Context, userID uuid.UUID) (service.User, error)
	UpdateToken(ctx context.Context, userID uuid.UUID, token string) error
	SyncTransactions(ctx context.Context, userID uuid.UUID) (int, error)
	GetTransactBalance(ctx context.Context, userID uuid.UUID) (int64, error)
}

type BucketService interface {
	ListBuckets(ctx context.Context, userID uuid.UUID) ([]service.Bucket, error)
	CreateBucket(ctx context.Context, bucket service.Bucket) (service.Bucket, error)
	GetBucket(ctx context.Context, bucketID, userID uuid.UUID) (service.Bucket, error)
	DeleteBucket(ctx context.Context, bucketID, userID uuid.UUID) error
	ListBucketTransactions(ctx context.Context, bucketID, userID uuid.UUID) ([]service.Transaction, error)
	ReorderBuckets(ctx context.Context, userID uuid.UUID, bucketIDs []uuid.UUID) error
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

type TrickleService interface {
	ListTrickles(ctx context.Context, userID uuid.UUID) ([]service.Trickle, error)
	GetTrickle(ctx context.Context, toBucketID, userID uuid.UUID) (service.Trickle, error)
	UpsertTrickle(ctx context.Context, trickle service.Trickle) (service.Trickle, error)
	DeleteTrickle(ctx context.Context, toBucketID, userID uuid.UUID) error
}

type RuleService interface {
	ListRules(ctx context.Context, userID uuid.UUID) ([]service.Rule, error)
	ListRulesByBucket(ctx context.Context, bucketID, userID uuid.UUID) ([]service.Rule, error)
	CreateRule(ctx context.Context, rule service.Rule) (service.Rule, error)
	UpdateRule(ctx context.Context, rule service.Rule, userID uuid.UUID) (service.Rule, error)
	DeleteRule(ctx context.Context, ruleID, userID uuid.UUID) error
	ApplyRulesToGeneral(ctx context.Context, userID uuid.UUID) (int, error)
}

type API struct {
	users        UserService
	buckets      BucketService
	transactions TransactionService
	transfers    TransferService
	push         PushService
	trickles     TrickleService
	rules        RuleService
}

func New(q database.Querier, fx *frankfurter.FXClient) *API {
	fxSvc := service.NewFXService(q, fx)
	return &API{
		users:        service.NewUserService(q),
		buckets:      service.NewBucketService(q, fxSvc),
		transactions: service.NewTransactionService(q),
		transfers:    service.NewTransferService(q),
		push:         service.NewPushService(q),
		trickles:     service.NewTrickleService(q),
		rules:        service.NewRuleService(q),
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
		trickles:     demoService,
		rules:        demoService,
	}, demoService.UserID()
}

func (a *API) Register(r *gin.RouterGroup) {
	r.GET("/user", a.getUser)
	r.PUT("/user/token", a.putUserToken)
	r.POST("/user/sync", a.postUserSync)
	r.GET("/user/balance", a.getTransactBalance)

	r.GET("/buckets", a.listBuckets)
	r.POST("/buckets", a.createBucket)
	r.PUT("/buckets/order", a.reorderBuckets)
	r.GET("/buckets/:bucketID", a.getBucket)
	r.DELETE("/buckets/:bucketID", a.deleteBucket)
	r.GET("/buckets/:bucketID/transactions", a.listBucketTransactions)
	r.GET("/buckets/:bucketID/rules", a.listBucketRules)
	r.POST("/buckets/:bucketID/rules", a.createRule)

	r.GET("/transactions", a.listTransactions)
	r.PUT("/transactions/:transactionID/bucket", a.assignTransactionToBucket)

	r.GET("/transfers", a.listTransfers)
	r.POST("/transfers", a.createTransfer)
	r.DELETE("/transfers/:transferID", a.deleteTransfer)

	r.GET("/trickles", a.listTrickles)
	r.GET("/buckets/:bucketID/trickle", a.getTrickle)
	r.PUT("/buckets/:bucketID/trickle", a.upsertTrickle)
	r.DELETE("/buckets/:bucketID/trickle", a.deleteTrickle)

	r.GET("/rules", a.listRules)
	r.PUT("/rules/:ruleID", a.updateRule)
	r.DELETE("/rules/:ruleID", a.deleteRule)
	r.POST("/rules/apply", a.applyRulesToGeneral)

	r.POST("/fcm-tokens", a.registerFCMToken)
	r.DELETE("/fcm-tokens", a.unregisterFCMToken)
}

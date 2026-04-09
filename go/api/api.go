package api

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/frankfurter"
	"dbut.dev/float/go/middleware"
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
	CloseBucket(ctx context.Context, bucketID, userID uuid.UUID) error
	ListBucketTransactions(ctx context.Context, bucketID, userID uuid.UUID) ([]service.Transaction, error)
	ReorderBuckets(ctx context.Context, userID uuid.UUID, bucketIDs []uuid.UUID) error
	UpdateBucketDescription(ctx context.Context, bucketID, userID uuid.UUID, description string) error
}

type TransactionService interface {
	ListTransactions(ctx context.Context, userID uuid.UUID) ([]service.Transaction, error)
	AssignToBucket(ctx context.Context, transactionID, bucketID, userID uuid.UUID) error
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

type ClassifierServiceInterface interface {
	ClassifyOne(ctx context.Context, userID, txID uuid.UUID) error
	StartReclassifyGeneral(userID uuid.UUID) bool
	GetReclassifyStatus(userID uuid.UUID) service.ReclassifyStatus
}

type CoverServiceInterface interface {
	CreateCover(ctx context.Context, userID, transactionID uuid.UUID, amountCents int64, note string) (service.Cover, error)
	DeleteCover(ctx context.Context, coverID, userID uuid.UUID) error
}

type API struct {
	users        UserService
	buckets      BucketService
	transactions TransactionService
	transfers    TransferService
	covers       CoverServiceInterface
	push         PushService
	trickles     TrickleService
	classifier   ClassifierServiceInterface
	health       HealthServiceInterface
}

func New(q database.Querier, fx *frankfurter.FXClient, classifier *service.ClassifierService, push *service.PushService) *API {
	fxSvc := service.NewFXService(q, fx)
	coverSvc := service.NewCoverService(q)
	return &API{
		users:        service.NewUserService(q, classifier),
		buckets:      service.NewBucketService(q, fxSvc, coverSvc),
		transactions: service.NewTransactionService(q),
		transfers:    service.NewTransferService(q),
		covers:       coverSvc,
		push:         push,
		trickles:     service.NewTrickleService(q),
		classifier:   classifier,
		health:       service.NewHealthService(q, push),
	}
}

func internalError(c *gin.Context, err error) {
	log.Printf("internal error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func (a *API) Register(r *gin.RouterGroup) {
	syncLimit := middleware.SyncRateLimit()

	r.GET("/user", a.getUser)
	r.PUT("/user/token", a.putUserToken)
	r.POST("/user/sync", syncLimit, a.postUserSync)
	r.GET("/user/balance", a.getTransactBalance)

	r.GET("/buckets", a.listBuckets)
	r.POST("/buckets", a.createBucket)
	r.PUT("/buckets/order", a.reorderBuckets)
	r.GET("/buckets/:bucketID", a.getBucket)
	r.DELETE("/buckets/:bucketID", a.deleteBucket)
	r.POST("/buckets/:bucketID/close", a.closeBucket)
	r.GET("/buckets/:bucketID/transactions", a.listBucketTransactions)
	r.PUT("/buckets/:bucketID/description", a.updateBucketDescription)

	r.GET("/transactions", a.listTransactions)
	r.PUT("/transactions/:transactionID/bucket", a.assignTransactionToBucket)

	r.GET("/transfers", a.listTransfers)
	r.POST("/transfers", a.createTransfer)
	r.DELETE("/transfers/:transferID", a.deleteTransfer)

	r.GET("/trickles", a.listTrickles)
	r.GET("/buckets/:bucketID/trickle", a.getTrickle)
	r.PUT("/buckets/:bucketID/trickle", a.upsertTrickle)
	r.DELETE("/buckets/:bucketID/trickle", a.deleteTrickle)

	r.POST("/transactions/:transactionID/covers", a.createCover)
	r.DELETE("/covers/:coverID", a.deleteCover)

	r.POST("/transactions/:transactionID/classify", a.classifyTransaction)
	r.POST("/classify/reclassify", a.reclassifyGeneral)
	r.GET("/classify/status", a.reclassifyStatus)

	r.POST("/fcm-tokens", a.registerFCMToken)
	r.DELETE("/fcm-tokens", a.unregisterFCMToken)

	r.GET("/health", a.getHealth)
	r.PUT("/buckets/:bucketID/trickle/apply-suggestion", a.applyTrickleSuggestion)
}

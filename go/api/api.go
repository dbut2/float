package api

import (
	"github.com/gin-gonic/gin"

	"dbut.dev/float/go/database"
)

type API struct {
	queries database.Querier
}

func New(queries database.Querier) *API {
	return &API{queries: queries}
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

func (a *API) getUser(c *gin.Context)                   { panic("not implemented") }
func (a *API) putUserToken(c *gin.Context)              { panic("not implemented") }
func (a *API) postUserSync(c *gin.Context)              { panic("not implemented") }
func (a *API) listBuckets(c *gin.Context)               { panic("not implemented") }
func (a *API) createBucket(c *gin.Context)              { panic("not implemented") }
func (a *API) getBucket(c *gin.Context)                 { panic("not implemented") }
func (a *API) deleteBucket(c *gin.Context)              { panic("not implemented") }
func (a *API) listBucketTransactions(c *gin.Context)    { panic("not implemented") }
func (a *API) listTransactions(c *gin.Context)          { panic("not implemented") }
func (a *API) assignTransactionToBucket(c *gin.Context) { panic("not implemented") }
func (a *API) listTransfers(c *gin.Context)             { panic("not implemented") }
func (a *API) createTransfer(c *gin.Context)            { panic("not implemented") }
func (a *API) deleteTransfer(c *gin.Context)            { panic("not implemented") }
func (a *API) registerFCMToken(c *gin.Context)          { panic("not implemented") }
func (a *API) unregisterFCMToken(c *gin.Context)        { panic("not implemented") }

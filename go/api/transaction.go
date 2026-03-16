package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/middleware"
)

func (a *API) listTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)

	txs, err := a.transactions.ListTransactions(c.Request.Context(), userID)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, txs)
}

func (a *API) assignTransactionToBucket(c *gin.Context) {
	userID := middleware.GetUserID(c)

	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var body struct {
		BucketID string `json:"bucket_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bucketID, err := uuid.Parse(body.BucketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	if err := a.transactions.AssignToBucket(c.Request.Context(), transactionID, bucketID, userID); err != nil {
		internalError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/middleware"
	"dbut.dev/float/go/service"
)

func (a *API) listBuckets(c *gin.Context) {
	userID := middleware.GetUserID(c)

	buckets, err := a.buckets.ListBuckets(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, buckets)
}

func (a *API) createBucket(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var body struct {
		Name         string  `json:"name"`
		CurrencyCode *string `json:"currency_code"`
		Description  string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bucket, err := a.buckets.CreateBucket(c.Request.Context(), service.Bucket{
		UserID:       userID,
		Name:         body.Name,
		CurrencyCode: body.CurrencyCode,
		Description:  body.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bucket)
}

func (a *API) getBucket(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	bucket, err := a.buckets.GetBucket(c.Request.Context(), bucketID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bucket)
}

func (a *API) deleteBucket(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	if err := a.buckets.DeleteBucket(c.Request.Context(), bucketID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *API) closeBucket(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	if err := a.buckets.CloseBucket(c.Request.Context(), bucketID, userID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *API) reorderBuckets(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var body struct {
		BucketIDs []uuid.UUID `json:"bucket_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.buckets.ReorderBuckets(c.Request.Context(), userID, body.BucketIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *API) listBucketTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	txs, err := a.buckets.ListBucketTransactions(c.Request.Context(), bucketID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, txs)
}

func (a *API) updateBucketDescription(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	var body struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.buckets.UpdateBucketDescription(c.Request.Context(), bucketID, userID, body.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *API) reclassifyGeneral(c *gin.Context) {
	userID := middleware.GetUserID(c)

	started := a.classifier.StartReclassifyGeneral(userID)
	if !started {
		c.JSON(http.StatusConflict, gin.H{"error": "reclassification already in progress"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "started"})
}

func (a *API) reclassifyStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	status := a.classifier.GetReclassifyStatus(userID)
	c.JSON(http.StatusOK, status)
}

func (a *API) classifyTransaction(c *gin.Context) {
	userID := middleware.GetUserID(c)

	txID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	if err := a.classifier.ClassifyOne(c.Request.Context(), userID, txID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

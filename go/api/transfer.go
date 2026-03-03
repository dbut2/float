package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/middleware"
	"dbut.dev/float/go/service"
)

func (a *API) listTransfers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	transfers, err := a.transfers.ListTransfers(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfers)
}

func (a *API) createTransfer(c *gin.Context) {
	var body struct {
		FromBucketID string `json:"from_bucket_id"`
		ToBucketID   string `json:"to_bucket_id"`
		AmountCents  int64  `json:"amount_cents"`
		Note         string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fromBucketID, err := uuid.Parse(body.FromBucketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_bucket_id"})
		return
	}

	toBucketID, err := uuid.Parse(body.ToBucketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_bucket_id"})
		return
	}

	transfer, err := a.transfers.CreateTransfer(c.Request.Context(), service.Transfer{
		FromBucketID: fromBucketID,
		ToBucketID:   toBucketID,
		AmountCents:  body.AmountCents,
		Note:         body.Note,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transfer)
}

func (a *API) deleteTransfer(c *gin.Context) {
	userID := middleware.GetUserID(c)

	transferID, err := uuid.Parse(c.Param("transferID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer ID"})
		return
	}

	if err := a.transfers.DeleteTransfer(c.Request.Context(), transferID, userID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

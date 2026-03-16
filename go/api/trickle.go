package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/middleware"
	"dbut.dev/float/go/service"
)

func (a *API) listTrickles(c *gin.Context) {
	userID := middleware.GetUserID(c)

	trickles, err := a.trickles.ListTrickles(c.Request.Context(), userID)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, trickles)
}

func (a *API) getTrickle(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	trickle, err := a.trickles.GetTrickle(c.Request.Context(), bucketID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, trickle)
}

func (a *API) upsertTrickle(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	var body struct {
		AmountCents int64   `json:"amount_cents"`
		Description string  `json:"description"`
		Period      string  `json:"period"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date, expected YYYY-MM-DD"})
		return
	}

	var endDate *time.Time
	if body.EndDate != nil {
		t, err := time.Parse("2006-01-02", *body.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date, expected YYYY-MM-DD"})
			return
		}
		endDate = &t
	}

	trickle, err := a.trickles.UpsertTrickle(c.Request.Context(), service.Trickle{
		UserID:      userID,
		ToBucketID:  bucketID,
		AmountCents: body.AmountCents,
		Description: body.Description,
		Period:      body.Period,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trickle)
}

func (a *API) deleteTrickle(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	if err := a.trickles.DeleteTrickle(c.Request.Context(), bucketID, userID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		internalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/middleware"
	"dbut.dev/float/go/service"
)

type HealthServiceInterface interface {
	GetHealthSummary(ctx context.Context, userID uuid.UUID) (service.HealthSummary, error)
}

func (a *API) getHealth(c *gin.Context) {
	userID := middleware.GetUserID(c)

	summary, err := a.health.GetHealthSummary(c.Request.Context(), userID)
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (a *API) applyTrickleSuggestion(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket ID"})
		return
	}

	var body struct {
		AmountCents int64  `json:"amount_cents"`
		Period      string `json:"period"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing trickle to determine a sensible start date.
	existing, err := a.trickles.GetTrickle(c.Request.Context(), bucketID, userID)
	if err != nil && !errors.Is(err, service.ErrNotFound) {
		internalError(c, err)
		return
	}

	// Re-use the existing start date if there is one, otherwise the trickle
	// upsert logic will compute the correct next occurrence.
	startDate := existing.StartDate

	trickle, err := a.trickles.UpsertTrickle(c.Request.Context(), service.Trickle{
		UserID:      userID,
		ToBucketID:  bucketID,
		AmountCents: body.AmountCents,
		Description: existing.Description,
		Period:      body.Period,
		StartDate:   startDate,
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

package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/middleware"
	"dbut.dev/float/go/service"
)

func (a *API) createCover(c *gin.Context) {
	userID := middleware.GetUserID(c)

	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var body struct {
		AmountCents int64  `json:"amount_cents"`
		Note        string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.AmountCents <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount_cents must be positive"})
		return
	}

	cover, err := a.covers.CreateCover(c.Request.Context(), userID, transactionID, body.AmountCents, body.Note)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
			return
		}
		if err.Error() == "only debit transactions can be covered" || err.Error() == "cover amount exceeds transaction amount" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, cover)
}

func (a *API) deleteCover(c *gin.Context) {
	userID := middleware.GetUserID(c)

	coverID, err := uuid.Parse(c.Param("coverID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cover ID"})
		return
	}

	if err := a.covers.DeleteCover(c.Request.Context(), coverID, userID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		internalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

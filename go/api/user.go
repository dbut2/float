package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"dbut.dev/float/go/middleware"
	"dbut.dev/float/go/service"
	"dbut.dev/float/go/utils"
)

func (a *API) getUser(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := a.users.GetUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (a *API) putUserToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var body struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.users.UpdateToken(c.Request.Context(), userID, body.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *API) getTransactBalance(c *gin.Context) {
	userID := middleware.GetUserID(c)

	balance, err := a.users.GetTransactBalance(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrTokenNotSet) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no token set"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance_cents":   balance,
		"balance_display": utils.FormatAmount(balance, "AUD"),
	})
}

func (a *API) postUserSync(c *gin.Context) {
	userID := middleware.GetUserID(c)

	synced, err := a.users.SyncTransactions(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrTokenNotSet) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no token set"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"synced": synced})
}

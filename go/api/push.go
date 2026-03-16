package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"dbut.dev/float/go/middleware"
)

func (a *API) registerFCMToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var body struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.push.RegisterToken(c.Request.Context(), userID, body.Token); err != nil {
		internalError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (a *API) unregisterFCMToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var body struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.push.UnregisterToken(c.Request.Context(), userID, body.Token); err != nil {
		internalError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

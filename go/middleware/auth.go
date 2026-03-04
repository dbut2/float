package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/service"
)

func Middleware(queries database.Querier, baseURL string) gin.HandlerFunc {
	webhookSvc := service.NewWebhookService(queries)
	return func(c *gin.Context) {
		email := c.GetHeader("Cf-Access-Authenticated-User-Email")

		if email == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, err := queries.UpsertUser(c.Request.Context(), email)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to upsert user"})
			return
		}

		if err := queries.EnsureGeneralBucket(c.Request.Context(), user.UserID); err != nil {
			log.Printf("failed to ensure General bucket for user %s: %v", user.UserID, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure General bucket"})
			return
		}

		c.Set("user_id", user.UserID)

		if baseURL != "" {
			if err := webhookSvc.EnsureWebhook(c.Request.Context(), user.UserID, baseURL); err != nil {
				log.Printf("failed to ensure webhook for user %s: %v", user.UserID, err)
			}
		}

		c.Next()
	}
}

func DemoAuth(userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uuid.UUID {
	return c.MustGet("user_id").(uuid.UUID)
}

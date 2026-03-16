package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/service"
)

func Middleware(queries database.Querier, baseURL string) gin.HandlerFunc {
	webhookSvc := service.NewWebhookService(queries, nil, nil)

	var emailToUserID sync.Map
	var generalBucketEnsured sync.Map
	var webhookConfirmed sync.Map

	return func(c *gin.Context) {
		email := c.GetHeader("Cf-Access-Authenticated-User-Email")

		if email == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var userID uuid.UUID
		if cached, ok := emailToUserID.Load(email); ok {
			userID = cached.(uuid.UUID)
		} else {
			user, err := queries.UpsertUser(c.Request.Context(), email)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to upsert user"})
				return
			}
			userID = user.UserID
			emailToUserID.Store(email, userID)
		}

		if _, ok := generalBucketEnsured.Load(userID); !ok {
			if err := queries.EnsureGeneralBucket(c.Request.Context(), userID); err != nil {
				log.Printf("failed to ensure General bucket for user %s: %v", userID, err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure General bucket"})
				return
			}
			generalBucketEnsured.Store(userID, struct{}{})
		}

		c.Set("user_id", userID)

		if baseURL != "" {
			if _, ok := webhookConfirmed.Load(userID); !ok {
				confirmed, err := webhookSvc.EnsureWebhook(c.Request.Context(), userID, baseURL)
				if err != nil {
					log.Printf("failed to ensure webhook for user %s: %v", userID, err)
				}
				if confirmed {
					webhookConfirmed.Store(userID, struct{}{})
				}
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

var syncRateLimiters sync.Map

// SyncRateLimit returns a per-user rate limiter middleware: 1 request per minute.
func SyncRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		v, _ := syncRateLimiters.LoadOrStore(userID, rate.NewLimiter(rate.Every(time.Minute), 1))
		if !v.(*rate.Limiter).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

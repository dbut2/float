package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/service"
	"dbut.dev/float/go/up"
)

type Service interface {
	GetWebhookSecret(ctx context.Context, userID uuid.UUID) (string, error)
	ProcessEvent(ctx context.Context, userID uuid.UUID, payload up.WebhookPayload) error
}

type Handler struct {
	svc Service
}

func New(q database.Querier, classifier *service.ClassifierService) *Handler {
	return &Handler{svc: service.NewWebhookService(q, classifier)}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/up/:userID", h.handle)
}

func (h *Handler) handle(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read body"})
		return
	}

	secret, err := h.svc.GetWebhookSecret(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get webhook secret"})
		return
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(c.GetHeader("X-Up-Authenticity-Signature")), []byte(expected)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	var payload up.WebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.svc.ProcessEvent(c.Request.Context(), userID, payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process event"})
		return
	}

	c.Status(http.StatusOK)
}

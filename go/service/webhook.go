package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
	"dbut.dev/float/go/up"
	"dbut.dev/float/go/utils"
)

type WebhookService struct {
	q          database.Querier
	classifier *ClassifierService
	push       *PushService
	health     *HealthService
}

func NewWebhookService(q database.Querier, classifier *ClassifierService, push *PushService) *WebhookService {
	return &WebhookService{q: q, classifier: classifier, push: push, health: NewHealthService(q, push)}
}

func (s *WebhookService) EnsureWebhook(ctx context.Context, userID uuid.UUID, baseURL string) (bool, error) {
	nullToken, err := s.q.GetUserToken(ctx, userID)
	if err != nil {
		return false, err
	}
	if !nullToken.Valid {
		return false, nil
	}

	client, err := up.NewUpClient(nullToken.String)
	if err != nil {
		return false, err
	}

	webhookURL := baseURL + "/webhook/up/" + userID.String()

	existing, err := client.ListWebhooks(ctx)
	if err != nil {
		return false, err
	}

	// If the correct webhook is already registered, nothing to do.
	for _, wh := range existing {
		if wh.Attributes.Url == webhookURL {
			return true, nil
		}
	}

	// Correct webhook not found — delete any stale float webhooks and re-register.
	for _, wh := range existing {
		if strings.HasPrefix(wh.Attributes.Url, baseURL+"/webhook/up/") {
			if err := client.DeleteWebhook(ctx, wh.Id); err != nil {
				log.Printf("failed to delete stale webhook %s: %v", wh.Id, err)
			}
		}
	}

	_, secretKey, err := client.RegisterWebhook(ctx, webhookURL)
	if err != nil {
		return false, err
	}

	if err := s.q.SetUserWebhookSecret(ctx, userID, sql.NullString{String: secretKey, Valid: true}); err != nil {
		return false, err
	}
	return true, nil
}

func (s *WebhookService) GetWebhookSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	ns, err := s.q.GetUserWebhookSecret(ctx, userID)
	if err != nil {
		return "", err
	}
	if !ns.Valid {
		return "", errors.New("webhook secret not set")
	}
	return ns.String, nil
}

func (s *WebhookService) ProcessEvent(ctx context.Context, userID uuid.UUID, payload up.WebhookPayload) error {
	eventType, _ := payload.Data.Attributes.EventType.(string)

	switch eventType {
	case "PING":
		return nil

	case "TRANSACTION_CREATED", "TRANSACTION_SETTLED":
		if payload.Data.Relationships.Transaction == nil {
			return errors.New("missing transaction relationship")
		}
		txID := payload.Data.Relationships.Transaction.Data.Id

		log.Printf("webhook %s: tx %s for user %s", eventType, txID, userID)

		nullToken, err := s.q.GetUserToken(ctx, userID)
		if err != nil {
			return err
		}
		if !nullToken.Valid {
			return ErrTokenNotSet
		}

		client, err := up.NewUpClient(nullToken.String)
		if err != nil {
			return err
		}

		tx, err := client.GetTransaction(ctx, txID)
		if err != nil {
			return err
		}

		params, err := upTransactionToParams(userID, *tx)
		if err != nil {
			return err
		}

		inserted, err := s.q.UpsertUpTransaction(ctx, params)
		if err != nil {
			return err
		}

		if inserted && eventType == "TRANSACTION_CREATED" {
			go s.push.SendNotification(context.Background(), userID,
				params.Description,
				utils.FormatSignedAmount(params.AmountCents, "AUD"),
			)
		}

		// Classify on CREATED if new, or on SETTLED always (SETTLED has full
		// category data which improves classification accuracy).
		shouldClassify := inserted || eventType == "TRANSACTION_SETTLED"
		if shouldClassify && s.classifier != nil {
			log.Printf("webhook: queuing classification for tx %s (inserted=%v, event=%s)", params.TransactionID, inserted, eventType)
			go func() {
				bgCtx := context.Background()
				if err := s.classifier.ClassifyTransaction(bgCtx, userID, params.TransactionID, params.Description, params.AmountCents, params.TransactionType, params.CategoryID, params.CreatedAt, params.ForeignCurrencyCode); err != nil {
					log.Printf("classify tx %s: %v", params.TransactionID, err)
					return
				}
				log.Printf("classify tx %s: done", params.TransactionID)
				// After classification, check if the assigned bucket is now critical
				// and send a health push notification if needed (once per day max).
				if s.health != nil {
					txRow, err := s.q.GetTransaction(bgCtx, params.TransactionID, userID)
					if err == nil {
						s.health.CheckAndNotify(bgCtx, userID, txRow.BucketID)
					}
				}
			}()
		}
		return nil

	case "TRANSACTION_DELETED":
		if payload.Data.Relationships.Transaction == nil {
			return errors.New("missing transaction relationship")
		}
		txID, err := uuid.Parse(payload.Data.Relationships.Transaction.Data.Id)
		if err != nil {
			return err
		}
		return s.q.DeleteUpTransaction(ctx, txID)

	default:
		log.Printf("unhandled webhook event type: %s", eventType)
		return nil
	}
}

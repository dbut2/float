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
)

type WebhookService struct {
	q database.Querier
}

func NewWebhookService(q database.Querier) *WebhookService {
	return &WebhookService{q: q}
}

func (s *WebhookService) EnsureWebhook(ctx context.Context, userID uuid.UUID, baseURL string) error {
	secret, err := s.q.GetUserWebhookSecret(ctx, userID)
	if err != nil {
		return err
	}
	if secret.Valid {
		return nil
	}

	nullToken, err := s.q.GetUserToken(ctx, userID)
	if err != nil {
		return err
	}
	if !nullToken.Valid {
		return nil
	}

	client, err := up.NewUpClient(nullToken.String)
	if err != nil {
		return err
	}

	webhookURL := baseURL + "/webhook/up/" + userID.String()

	existing, err := client.ListWebhooks(ctx)
	if err != nil {
		return err
	}
	for _, wh := range existing {
		if strings.HasPrefix(wh.Attributes.Url, baseURL+"/webhook/up/") {
			if err := client.DeleteWebhook(ctx, wh.Id); err != nil {
				log.Printf("failed to delete stale webhook %s: %v", wh.Id, err)
			}
		}
	}

	_, secretKey, err := client.RegisterWebhook(ctx, webhookURL)
	if err != nil {
		return err
	}

	return s.q.SetUserWebhookSecret(ctx, userID, sql.NullString{String: secretKey, Valid: true})
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

		_, err = s.q.UpsertUpTransaction(ctx, params)
		return err

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

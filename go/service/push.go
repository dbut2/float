package service

import (
	"context"
	"log"

	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

type PushService struct {
	q   database.Querier
	fcm *messaging.Client
}

func NewPushService(q database.Querier, fcm *messaging.Client) *PushService {
	return &PushService{q: q, fcm: fcm}
}

func (s *PushService) RegisterToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.q.RegisterFCMToken(ctx, userID, token)
}

func (s *PushService) UnregisterToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.q.UnregisterFCMToken(ctx, userID, token)
}

func (s *PushService) SendNotification(ctx context.Context, userID uuid.UUID, title, body string) {
	if s.fcm == nil {
		return
	}
	tokens, err := s.q.GetUserFCMTokens(ctx, userID)
	if err != nil || len(tokens) == 0 {
		if err != nil {
			log.Printf("push: get tokens for user %s: %v", userID, err)
		}
		return
	}
	resp, err := s.fcm.SendEachForMulticast(ctx, &messaging.MulticastMessage{
		Notification: &messaging.Notification{Title: title, Body: body},
		Tokens:       tokens,
	})
	if err != nil {
		log.Printf("push: send FCM for user %s: %v", userID, err)
		return
	}
	log.Printf("push: sent %d/%d for user %s", resp.SuccessCount, len(tokens), userID)
}

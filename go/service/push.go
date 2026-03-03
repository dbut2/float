package service

import (
	"context"

	"github.com/google/uuid"

	"dbut.dev/float/go/database"
)

type PushService struct {
	q database.Querier
}

func NewPushService(q database.Querier) *PushService {
	return &PushService{q: q}
}

func (s *PushService) RegisterToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.q.RegisterFCMToken(ctx, userID, token)
}

func (s *PushService) UnregisterToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.q.UnregisterFCMToken(ctx, userID, token)
}

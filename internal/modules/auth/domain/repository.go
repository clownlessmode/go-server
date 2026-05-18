package domain

import (
	"context"
	"time"
)

type RefreshSession struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Repository interface {
	CreateRefreshSession(ctx context.Context, session *RefreshSession) (*RefreshSession, error)
	GetRefreshSessionByHash(ctx context.Context, tokenHash string) (*RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenHash string) error
}

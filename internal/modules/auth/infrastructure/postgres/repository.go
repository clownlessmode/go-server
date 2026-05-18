package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"project/internal/modules/auth/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateRefreshSession(ctx context.Context, session *domain.RefreshSession) (*domain.RefreshSession, error) {
	created := &domain.RefreshSession{}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO refresh_sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at
	`, session.UserID, session.TokenHash, session.ExpiresAt).Scan(
		&created.ID,
		&created.UserID,
		&created.TokenHash,
		&created.ExpiresAt,
		&created.RevokedAt,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create refresh session: %w", err)
	}

	return created, nil
}

func (r *Repository) GetRefreshSessionByHash(ctx context.Context, tokenHash string) (*domain.RefreshSession, error) {
	session := &domain.RefreshSession{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_sessions
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh session: %w", err)
	}

	return session, nil
}

func (r *Repository) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE token_hash = $1
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh session: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke refresh session affected rows: %w", err)
	}
	if affected == 0 {
		return domain.ErrInvalidToken
	}

	return nil
}

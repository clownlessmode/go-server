package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"project/internal/modules/access/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByUserID(ctx context.Context, userID int64) ([]*domain.AccessGrant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ag.id, ag.user_id, ag.bank_id, b.code, b.name,
			ag.granted_at, ag.expires_at, ag.grant_reason,
			ag.revoked_at, ag.revoke_reason, ag.created_at, ag.updated_at
		FROM access_grants ag
		JOIN bank_catalog b ON b.id = ag.bank_id
		WHERE ag.user_id = $1
		ORDER BY ag.granted_at DESC, ag.id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user accesses: %w", err)
	}
	defer rows.Close()

	accesses := make([]*domain.AccessGrant, 0)
	for rows.Next() {
		access, err := scanAccessGrant(rows)
		if err != nil {
			return nil, err
		}

		accesses = append(accesses, access)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user accesses rows: %w", err)
	}

	return accesses, nil
}

func (r *Repository) HasActiveAccess(ctx context.Context, userID int64, bankID int64) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM access_grants
			WHERE user_id = $1
				AND bank_id = $2
				AND revoked_at IS NULL
				AND expires_at > NOW()
		)
	`, userID, bankID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check active access: %w", err)
	}

	return exists, nil
}

func (r *Repository) Grant(ctx context.Context, access *domain.AccessGrant) (*domain.AccessGrant, error) {
	created := &domain.AccessGrant{}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO access_grants (user_id, bank_id, expires_at, grant_reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, bank_id, '', '', granted_at, expires_at, grant_reason,
			revoked_at, revoke_reason, created_at, updated_at
	`, access.UserID, access.BankID, access.ExpiresAt, access.GrantReason).Scan(
		&created.ID,
		&created.UserID,
		&created.BankID,
		&created.BankCode,
		&created.BankName,
		&created.GrantedAt,
		&created.ExpiresAt,
		&created.GrantReason,
		&created.RevokedAt,
		&created.RevokeReason,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return nil, domain.ErrAccessAlreadyExists
	}
	if err != nil {
		return nil, fmt.Errorf("grant access: %w", err)
	}

	return r.getByID(ctx, created.ID)
}

func (r *Repository) Revoke(ctx context.Context, userID int64, bankID int64, reason string) (*domain.AccessGrant, error) {
	revoked := &domain.AccessGrant{}

	err := r.db.QueryRowContext(ctx, `
		UPDATE access_grants
		SET revoked_at = NOW(),
			revoke_reason = $3,
			updated_at = NOW()
		WHERE user_id = $1
			AND bank_id = $2
			AND revoked_at IS NULL
		RETURNING id, user_id, bank_id, '', '', granted_at, expires_at, grant_reason,
			revoked_at, revoke_reason, created_at, updated_at
	`, userID, bankID, reason).Scan(
		&revoked.ID,
		&revoked.UserID,
		&revoked.BankID,
		&revoked.BankCode,
		&revoked.BankName,
		&revoked.GrantedAt,
		&revoked.ExpiresAt,
		&revoked.GrantReason,
		&revoked.RevokedAt,
		&revoked.RevokeReason,
		&revoked.CreatedAt,
		&revoked.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccessNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("revoke access: %w", err)
	}

	return r.getByID(ctx, revoked.ID)
}

func (r *Repository) getByID(ctx context.Context, id int64) (*domain.AccessGrant, error) {
	access := &domain.AccessGrant{}

	err := r.db.QueryRowContext(ctx, `
		SELECT ag.id, ag.user_id, ag.bank_id, b.code, b.name,
			ag.granted_at, ag.expires_at, ag.grant_reason,
			ag.revoked_at, ag.revoke_reason, ag.created_at, ag.updated_at
		FROM access_grants ag
		JOIN bank_catalog b ON b.id = ag.bank_id
		WHERE ag.id = $1
	`, id).Scan(
		&access.ID,
		&access.UserID,
		&access.BankID,
		&access.BankCode,
		&access.BankName,
		&access.GrantedAt,
		&access.ExpiresAt,
		&access.GrantReason,
		&access.RevokedAt,
		&access.RevokeReason,
		&access.CreatedAt,
		&access.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccessNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get access by id: %w", err)
	}

	return access, nil
}

type accessGrantScanner interface {
	Scan(dest ...any) error
}

func scanAccessGrant(scanner accessGrantScanner) (*domain.AccessGrant, error) {
	access := &domain.AccessGrant{}
	if err := scanner.Scan(
		&access.ID,
		&access.UserID,
		&access.BankID,
		&access.BankCode,
		&access.BankName,
		&access.GrantedAt,
		&access.ExpiresAt,
		&access.GrantReason,
		&access.RevokedAt,
		&access.RevokeReason,
		&access.CreatedAt,
		&access.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan access grant: %w", err)
	}

	return access, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

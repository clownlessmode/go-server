package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"project/internal/modules/banks/beeline/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetConfig(ctx context.Context) (*domain.Config, error) {
	config := &domain.Config{}

	err := r.db.QueryRowContext(ctx, `
		SELECT created_at, updated_at
		FROM beeline_configs
		WHERE id = 1
	`).Scan(
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &domain.Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get beeline config: %w", err)
	}

	return config, nil
}

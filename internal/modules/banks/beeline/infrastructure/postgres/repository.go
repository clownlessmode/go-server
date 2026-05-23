package postgres

import (
	"database/sql"

	"project/internal/modules/banks/beeline/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

var _ domain.Repository = (*Repository)(nil)

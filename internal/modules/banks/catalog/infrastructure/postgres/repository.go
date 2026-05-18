package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"project/internal/modules/banks/catalog/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]*domain.Bank, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, created_at, updated_at
		FROM bank_catalog
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list banks: %w", err)
	}
	defer rows.Close()

	banks := make([]*domain.Bank, 0)
	for rows.Next() {
		bank, err := scanBank(rows)
		if err != nil {
			return nil, err
		}

		banks = append(banks, bank)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list banks rows: %w", err)
	}

	return banks, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.Bank, error) {
	bank := &domain.Bank{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, created_at, updated_at
		FROM bank_catalog
		WHERE id = $1
	`, id).Scan(
		&bank.ID,
		&bank.Code,
		&bank.Name,
		&bank.CreatedAt,
		&bank.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrBankNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get bank by id: %w", err)
	}

	return bank, nil
}

type bankScanner interface {
	Scan(dest ...any) error
}

func scanBank(scanner bankScanner) (*domain.Bank, error) {
	bank := &domain.Bank{}
	if err := scanner.Scan(
		&bank.ID,
		&bank.Code,
		&bank.Name,
		&bank.CreatedAt,
		&bank.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan bank: %w", err)
	}

	return bank, nil
}

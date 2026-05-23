package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"project/internal/modules/banks/beeline/domain"
)

func (r *Repository) CreateSim(ctx context.Context, sim domain.Sim) (domain.Sim, error) {
	created := sim

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO beeline_sims (number, balance)
		VALUES ($1, $2)
		RETURNING balance, created_at, updated_at
	`, created.Number, created.Balance).Scan(
		&created.Balance,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Sim{}, domain.ErrSimAlreadyExists
		}
		return domain.Sim{}, fmt.Errorf("create beeline sim: %w", err)
	}

	return created, nil
}

func (r *Repository) EnsureSim(ctx context.Context, number string) (domain.Sim, error) {
	number = domain.NormalizeSimNumber(number)
	if err := domain.ValidateSimNumber(number); err != nil {
		return domain.Sim{}, err
	}

	sim, err := r.GetSim(ctx, number)
	if err == nil {
		return sim, nil
	}
	if !errors.Is(err, domain.ErrSimNotFound) {
		return domain.Sim{}, err
	}

	return r.CreateSim(ctx, domain.Sim{Number: number})
}

func (r *Repository) ListSims(ctx context.Context) ([]domain.Sim, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT number, balance, created_at, updated_at
		FROM beeline_sims
		ORDER BY number ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list beeline sims: %w", err)
	}
	defer rows.Close()

	sims := make([]domain.Sim, 0)
	for rows.Next() {
		sim, err := scanSim(rows)
		if err != nil {
			return nil, err
		}
		sims = append(sims, sim)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list beeline sims rows: %w", err)
	}

	return sims, nil
}

func (r *Repository) GetSim(ctx context.Context, number string) (domain.Sim, error) {
	number = domain.NormalizeSimNumber(number)
	row := r.db.QueryRowContext(ctx, `
		SELECT number, balance, created_at, updated_at
		FROM beeline_sims
		WHERE number = $1
	`, number)

	sim, err := scanSim(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Sim{}, domain.ErrSimNotFound
	}
	if err != nil {
		return domain.Sim{}, fmt.Errorf("get beeline sim: %w", err)
	}

	return sim, nil
}

func (r *Repository) DeleteSim(ctx context.Context, number string) error {
	number = domain.NormalizeSimNumber(number)
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM beeline_sims
		WHERE number = $1
	`, number)
	if err != nil {
		return fmt.Errorf("delete beeline sim: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete beeline sim rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSimNotFound
	}

	return nil
}

func (r *Repository) UpdateBalance(ctx context.Context, number string, balance *float64) (domain.Sim, error) {
	number = domain.NormalizeSimNumber(number)
	sim := domain.Sim{}

	err := r.db.QueryRowContext(ctx, `
		UPDATE beeline_sims
		SET balance = $2,
			updated_at = NOW()
		WHERE number = $1
		RETURNING number, balance, created_at, updated_at
	`, number, balance).Scan(
		&sim.Number,
		&sim.Balance,
		&sim.CreatedAt,
		&sim.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Sim{}, domain.ErrSimNotFound
	}
	if err != nil {
		return domain.Sim{}, fmt.Errorf("update beeline balance: %w", err)
	}

	return sim, nil
}

func (r *Repository) GetEffectiveBalance(ctx context.Context, number string) (*float64, error) {
	sim, err := r.GetSim(ctx, number)
	if err != nil {
		return nil, err
	}
	if sim.Balance == nil {
		return nil, nil
	}

	total, err := r.SumPaymentTotals(ctx, number)
	if err != nil {
		return nil, err
	}

	incoming, err := r.SumIncomingTotals(ctx, number)
	if err != nil {
		return nil, err
	}

	return domain.EffectiveBalance(sim.Balance, total, incoming), nil
}

func (r *Repository) FindConfiguredSimAmong(ctx context.Context, numbers []string) (string, bool) {
	for _, number := range numbers {
		number = domain.NormalizeSimNumber(number)
		if err := domain.ValidateSimNumber(number); err != nil {
			continue
		}

		if _, err := r.GetDetalizationSnapshot(ctx, number); err == nil {
			return number, true
		}

		sim, err := r.GetSim(ctx, number)
		if err != nil {
			continue
		}
		if sim.Balance != nil {
			return number, true
		}

		payments, err := r.ListPayments(ctx, number)
		if err == nil && len(payments) > 0 {
			return number, true
		}
	}

	return "", false
}

type simRowScanner interface {
	Scan(dest ...any) error
}

func scanSim(scanner simRowScanner) (domain.Sim, error) {
	sim := domain.Sim{}

	err := scanner.Scan(
		&sim.Number,
		&sim.Balance,
		&sim.CreatedAt,
		&sim.UpdatedAt,
	)
	if err != nil {
		return domain.Sim{}, err
	}

	return sim, nil
}

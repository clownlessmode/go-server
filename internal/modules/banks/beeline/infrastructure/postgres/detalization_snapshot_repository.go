package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"project/internal/modules/banks/beeline/domain"
)

func (r *Repository) SaveDetalizationSnapshot(ctx context.Context, snapshot domain.DetalizationSnapshot) (domain.DetalizationSnapshot, error) {
	number := domain.NormalizeSimNumber(snapshot.SimNumber)
	saved := snapshot
	saved.SimNumber = number

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO beeline_detalization_snapshots (
			sim_number, period_start, period_end, snapshot, computed_balance
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sim_number) DO UPDATE
		SET period_start = EXCLUDED.period_start,
			period_end = EXCLUDED.period_end,
			snapshot = EXCLUDED.snapshot,
			computed_balance = EXCLUDED.computed_balance,
			updated_at = NOW()
		RETURNING computed_balance, created_at, updated_at
	`,
		saved.SimNumber,
		saved.PeriodStart,
		saved.PeriodEnd,
		saved.Data,
		saved.ComputedBalance,
	).Scan(&saved.ComputedBalance, &saved.CreatedAt, &saved.UpdatedAt)
	if err != nil {
		return domain.DetalizationSnapshot{}, fmt.Errorf("save beeline detalization snapshot: %w", err)
	}

	return saved, nil
}

func (r *Repository) GetDetalizationSnapshot(ctx context.Context, number string) (domain.DetalizationSnapshot, error) {
	number = domain.NormalizeSimNumber(number)

	snapshot := domain.DetalizationSnapshot{}
	err := r.db.QueryRowContext(ctx, `
		SELECT sim_number, period_start, period_end, snapshot, computed_balance, created_at, updated_at
		FROM beeline_detalization_snapshots
		WHERE sim_number = $1
	`, number).Scan(
		&snapshot.SimNumber,
		&snapshot.PeriodStart,
		&snapshot.PeriodEnd,
		&snapshot.Data,
		&snapshot.ComputedBalance,
		&snapshot.CreatedAt,
		&snapshot.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DetalizationSnapshot{}, domain.ErrDetalizationSnapshotNotFound
	}
	if err != nil {
		return domain.DetalizationSnapshot{}, fmt.Errorf("get beeline detalization snapshot: %w", err)
	}

	return snapshot, nil
}

func (r *Repository) HasDetalizationSnapshot(ctx context.Context, number string) (bool, error) {
	number = domain.NormalizeSimNumber(number)

	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM beeline_detalization_snapshots
			WHERE sim_number = $1
		)
	`, number).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check beeline detalization snapshot: %w", err)
	}

	return exists, nil
}

func (r *Repository) UpdateDetalizationComputedBalance(ctx context.Context, number string, balance float64) error {
	number = domain.NormalizeSimNumber(number)
	balance = domain.RoundMoney(balance)

	result, err := r.db.ExecContext(ctx, `
		UPDATE beeline_detalization_snapshots
		SET computed_balance = $2,
			updated_at = NOW()
		WHERE sim_number = $1
	`, number, balance)
	if err != nil {
		return fmt.Errorf("update beeline detalization computed balance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update beeline detalization computed balance rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrDetalizationSnapshotNotFound
	}

	return nil
}

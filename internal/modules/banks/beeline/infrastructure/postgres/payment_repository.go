package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func (r *Repository) SumPaymentTotals(ctx context.Context, number string) (float64, error) {
	number = domain.NormalizeSimNumber(number)
	var total sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total), 0)
		FROM beeline_payments
		WHERE sim_number = $1 AND direction = 'outgoing'
	`, number).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("sum beeline payment totals: %w", err)
	}
	if !total.Valid {
		return 0, nil
	}

	return domain.RoundMoney(total.Float64), nil
}

func (r *Repository) SumIncomingTotals(ctx context.Context, number string) (float64, error) {
	number = domain.NormalizeSimNumber(number)
	var total sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM beeline_payments
		WHERE sim_number = $1 AND direction = 'incoming'
	`, number).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("sum beeline incoming totals: %w", err)
	}
	if !total.Valid {
		return 0, nil
	}

	return domain.RoundMoney(total.Float64), nil
}

func (r *Repository) CreatePayment(ctx context.Context, number string, payment domain.Payment) (domain.Payment, error) {
	number = domain.NormalizeSimNumber(number)
	created := payment
	created.SimNumber = number

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO beeline_payments (
			id, sim_number, direction, receiver_card, amount, commission, total, source, paid_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`,
		created.ID,
		created.SimNumber,
		string(created.Direction),
		created.ReceiverCard,
		created.Amount,
		created.Commission,
		created.Total,
		string(created.Source),
		created.PaidAt,
	).Scan(&created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("create beeline payment: %w", err)
	}

	return created, nil
}

func (r *Repository) ListPayments(ctx context.Context, number string) ([]domain.Payment, error) {
	number = domain.NormalizeSimNumber(number)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, sim_number, direction, receiver_card, amount, commission, total, source, paid_at, created_at, updated_at
		FROM beeline_payments
		WHERE sim_number = $1
		ORDER BY paid_at DESC, created_at DESC
	`, number)
	if err != nil {
		return nil, fmt.Errorf("list beeline payments: %w", err)
	}
	defer rows.Close()

	payments := make([]domain.Payment, 0)
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list beeline payments rows: %w", err)
	}

	return payments, nil
}

func (r *Repository) ListPaymentsInPeriod(ctx context.Context, number string, start, end time.Time) ([]domain.Payment, error) {
	number = domain.NormalizeSimNumber(number)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, sim_number, direction, receiver_card, amount, commission, total, source, paid_at, created_at, updated_at
		FROM beeline_payments
		WHERE sim_number = $1 AND paid_at >= $2 AND paid_at <= $3
		ORDER BY paid_at DESC, created_at DESC
	`, number, start, end)
	if err != nil {
		return nil, fmt.Errorf("list beeline payments in period: %w", err)
	}
	defer rows.Close()

	payments := make([]domain.Payment, 0)
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list beeline payments in period rows: %w", err)
	}

	return payments, nil
}

func (r *Repository) GetPayment(ctx context.Context, number, id string) (domain.Payment, error) {
	number = domain.NormalizeSimNumber(number)
	row := r.db.QueryRowContext(ctx, `
		SELECT id, sim_number, direction, receiver_card, amount, commission, total, source, paid_at, created_at, updated_at
		FROM beeline_payments
		WHERE id = $1 AND sim_number = $2
	`, id, number)

	payment, err := scanPayment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, domain.ErrPaymentNotFound
	}
	if err != nil {
		return domain.Payment{}, fmt.Errorf("get beeline payment: %w", err)
	}

	return payment, nil
}

func (r *Repository) UpdatePayment(ctx context.Context, number string, payment domain.Payment) (domain.Payment, error) {
	number = domain.NormalizeSimNumber(number)
	updated := payment
	updated.SimNumber = number

	err := r.db.QueryRowContext(ctx, `
		UPDATE beeline_payments
		SET direction = $3,
			receiver_card = $4,
			amount = $5,
			commission = $6,
			total = $7,
			paid_at = $8,
			updated_at = NOW()
		WHERE id = $1 AND sim_number = $2
		RETURNING source, created_at, updated_at
	`,
		updated.ID,
		updated.SimNumber,
		string(updated.Direction),
		updated.ReceiverCard,
		updated.Amount,
		updated.Commission,
		updated.Total,
		updated.PaidAt,
	).Scan(&updated.Source, &updated.CreatedAt, &updated.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, domain.ErrPaymentNotFound
	}
	if err != nil {
		return domain.Payment{}, fmt.Errorf("update beeline payment: %w", err)
	}

	return updated, nil
}

func (r *Repository) DeletePayment(ctx context.Context, number, id string) error {
	number = domain.NormalizeSimNumber(number)
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM beeline_payments
		WHERE id = $1 AND sim_number = $2
	`, id, number)
	if err != nil {
		return fmt.Errorf("delete beeline payment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete beeline payment rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrPaymentNotFound
	}

	return nil
}

type paymentRowScanner interface {
	Scan(dest ...any) error
}

func scanPayment(scanner paymentRowScanner) (domain.Payment, error) {
	payment := domain.Payment{}
	var source string
	var direction string

	err := scanner.Scan(
		&payment.ID,
		&payment.SimNumber,
		&direction,
		&payment.ReceiverCard,
		&payment.Amount,
		&payment.Commission,
		&payment.Total,
		&source,
		&payment.PaidAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return domain.Payment{}, err
	}

	payment.Source = domain.PaymentSource(source)
	payment.Direction = domain.PaymentDirection(direction)
	if payment.Direction == "" {
		payment.Direction = domain.PaymentDirectionOutgoing
	}

	return payment, nil
}

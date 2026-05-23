package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"project/internal/modules/smsagent/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Enqueue(ctx context.Context, message domain.OutboundMessage) (domain.OutboundMessage, error) {
	created := message

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO sms_agent_messages (id, address, body, bank, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`,
		created.ID,
		created.Address,
		created.Body,
		created.Bank,
		string(created.Status),
	).Scan(&created.CreatedAt)
	if err != nil {
		return domain.OutboundMessage{}, fmt.Errorf("enqueue sms agent message: %w", err)
	}

	return created, nil
}

func (r *Repository) ListPending(ctx context.Context, limit int) ([]domain.OutboundMessage, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, address, body, bank, status, device_id, error_message, created_at, delivered_at
		FROM sms_agent_messages
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending sms agent messages: %w", err)
	}
	defer rows.Close()

	messages := make([]domain.OutboundMessage, 0)
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list pending sms agent messages rows: %w", err)
	}

	return messages, nil
}

func (r *Repository) Ack(ctx context.Context, id string, status domain.MessageStatus, deviceID, errorMessage string) error {
	var deviceArg any
	if deviceID != "" {
		deviceArg = deviceID
	}

	var errorArg any
	if errorMessage != "" {
		errorArg = errorMessage
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE sms_agent_messages
		SET status = $2,
			device_id = COALESCE($3, device_id),
			error_message = COALESCE($4, error_message),
			delivered_at = CASE WHEN $2 = 'delivered' THEN NOW() ELSE delivered_at END
		WHERE id = $1 AND status = 'pending'
	`, id, string(status), deviceArg, errorArg)
	if err != nil {
		return fmt.Errorf("ack sms agent message: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ack sms agent message rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrMessageNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMessage(scanner rowScanner) (domain.OutboundMessage, error) {
	message := domain.OutboundMessage{}
	var status string
	var deviceID sql.NullString
	var errorMessage sql.NullString
	var deliveredAt sql.NullTime

	err := scanner.Scan(
		&message.ID,
		&message.Address,
		&message.Body,
		&message.Bank,
		&status,
		&deviceID,
		&errorMessage,
		&message.CreatedAt,
		&deliveredAt,
	)
	if err != nil {
		return domain.OutboundMessage{}, err
	}

	message.Status = domain.MessageStatus(status)
	if deviceID.Valid {
		value := deviceID.String
		message.DeviceID = &value
	}
	if errorMessage.Valid {
		value := errorMessage.String
		message.ErrorMessage = &value
	}
	if deliveredAt.Valid {
		value := deliveredAt.Time
		message.DeliveredAt = &value
	}

	return message, nil
}

var _ domain.Repository = (*Repository)(nil)

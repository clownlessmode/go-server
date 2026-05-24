package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"project/internal/modules/banks/rocketbank/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetConfig(ctx context.Context) (*domain.Config, error) {
	config := &domain.Config{}
	var history []byte
	var hiddenHistoryIDs []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT balance, first_name, middle_name, last_name, phone_number, card_number, history, hidden_history_ids, created_at, updated_at
		FROM rocketbank_configs
		WHERE id = 1
	`).Scan(
		&config.Balance,
		&config.ClientInfo.FirstName,
		&config.ClientInfo.MiddleName,
		&config.ClientInfo.LastName,
		&config.ClientInfo.PhoneNumber,
		&config.ClientInfo.CardNumber,
		&history,
		&hiddenHistoryIDs,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &domain.Config{
			Balance: nil,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get rocketbank config: %w", err)
	}
	if err := scanHistory(history, &config.History); err != nil {
		return nil, err
	}
	if err := scanHiddenHistoryIDs(hiddenHistoryIDs, &config.HiddenHistoryIDs); err != nil {
		return nil, err
	}

	return config, nil
}

func (r *Repository) UpdateBalance(ctx context.Context, balance *float64) (*domain.Config, error) {
	config := &domain.Config{}
	var history []byte
	var hiddenHistoryIDs []byte

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO rocketbank_configs (id, balance)
		VALUES (1, $1)
		ON CONFLICT (id) DO UPDATE
		SET balance = EXCLUDED.balance,
			updated_at = NOW()
		RETURNING balance, first_name, middle_name, last_name, phone_number, card_number, history, hidden_history_ids, created_at, updated_at
	`, balance).Scan(
		&config.Balance,
		&config.ClientInfo.FirstName,
		&config.ClientInfo.MiddleName,
		&config.ClientInfo.LastName,
		&config.ClientInfo.PhoneNumber,
		&config.ClientInfo.CardNumber,
		&history,
		&hiddenHistoryIDs,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update rocketbank balance: %w", err)
	}
	if err := scanHistory(history, &config.History); err != nil {
		return nil, err
	}
	if err := scanHiddenHistoryIDs(hiddenHistoryIDs, &config.HiddenHistoryIDs); err != nil {
		return nil, err
	}

	return config, nil
}

func (r *Repository) UpdateClientInfo(ctx context.Context, clientInfo domain.ClientInfo) (*domain.Config, error) {
	config := &domain.Config{}
	var history []byte
	var hiddenHistoryIDs []byte

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO rocketbank_configs (id, first_name, middle_name, last_name, phone_number, card_number)
		VALUES (1, $1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET first_name = EXCLUDED.first_name,
			middle_name = EXCLUDED.middle_name,
			last_name = EXCLUDED.last_name,
			phone_number = EXCLUDED.phone_number,
			card_number = EXCLUDED.card_number,
			updated_at = NOW()
		RETURNING balance, first_name, middle_name, last_name, phone_number, card_number, history, hidden_history_ids, created_at, updated_at
	`, clientInfo.FirstName, clientInfo.MiddleName, clientInfo.LastName, clientInfo.PhoneNumber, clientInfo.CardNumber).Scan(
		&config.Balance,
		&config.ClientInfo.FirstName,
		&config.ClientInfo.MiddleName,
		&config.ClientInfo.LastName,
		&config.ClientInfo.PhoneNumber,
		&config.ClientInfo.CardNumber,
		&history,
		&hiddenHistoryIDs,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update rocketbank client info: %w", err)
	}
	if err := scanHistory(history, &config.History); err != nil {
		return nil, err
	}
	if err := scanHiddenHistoryIDs(hiddenHistoryIDs, &config.HiddenHistoryIDs); err != nil {
		return nil, err
	}

	return config, nil
}

func (r *Repository) ListHistory(ctx context.Context) ([]domain.HistoryItem, error) {
	history, err := r.getHistory(ctx)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func (r *Repository) GetHistoryItem(ctx context.Context, id string) (domain.HistoryItem, error) {
	history, err := r.getHistory(ctx)
	if err != nil {
		return domain.HistoryItem{}, err
	}

	for _, item := range history {
		if domain.HistoryItemID(item) == id {
			return item, nil
		}
	}

	return domain.HistoryItem{}, domain.ErrHistoryItemNotFound
}

func (r *Repository) CreateHistoryItem(ctx context.Context, item domain.HistoryItem) (domain.HistoryItem, error) {
	id := domain.HistoryItemID(item)
	if id == "" {
		return domain.HistoryItem{}, fmt.Errorf("create rocketbank history item: missing transaction id")
	}

	history, err := r.getHistory(ctx)
	if err != nil {
		return domain.HistoryItem{}, err
	}

	for _, existing := range history {
		if domain.HistoryItemID(existing) == id {
			return domain.HistoryItem{}, domain.ErrHistoryItemExists
		}
	}

	history = append([]domain.HistoryItem{item}, history...)
	if err := r.saveHistory(ctx, history); err != nil {
		return domain.HistoryItem{}, err
	}

	return item, nil
}

func (r *Repository) UpdateHistoryItem(ctx context.Context, id string, item domain.HistoryItem) (domain.HistoryItem, error) {
	newID := domain.HistoryItemID(item)
	if newID == "" {
		return domain.HistoryItem{}, fmt.Errorf("update rocketbank history item: missing transaction id")
	}

	history, err := r.getHistory(ctx)
	if err != nil {
		return domain.HistoryItem{}, err
	}

	for _, existing := range history {
		existingID := domain.HistoryItemID(existing)
		if existingID != "" && existingID != id && existingID == newID {
			return domain.HistoryItem{}, domain.ErrHistoryItemExists
		}
	}

	for index, existing := range history {
		if domain.HistoryItemID(existing) == id {
			history[index] = item
			if err := r.saveHistory(ctx, history); err != nil {
				return domain.HistoryItem{}, err
			}

			return item, nil
		}
	}

	return domain.HistoryItem{}, domain.ErrHistoryItemNotFound
}

func (r *Repository) DeleteHistoryItem(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("delete rocketbank history item: empty transaction id")
	}

	history, err := r.getHistory(ctx)
	if err != nil {
		return err
	}

	hiddenHistoryIDs, err := r.getHiddenHistoryIDs(ctx)
	if err != nil {
		return err
	}

	for index, item := range history {
		if domain.HistoryItemID(item) == id {
			history = append(history[:index], history[index+1:]...)
			if err := r.saveHistory(ctx, history); err != nil {
				return err
			}
			break
		}
	}

	if domain.IsHiddenHistoryID(hiddenHistoryIDs, id) {
		return nil
	}

	hiddenHistoryIDs = append(hiddenHistoryIDs, id)
	return r.saveHiddenHistoryIDs(ctx, hiddenHistoryIDs)
}

func (r *Repository) ClearHistory(ctx context.Context) error {
	if err := r.saveHistory(ctx, []domain.HistoryItem{}); err != nil {
		return err
	}

	return r.saveHiddenHistoryIDs(ctx, []string{})
}

func (r *Repository) getHistory(ctx context.Context) ([]domain.HistoryItem, error) {
	var raw []byte
	if err := r.db.QueryRowContext(ctx, `
		SELECT history
		FROM rocketbank_configs
		WHERE id = 1
	`).Scan(&raw); err != nil {
		return nil, fmt.Errorf("get rocketbank history: %w", err)
	}

	var history []domain.HistoryItem
	if err := scanHistory(raw, &history); err != nil {
		return nil, err
	}

	return history, nil
}

func (r *Repository) saveHistory(ctx context.Context, history []domain.HistoryItem) error {
	raw, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("marshal rocketbank history: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `
		UPDATE rocketbank_configs
		SET history = $1::jsonb,
			updated_at = NOW()
		WHERE id = 1
	`, raw); err != nil {
		return fmt.Errorf("save rocketbank history: %w", err)
	}

	return nil
}

func (r *Repository) saveHiddenHistoryIDs(ctx context.Context, hiddenHistoryIDs []string) error {
	raw, err := json.Marshal(hiddenHistoryIDs)
	if err != nil {
		return fmt.Errorf("marshal rocketbank hidden history ids: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `
		UPDATE rocketbank_configs
		SET hidden_history_ids = $1::jsonb,
			updated_at = NOW()
		WHERE id = 1
	`, raw); err != nil {
		return fmt.Errorf("save rocketbank hidden history ids: %w", err)
	}

	return nil
}

func (r *Repository) getHiddenHistoryIDs(ctx context.Context) ([]string, error) {
	var raw []byte
	if err := r.db.QueryRowContext(ctx, `
		SELECT hidden_history_ids
		FROM rocketbank_configs
		WHERE id = 1
	`).Scan(&raw); err != nil {
		return nil, fmt.Errorf("get rocketbank hidden history ids: %w", err)
	}

	var hiddenHistoryIDs []string
	if err := scanHiddenHistoryIDs(raw, &hiddenHistoryIDs); err != nil {
		return nil, err
	}

	return hiddenHistoryIDs, nil
}

func scanHiddenHistoryIDs(raw []byte, hiddenHistoryIDs *[]string) error {
	if len(raw) == 0 {
		*hiddenHistoryIDs = []string{}
		return nil
	}

	if err := json.Unmarshal(raw, hiddenHistoryIDs); err != nil {
		return fmt.Errorf("scan rocketbank hidden history ids: %w", err)
	}

	return nil
}

func scanHistory(raw []byte, history *[]domain.HistoryItem) error {
	if len(raw) == 0 {
		*history = []domain.HistoryItem{}
		return nil
	}

	var rawItems []json.RawMessage
	if err := json.Unmarshal(raw, &rawItems); err != nil {
		return fmt.Errorf("scan rocketbank history: %w", err)
	}

	parsed := make([]domain.HistoryItem, 0, len(rawItems))
	for _, rawItem := range rawItems {
		var item domain.HistoryItem
		if err := json.Unmarshal(rawItem, &item); err == nil && item.Type != "" {
			parsed = append(parsed, item)
			continue
		}

		var legacy map[string]any
		if err := json.Unmarshal(rawItem, &legacy); err == nil {
			if item, ok := domain.HistoryItemFromLegacyOperation(legacy); ok {
				parsed = append(parsed, item)
			}
		}
	}

	*history = parsed
	return nil
}

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"project/internal/modules/banks/beeline/detalization"
	"project/internal/modules/banks/beeline/domain"
)

func (r *Repository) ListHiddenTransactionIDs(ctx context.Context, number string) ([]string, error) {
	number = domain.NormalizeSimNumber(number)
	var raw []byte
	if err := r.db.QueryRowContext(ctx, `
		SELECT hidden_transaction_ids
		FROM beeline_sims
		WHERE number = $1
	`, number).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSimNotFound
		}
		return nil, fmt.Errorf("list beeline hidden transactions: %w", err)
	}

	return scanHiddenTransactionIDs(raw)
}

func (r *Repository) HideTransaction(ctx context.Context, number, id string) error {
	number = domain.NormalizeSimNumber(number)
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("hide beeline transaction: empty id")
	}

	hiddenIDs, err := r.ListHiddenTransactionIDs(ctx, number)
	if err != nil {
		return err
	}

	keys := []string{id}
	if snapshot, err := r.GetDetalizationSnapshot(ctx, number); err == nil {
		if data, err := detalization.DecodeSnapshotData(snapshot.Data); err == nil {
			if tx, ok := detalization.FindTransactionByID(data, id); ok {
				keys = detalization.HiddenKeysForTransaction(tx)
			}
		}
	}

	changed := false
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || domain.IsHiddenTransactionID(hiddenIDs, key) {
			continue
		}
		hiddenIDs = append(hiddenIDs, key)
		changed = true
	}

	if changed {
		raw, err := json.Marshal(hiddenIDs)
		if err != nil {
			return fmt.Errorf("marshal beeline hidden transactions: %w", err)
		}

		if _, err := r.db.ExecContext(ctx, `
			UPDATE beeline_sims
			SET hidden_transaction_ids = $2::jsonb,
				updated_at = NOW()
			WHERE number = $1
		`, number, raw); err != nil {
			return fmt.Errorf("hide beeline transaction: %w", err)
		}
	}

	return r.CompactDetalizationSnapshot(ctx, number)
}

func (r *Repository) CompactDetalizationSnapshot(ctx context.Context, number string) error {
	// Hidden transactions stay in the snapshot so their amounts can adjust display balance.
	return nil
}

func scanHiddenTransactionIDs(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}

	var hiddenIDs []string
	if err := json.Unmarshal(raw, &hiddenIDs); err != nil {
		return nil, fmt.Errorf("scan beeline hidden transactions: %w", err)
	}

	return hiddenIDs, nil
}

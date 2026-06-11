package detalization

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func TransactionFingerprint(tx map[string]any) string {
	return fmt.Sprintf(
		"%s|%s|%s|%.2f",
		jsonString(tx["dateTime"]),
		jsonString(tx["category"]),
		strings.TrimSpace(jsonString(tx["name"])),
		domain.RoundMoney(transactionChangeValue(tx)),
	)
}

func HiddenKeysForTransaction(tx map[string]any) []string {
	keys := []string{TransactionID(tx)}
	fingerprint := TransactionFingerprint(tx)
	if fingerprint != "|||0.00" {
		keys = append(keys, fingerprint)
	}

	return keys
}

func IsTransactionHidden(tx map[string]any, hiddenIDs []string) bool {
	if domain.IsHiddenTransactionID(hiddenIDs, TransactionID(tx)) {
		return true
	}

	return domain.IsHiddenTransactionID(hiddenIDs, TransactionFingerprint(tx))
}

func FindTransactionByID(data map[string]any, id string) (map[string]any, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, false
	}

	transactions, ok := data["transactions"].([]any)
	if !ok {
		return nil, false
	}

	for _, item := range transactions {
		tx, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if existingID, ok := tx["id"].(string); ok && strings.TrimSpace(existingID) == id {
			return tx, true
		}
		if TransactionID(tx) == id {
			return tx, true
		}
	}

	return nil, false
}

func HiddenTransactionsNetChange(data map[string]any, hiddenIDs []string) float64 {
	if len(hiddenIDs) == 0 {
		return 0
	}

	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) == 0 {
		return 0
	}

	seen := make(map[string]struct{})
	var total float64
	for _, item := range transactions {
		tx, ok := item.(map[string]any)
		if !ok || !IsTransactionHidden(tx, hiddenIDs) {
			continue
		}

		id := TransactionID(tx)
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		total += transactionChangeValue(tx)
	}

	return domain.RoundMoney(total)
}

func PurgeHiddenFromData(data map[string]any, hiddenIDs []string) (map[string]any, *float64, error) {
	working, err := CloneData(data)
	if err != nil {
		return nil, nil, err
	}

	FilterHiddenTransactions(working, hiddenIDs)

	balance, ok := recalculateBalances(working, nil)
	if !ok {
		return working, nil, fmt.Errorf("recalculate detalization balances")
	}

	value := domain.RoundMoney(balance)
	return working, &value, nil
}

func TransactionID(tx map[string]any) string {
	if id, ok := tx["id"].(string); ok {
		id = strings.TrimSpace(id)
		if id != "" {
			return id
		}
	}

	seed := fmt.Sprintf(
		"%s|%s|%s|%s|%.2f|%.0f",
		jsonString(tx["dateTime"]),
		jsonString(tx["category"]),
		jsonString(tx["name"]),
		jsonString(tx["typeCall"]),
		transactionChangeValue(tx),
		jsonNumber(tx["volume"]),
	)
	sum := sha256.Sum256([]byte(seed))

	return hex.EncodeToString(sum[:16])
}

func IsHiddenTransactionID(hiddenIDs []string, id string) bool {
	return domain.IsHiddenTransactionID(hiddenIDs, id)
}

func FilterHiddenTransactions(data map[string]any, hiddenIDs []string) bool {
	if len(hiddenIDs) == 0 {
		return false
	}

	hiddenSet := make(map[string]struct{}, len(hiddenIDs))
	for _, hiddenID := range hiddenIDs {
		hiddenID = strings.TrimSpace(hiddenID)
		if hiddenID == "" {
			continue
		}
		hiddenSet[hiddenID] = struct{}{}
	}
	if len(hiddenSet) == 0 {
		return false
	}

	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) == 0 {
		return false
	}

	filtered := make([]any, 0, len(transactions))
	changed := false
	for _, item := range transactions {
		tx, ok := item.(map[string]any)
		if !ok {
			filtered = append(filtered, item)
			continue
		}
		if IsTransactionHidden(tx, hiddenIDs) {
			changed = true
			continue
		}
		filtered = append(filtered, item)
	}

	if !changed {
		return false
	}

	data["transactions"] = filtered
	_, _ = recalculateBalances(data, nil)

	return true
}

func AnnotateTransactionIDs(data map[string]any, payments []domain.Payment) {
	paymentByFingerprint := make(map[string]string, len(payments))
	paymentIDSet := make(map[string]struct{}, len(payments))
	for _, payment := range payments {
		paymentByFingerprint[paymentFingerprint(payment)] = payment.ID
		paymentIDSet[payment.ID] = struct{}{}
	}

	transactions, ok := data["transactions"].([]any)
	if !ok {
		return
	}

	for _, item := range transactions {
		tx, ok := item.(map[string]any)
		if !ok {
			continue
		}

		id := strings.TrimSpace(jsonString(tx["id"]))
		if id != "" {
			if _, isPayment := paymentIDSet[id]; isPayment {
				tx["source"] = "payment"
				tx["hideable"] = false
				continue
			}
			if strings.EqualFold(jsonString(tx["source"]), "payment") {
				tx["hideable"] = false
				continue
			}
			if isSyntheticPaddingSMS(tx) {
				continue
			}
			if _, hasSource := tx["source"]; !hasSource {
				tx["source"] = "beeline"
			}
			tx["hideable"] = true
			continue
		}

		if paymentID, ok := paymentByFingerprint[transactionFingerprint(tx)]; ok {
			tx["id"] = paymentID
			tx["source"] = "payment"
			tx["hideable"] = false
			continue
		}

		tx["id"] = TransactionID(tx)
		tx["source"] = "beeline"
		tx["hideable"] = true
	}
}

func BuildView(baseData map[string]any, payments []domain.Payment, hiddenIDs []string, configuredOpening *float64, simNumber string, now time.Time) (map[string]any, float64, error) {
	working, err := CloneData(baseData)
	if err != nil {
		return nil, 0, err
	}

	FilterHiddenTransactions(working, hiddenIDs)

	balance, ok := ApplyPayments(working, payments, configuredOpening)
	if !ok {
		return nil, 0, fmt.Errorf("build beeline detalization view")
	}

	PadTodayOutgoingSMS(working, simNumber, now)

	balance, ok = recalculateBalances(working, configuredOpening)
	if !ok {
		return nil, 0, fmt.Errorf("build beeline detalization view")
	}

	AnnotateTransactionIDs(working, payments)
	syncBalanceReturnPaymentLabels(working, payments)

	return working, balance, nil
}

func paymentFingerprint(payment domain.Payment) string {
	return TransactionFingerprint(paymentTransaction(payment))
}

func syncBalanceReturnPaymentLabels(data map[string]any, payments []domain.Payment) {
	for _, payment := range payments {
		if payment.Direction != domain.PaymentDirectionBalanceReturn {
			continue
		}

		tx, ok := FindTransactionByID(data, payment.ID)
		if !ok {
			continue
		}

		applyPaymentTransactionLabels(tx, paymentTransaction(payment))
		tx["source"] = "payment"
	}
}

func transactionFingerprint(tx map[string]any) string {
	return TransactionFingerprint(tx)
}

func BalanceReturnTransactionSummaries(data map[string]any) []string {
	transactions, ok := data["transactions"].([]any)
	if !ok {
		return nil
	}

	out := make([]string, 0)
	for _, item := range transactions {
		tx, ok := item.(map[string]any)
		if !ok || !isBalanceReturnTransaction(tx) {
			continue
		}

		out = append(out, fmt.Sprintf(
			"id=%s name=%q categoryName=%q formattedNumber=%q source=%s",
			jsonString(tx["id"]),
			jsonString(tx["name"]),
			jsonString(tx["categoryName"]),
			jsonString(tx["formattedNumber"]),
			jsonString(tx["source"]),
		))
	}

	return out
}

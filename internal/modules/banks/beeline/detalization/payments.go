package detalization

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func CloneData(data map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal detalization data: %w", err)
	}

	cloned := make(map[string]any)
	if err := json.Unmarshal(raw, &cloned); err != nil {
		return nil, fmt.Errorf("unmarshal detalization data: %w", err)
	}

	return cloned, nil
}

func DecodeSnapshotData(raw []byte) (map[string]any, error) {
	data := make(map[string]any)
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("unmarshal detalization snapshot: %w", err)
	}

	return data, nil
}

func PaymentTotals(payments []domain.Payment) (outgoingTotal, incomingTotal float64) {
	for _, payment := range payments {
		switch payment.Direction {
		case domain.PaymentDirectionIncoming:
			incomingTotal += payment.Amount
		default:
			outgoingTotal += payment.Total
		}
	}

	return domain.RoundMoney(outgoingTotal), domain.RoundMoney(incomingTotal)
}

func ApplyPayments(data map[string]any, payments []domain.Payment) (float64, bool) {
	existing, _ := data["transactions"].([]any)
	injected := make([]map[string]any, 0, len(payments))

	var outgoingSum float64
	var incomingSum float64

	for _, payment := range payments {
		tx := paymentTransaction(payment)
		if transactionExists(existing, tx) {
			continue
		}
		injected = append(injected, tx)

		switch payment.Direction {
		case domain.PaymentDirectionIncoming:
			incomingSum += payment.Amount
		default:
			outgoingSum += payment.Total
		}
	}

	merged := existing
	if len(injected) > 0 {
		merged = append(injectedAsAny(injected), existing...)
		sort.SliceStable(merged, func(i, j int) bool {
			return transactionDateTime(merged[i]) > transactionDateTime(merged[j])
		})
		data["transactions"] = merged
	}

	if outgoingSum > 0 || incomingSum > 0 {
		updateCategories(data, outgoingSum, incomingSum)
		updateSummaryAmounts(data, outgoingSum, incomingSum)
	}

	return recalculateBalances(data)
}

func paymentTransaction(payment domain.Payment) map[string]any {
	dateTime := paymentDateTime(payment.PaidAt)

	if payment.Source == domain.PaymentSourcePaymentFlowSMS {
		return paymentFlowSMSTransaction(payment.ID, dateTime)
	}

	if payment.Direction == domain.PaymentDirectionIncoming {
		return map[string]any{
			"id": payment.ID,
			"balances": []any{
				map[string]any{
					"changeValue": payment.Amount,
					"code":        "coreBalance",
					"name":        "личный баланс",
					"unit":        "RUB",
				},
			},
			"category":        "refill",
			"categoryName":    "пополнение баланса",
			"dateTime":        dateTime,
			"formattedNumber": "",
			"icon":            "download",
			"name":            "пополнение баланса",
			"number":          "",
			"roaming":         false,
			"typeCall":        "recharge",
			"unit":            "",
			"volume":          0,
		}
	}

	return map[string]any{
		"id": payment.ID,
		"balances": []any{
			map[string]any{
				"changeValue": -payment.Total,
				"code":        "coreBalance",
				"name":        "личный баланс",
				"unit":        "RUB",
			},
		},
		"category":        "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
		"categoryName":    "платежи и переводы",
		"dateTime":        dateTime,
		"formattedNumber": "",
		"name":            "списание за мобильную коммерцию",
		"number":          "",
		"roaming":         false,
		"typeCall":        "incomingCall",
		"unit":            "",
		"volume":          0,
	}
}

func paymentFlowSMSTransaction(id, dateTime string) map[string]any {
	return map[string]any{
		"id": id,
		"balances": []any{
			map[string]any{
				"changeValue": 0,
				"code":        "coreBalance",
				"name":        "личный баланс",
				"unit":        "RUB",
			},
		},
		"category":        "SMS_MMS",
		"categoryName":    "сообщения",
		"dateTime":        dateTime,
		"formattedNumber": domain.PaymentFlowSMSNumber,
		"icon":            "smsMms",
		"name":            "исходящее SMS",
		"number":          domain.PaymentFlowSMSNumber,
		"roaming":         false,
		"typeCall":        "outgoingCall",
		"unit":            "PIECE",
		"volume":          1,
	}
}

func paymentDateTime(paidAt time.Time) string {
	return paidAt.Format("2006-01-02T15:04:05")
}

func transactionExists(existing []any, candidate map[string]any) bool {
	candidateDate := transactionDateTime(candidate)
	candidateChange := transactionChangeValue(candidate)

	for _, item := range existing {
		tx, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if tx["dateTime"] != candidateDate {
			continue
		}
		if transactionChangeValue(tx) == candidateChange {
			return true
		}
	}

	return false
}

func recalculateBalances(data map[string]any) (float64, bool) {
	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) == 0 {
		return 0, false
	}

	summary, ok := coreBalanceSummary(data)
	if !ok {
		return 0, false
	}

	sorted := append([]any(nil), transactions...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := transactionDateTime(sorted[i])
		right := transactionDateTime(sorted[j])
		if left == right {
			return false
		}
		return left < right
	})

	running := findOpeningBalance(sorted)
	summary["startValue"] = running
	periodStartValue := running

	for _, item := range sorted {
		tx, ok := item.(map[string]any)
		if !ok {
			continue
		}

		balance, ok := coreBalanceEntry(tx)
		if !ok {
			continue
		}

		change := jsonNumber(balance["changeValue"])
		balance["startValue"] = running
		if change != 0 {
			running = domain.RoundMoney(running + change)
			balance["changeValue"] = change
			balance["endValue"] = running
			continue
		}

		balance["changeValue"] = 0
		balance["endValue"] = running
	}

	summary["endValue"] = running
	if periodStartValue != 0 || running != 0 {
		summary["changeValue"] = domain.RoundMoney(running - periodStartValue)
	}

	return running, true
}

func findOpeningBalance(transactions []any) float64 {
	for _, item := range transactions {
		balance, ok := coreBalanceEntry(item)
		if !ok {
			continue
		}

		change := jsonNumber(balance["changeValue"])
		start := jsonNumber(balance["startValue"])
		if change == 0 && start != 0 {
			return domain.RoundMoney(start)
		}
	}

	for _, item := range transactions {
		balance, ok := coreBalanceEntry(item)
		if !ok {
			continue
		}

		change := jsonNumber(balance["changeValue"])
		if change == 0 {
			continue
		}

		if start := jsonNumber(balance["startValue"]); start != 0 {
			return domain.RoundMoney(start)
		}
	}

	return 0
}

func coreBalanceSummary(data map[string]any) (map[string]any, bool) {
	balances, ok := data["balances"].([]any)
	if !ok || len(balances) == 0 {
		return nil, false
	}

	summary, ok := balances[0].(map[string]any)
	if !ok {
		return nil, false
	}

	return summary, true
}

func coreBalanceEntry(item any) (map[string]any, bool) {
	tx, ok := item.(map[string]any)
	if !ok {
		return nil, false
	}

	balances, ok := tx["balances"].([]any)
	if !ok || len(balances) == 0 {
		return nil, false
	}

	balance, ok := balances[0].(map[string]any)
	if !ok {
		return nil, false
	}

	if code := jsonString(balance["code"]); code != "" && code != "coreBalance" {
		return nil, false
	}

	return balance, true
}

func transactionDateTime(item any) string {
	tx, ok := item.(map[string]any)
	if !ok {
		return ""
	}

	value, _ := tx["dateTime"].(string)
	return value
}

func transactionChangeValue(item any) float64 {
	balance, ok := coreBalanceEntry(item)
	if !ok {
		return 0
	}

	return jsonNumber(balance["changeValue"])
}

func jsonNumber(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}

func injectedAsAny(items []map[string]any) []any {
	result := make([]any, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}

	return result
}

func updateSummaryAmounts(data map[string]any, outgoingSum, incomingSum float64) {
	summary, ok := coreBalanceSummary(data)
	if !ok {
		return
	}

	if outgoingSum > 0 {
		summary["spentValue"] = jsonNumber(summary["spentValue"]) + outgoingSum
	}
	if incomingSum > 0 {
		summary["paidValue"] = jsonNumber(summary["paidValue"]) + incomingSum
	}
}

func updateCategories(data map[string]any, outgoingSum, incomingSum float64) {
	categories, ok := data["categories"].([]any)
	if !ok {
		categories = make([]any, 0)
	}

	if outgoingSum > 0 {
		categories = upsertCategory(
			categories,
			"SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
			"платежи и переводы",
			-outgoingSum,
			[]string{"operationExpenses"},
		)
	}
	if incomingSum > 0 {
		categories = upsertCategory(
			categories,
			"REFILL",
			"пополнение баланса",
			incomingSum,
			[]string{"operationRefill"},
		)
	}

	data["categories"] = categories
}

func upsertCategory(
	categories []any,
	id, name string,
	charge float64,
	screen []string,
) []any {
	for _, item := range categories {
		category, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if strings.EqualFold(jsonString(category["id"]), id) {
			category["totalMonetaryCharge"] = jsonNumber(category["totalMonetaryCharge"]) + charge
			return categories
		}
	}

	return append(categories, map[string]any{
		"id":                  id,
		"name":                name,
		"screen":              screen,
		"subCategory":         []any{},
		"totalMonetaryCharge": charge,
		"totalTrafficCharge":  0,
		"unit":                "",
	})
}

func jsonString(value any) string {
	text, _ := value.(string)
	return text
}

package proxy

import (
	"sort"
	"strings"
	"time"

	beelinedomain "project/internal/modules/banks/beeline/domain"
)

func applyBeelineDetalizationPayments(data map[string]any, payments []beelinedomain.Payment) (float64, bool) {
	existing, _ := data["transactions"].([]any)
	injected := make([]map[string]any, 0, len(payments))

	var outgoingSum float64
	var incomingSum float64

	for _, payment := range payments {
		tx := beelineDetalizationTransaction(payment)
		if detalizationTransactionExists(existing, tx) {
			continue
		}
		injected = append(injected, tx)

		switch payment.Direction {
		case beelinedomain.PaymentDirectionIncoming:
			incomingSum += payment.Amount
		default:
			outgoingSum += payment.Total
		}
	}

	merged := existing
	if len(injected) > 0 {
		merged = append(injectedAsAny(injected), existing...)
		sort.SliceStable(merged, func(i, j int) bool {
			return detalizationTransactionDateTime(merged[i]) > detalizationTransactionDateTime(merged[j])
		})
		data["transactions"] = merged
	}

	if outgoingSum > 0 || incomingSum > 0 {
		updateDetalizationCategories(data, outgoingSum, incomingSum)
		updateDetalizationSummaryAmounts(data, outgoingSum, incomingSum)
	}

	finalBalance, ok := recalculateDetalizationBalances(data)
	if !ok {
		return 0, false
	}

	return finalBalance, true
}

func beelineDetalizationTransaction(payment beelinedomain.Payment) map[string]any {
	dateTime := beelineDetalizationDateTime(payment.PaidAt)

	if payment.Direction == beelinedomain.PaymentDirectionIncoming {
		return map[string]any{
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

func beelineDetalizationDateTime(paidAt time.Time) string {
	return paidAt.Format("2006-01-02T15:04:05")
}

func detalizationTransactionExists(existing []any, candidate map[string]any) bool {
	candidateDate := detalizationTransactionDateTime(candidate)
	candidateChange := detalizationTransactionChangeValue(candidate)

	for _, item := range existing {
		tx, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if tx["dateTime"] != candidateDate {
			continue
		}
		if detalizationTransactionChangeValue(tx) == candidateChange {
			return true
		}
	}

	return false
}

func recalculateDetalizationBalances(data map[string]any) (float64, bool) {
	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) == 0 {
		return 0, false
	}

	summary, ok := detalizationCoreBalanceSummary(data)
	if !ok {
		return 0, false
	}

	sorted := append([]any(nil), transactions...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := detalizationTransactionDateTime(sorted[i])
		right := detalizationTransactionDateTime(sorted[j])
		if left == right {
			return false
		}
		return left < right
	})

	running := findDetalizationOpeningBalance(sorted)
	periodStartValue := jsonNumber(summary["startValue"])

	for _, item := range sorted {
		tx, ok := item.(map[string]any)
		if !ok {
			continue
		}

		balance, ok := detalizationCoreBalanceEntry(tx)
		if !ok {
			continue
		}

		change := jsonNumber(balance["changeValue"])
		balance["startValue"] = running
		if change != 0 {
			running = beelinedomain.RoundMoney(running + change)
			balance["changeValue"] = change
			balance["endValue"] = running
			continue
		}

		balance["changeValue"] = 0
		balance["endValue"] = running
	}

	summary["endValue"] = running
	if periodStartValue != 0 || running != 0 {
		summary["changeValue"] = beelinedomain.RoundMoney(running - periodStartValue)
	}

	return running, true
}

func findDetalizationOpeningBalance(transactions []any) float64 {
	for _, item := range transactions {
		balance, ok := detalizationCoreBalanceEntry(item)
		if !ok {
			continue
		}

		change := jsonNumber(balance["changeValue"])
		if change == 0 {
			continue
		}

		return jsonNumber(balance["startValue"])
	}

	return 0
}

func detalizationCoreBalanceSummary(data map[string]any) (map[string]any, bool) {
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

func detalizationCoreBalanceEntry(item any) (map[string]any, bool) {
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

func detalizationTransactionDateTime(item any) string {
	tx, ok := item.(map[string]any)
	if !ok {
		return ""
	}

	value, _ := tx["dateTime"].(string)
	return value
}

func detalizationTransactionChangeValue(item any) float64 {
	balance, ok := detalizationCoreBalanceEntry(item)
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

func updateDetalizationSummaryAmounts(data map[string]any, outgoingSum, incomingSum float64) {
	summary, ok := detalizationCoreBalanceSummary(data)
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

func updateDetalizationCategories(data map[string]any, outgoingSum, incomingSum float64) {
	categories, ok := data["categories"].([]any)
	if !ok {
		categories = make([]any, 0)
	}

	if outgoingSum > 0 {
		categories = upsertDetalizationCategory(
			categories,
			"SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
			"платежи и переводы",
			-outgoingSum,
			[]string{"operationExpenses"},
		)
	}
	if incomingSum > 0 {
		categories = upsertDetalizationCategory(
			categories,
			"REFILL",
			"пополнение баланса",
			incomingSum,
			[]string{"operationRefill"},
		)
	}

	data["categories"] = categories
}

func upsertDetalizationCategory(
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

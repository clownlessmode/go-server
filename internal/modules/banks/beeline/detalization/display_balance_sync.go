package detalization

import "project/internal/modules/banks/beeline/domain"

// ReconcileEndBalance shifts the opening anchor so the transaction chain ends at targetEnd.
func ReconcileEndBalance(data map[string]any, targetEnd float64) (float64, bool) {
	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) == 0 {
		return 0, false
	}

	targetEnd = domain.RoundMoney(targetEnd)

	var totalChange float64
	for _, item := range transactions {
		totalChange += transactionChangeValue(item)
	}
	totalChange = domain.RoundMoney(totalChange)
	opening := domain.RoundMoney(targetEnd - totalChange)

	return recalculateBalances(data, &opening)
}

// SyncDisplayBalance aligns detalization end balance with balance/main formula.
func SyncDisplayBalance(data map[string]any, apiBalance *float64, outgoingTotal, incomingTotal, hiddenNetChange float64) (float64, bool) {
	display := domain.DisplayBalanceFromAPI(apiBalance, outgoingTotal, incomingTotal, hiddenNetChange)
	if display == nil {
		return 0, false
	}

	return ReconcileEndBalance(data, *display)
}

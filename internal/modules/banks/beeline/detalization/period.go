package detalization

import (
	"fmt"
	"sort"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func ReportPeriodForMode(mode string, snapshotStart, snapshotEnd time.Time) (time.Time, time.Time) {
	switch mode {
	case "day":
		return lastDayPeriod(snapshotEnd)
	case "week":
		return lastWeekPeriod(snapshotEnd)
	default:
		return snapshotStart, snapshotEnd
	}
}

func lastDayPeriod(periodEnd time.Time) (time.Time, time.Time) {
	end := endOfReportDay(periodEnd)
	start := startOfReportDay(end)

	return start.UTC(), end.UTC()
}

func lastWeekPeriod(periodEnd time.Time) (time.Time, time.Time) {
	end := endOfReportDay(periodEnd)
	start := startOfReportDay(end.AddDate(0, 0, -6))

	return start.UTC(), end.UTC()
}

func TrimViewToPeriod(data map[string]any, periodStart, periodEnd time.Time) (map[string]any, float64, error) {
	working, err := CloneData(data)
	if err != nil {
		return nil, 0, err
	}

	transactions, ok := working["transactions"].([]any)
	if !ok || len(transactions) == 0 {
		return working, 0, fmt.Errorf("trim period: no transactions")
	}

	start := startOfReportDay(periodStart)
	end := endOfReportDay(periodEnd)

	openingBalance := balanceBeforePeriod(transactions, start)
	filtered := filterTransactionsInPeriod(transactions, start, end)
	if len(filtered) == 0 {
		return working, openingBalance, fmt.Errorf("trim period: no transactions in range")
	}

	if summary, ok := coreBalanceSummary(working); ok {
		summary["startValue"] = openingBalance
	}

	if first, ok := filtered[0].(map[string]any); ok {
		if balance, ok := coreBalanceEntry(first); ok {
			balance["startValue"] = openingBalance
		}
	}

	working["transactions"] = filtered

	finalBalance, ok := recalculateBalances(working)
	if !ok {
		return nil, 0, fmt.Errorf("trim period: recalculate balances")
	}

	return working, finalBalance, nil
}

func balanceBeforePeriod(transactions []any, periodStart time.Time) float64 {
	sorted := append([]any(nil), transactions...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return transactionDateTime(sorted[i]) < transactionDateTime(sorted[j])
	})

	running := findOpeningBalance(sorted)
	for _, item := range sorted {
		dateTime, ok := parseReportTransactionDateTime(transactionDateTime(item))
		if !ok {
			continue
		}
		if !dateTime.Before(periodStart) {
			break
		}

		change := transactionChangeValue(item)
		if change != 0 {
			running = domain.RoundMoney(running + change)
		}
	}

	return running
}

func filterTransactionsInPeriod(transactions []any, periodStart, periodEnd time.Time) []any {
	filtered := make([]any, 0, len(transactions))
	for _, item := range transactions {
		dateTime, ok := parseReportTransactionDateTime(transactionDateTime(item))
		if !ok {
			continue
		}
		if dateTime.Before(periodStart) || dateTime.After(periodEnd) {
			continue
		}

		filtered = append(filtered, item)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return transactionDateTime(filtered[i]) < transactionDateTime(filtered[j])
	})

	return filtered
}

func startOfReportDay(value time.Time) time.Time {
	location := reportLocation()
	value = value.In(location)

	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, location)
}

func endOfReportDay(value time.Time) time.Time {
	location := reportLocation()
	value = value.In(location)

	return time.Date(
		value.Year(),
		value.Month(),
		value.Day(),
		23, 59, 59,
		int(time.Second-time.Nanosecond),
		location,
	)
}

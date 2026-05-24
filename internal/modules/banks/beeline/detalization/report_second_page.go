package detalization

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

const (
	ReportSecondPageTransactionLimit = 13
	ReportDataPageTransactionLimit     = 17
)

type ReportSecondPageTransaction struct {
	DateTime    time.Time
	Title       string
	Description string
	Amount      string
}

func SecondPageTransactions(data map[string]any, limit int) []ReportSecondPageTransaction {
	return ReportTransactions(data, 0, limit)
}

func ReportTransactions(data map[string]any, offset, limit int) []ReportSecondPageTransaction {
	if limit <= 0 || offset < 0 {
		return nil
	}

	sorted := sortedReportTransactions(data)
	if len(sorted) == 0 || offset >= len(sorted) {
		return nil
	}

	end := offset + limit
	if end > len(sorted) {
		end = len(sorted)
	}

	return sorted[offset:end]
}

func CountReportTransactions(data map[string]any) int {
	return len(sortedReportTransactions(data))
}

func sortedReportTransactions(data map[string]any) []ReportSecondPageTransaction {
	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) == 0 {
		return nil
	}

	sorted := append([]any(nil), transactions...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return transactionDateTime(sorted[i]) > transactionDateTime(sorted[j])
	})

	result := make([]ReportSecondPageTransaction, 0, len(sorted))
	for _, item := range sorted {
		tx, ok := item.(map[string]any)
		if !ok {
			continue
		}

		dateTime, ok := parseReportTransactionDateTime(transactionDateTime(tx))
		if !ok {
			continue
		}

		result = append(result, ReportSecondPageTransaction{
			DateTime:    dateTime,
			Title:       reportTransactionTitle(tx),
			Description: FormatReportTransactionDescription(tx),
			Amount:      FormatReportTransactionAmount(transactionChangeValue(tx)),
		})
	}

	return result
}

func SameReportCalendarDay(left, right time.Time) bool {
	left = left.In(reportLocation())
	right = right.In(reportLocation())

	return left.Year() == right.Year() &&
		left.Month() == right.Month() &&
		left.Day() == right.Day()
}

func FormatReportSecondPageSectionDate(value time.Time) string {
	value = value.In(reportLocation())
	return fmt.Sprintf(
		"%d %s %d г.",
		value.Day(),
		russianMonthGenitive(value.Month()),
		value.Year(),
	)
}

func FormatReportTransactionDateTime(value time.Time) string {
	value = value.In(reportLocation())
	return fmt.Sprintf(
		"%d %s %d %02d:%02d",
		value.Day(),
		russianMonthShort(value.Month()),
		value.Year(),
		value.Hour(),
		value.Minute(),
	)
}

func FormatReportTransactionDescription(tx map[string]any) string {
	title := strings.ToLower(strings.TrimSpace(reportTransactionTitle(tx)))

	if strings.HasPrefix(title, "sms ") || title == "исходящее sms" {
		return "1 шт (основной баланс)"
	}

	if title == "безлимитный интернет" {
		return formatUnlimitedInternetDescription()
	}

	if strings.HasPrefix(title, "исходящий звонок") || strings.HasPrefix(title, "входящий звонок") {
		return formatCallDurationDescription()
	}

	switch title {
	case "начисление пакета минут":
		return "пакет минут"
	case "начисление пакета трафика":
		return "пакет интернета"
	case "пополнение баланса",
		"компенсация затрат на пополнение баланса",
		"плати с билайн: перевод на баланс билайн",
		"списание за мобильную коммерцию",
		"плата за подключение",
		"абонентская плата за тариф":
		return "основной баланс"
	default:
		return "основной баланс"
	}
}

func FormatReportTransactionAmount(change float64) string {
	change = domain.RoundMoney(change)
	if change > 0 {
		return "+" + formatReportAmount(change) + " ₽"
	}
	if change < 0 {
		return "-" + formatReportAmount(-change) + " ₽"
	}

	return "0,00 ₽"
}

func reportTransactionTitle(tx map[string]any) string {
	if name := strings.TrimSpace(jsonString(tx["name"])); name != "" {
		return name
	}

	return strings.TrimSpace(jsonString(tx["categoryName"]))
}

func parseReportTransactionDateTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	location := reportLocation()

	for _, layout := range layouts {
		if value, err := time.ParseInLocation(layout, raw, location); err == nil {
			return value, true
		}
	}

	return time.Time{}, false
}

func formatUnlimitedInternetDescription() string {
	integerPart := rand.Intn(99) + 1
	decimalPart := rand.Intn(10)

	unit := "мб"
	if rand.Intn(2) == 1 {
		unit = "кб"
	}

	return fmt.Sprintf("%d,%d %s (основной баланс)", integerPart, decimalPart, unit)
}

func formatCallDurationDescription() string {
	minutes := rand.Intn(15) + 1
	seconds := rand.Intn(60)

	return fmt.Sprintf("00:%02d:%02d (основной баланс)", minutes, seconds)
}

func russianMonthShort(month time.Month) string {
	switch month {
	case time.January:
		return "янв."
	case time.February:
		return "фев."
	case time.March:
		return "мар."
	case time.April:
		return "апр."
	case time.May:
		return "мая"
	case time.June:
		return "июн."
	case time.July:
		return "июл."
	case time.August:
		return "авг."
	case time.September:
		return "сен."
	case time.October:
		return "окт."
	case time.November:
		return "ноя."
	case time.December:
		return "дек."
	default:
		return ""
	}
}

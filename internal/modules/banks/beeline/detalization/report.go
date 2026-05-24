package detalization

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

const paymentsTransfersCategoryID = "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS"

type ReportFinance struct {
	Spent             float64
	Paid              float64
	Balance           float64
	OpeningBalance    float64
	PaymentsTransfers float64
	OtherSpent        float64
}

func FinanceTotals(data map[string]any) (ReportFinance, bool) {
	transactions, ok := data["transactions"].([]any)
	if !ok {
		return ReportFinance{}, false
	}

	var spent float64
	var paid float64
	var paymentsTransfers float64
	for _, item := range transactions {
		change := transactionChangeValue(item)
		switch {
		case change < 0:
			spent += -change
			if transactionCategoryID(item) == paymentsTransfersCategoryID {
				paymentsTransfers += -change
			}
		case change > 0:
			paid += change
		}
	}

	summary, ok := coreBalanceSummary(data)
	if !ok {
		return ReportFinance{}, false
	}

	spent = domain.RoundMoney(spent)
	paid = domain.RoundMoney(paid)
	paymentsTransfers = domain.RoundMoney(paymentsTransfers)
	otherSpent := domain.RoundMoney(spent - paymentsTransfers)
	if otherSpent < 0 {
		otherSpent = 0
	}

	return ReportFinance{
		Spent:             spent,
		Paid:              paid,
		Balance:           domain.RoundMoney(jsonNumber(summary["endValue"])),
		OpeningBalance:    openingBalanceAtPeriodStart(data),
		PaymentsTransfers: paymentsTransfers,
		OtherSpent:        otherSpent,
	}, true
}

func openingBalanceAtPeriodStart(data map[string]any) float64 {
	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) == 0 {
		if summary, ok := coreBalanceSummary(data); ok {
			return domain.RoundMoney(jsonNumber(summary["startValue"]))
		}

		return 0
	}

	sorted := append([]any(nil), transactions...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return transactionDateTime(sorted[i]) < transactionDateTime(sorted[j])
	})

	return findOpeningBalance(sorted)
}

func transactionCategoryID(item any) string {
	tx, ok := item.(map[string]any)
	if !ok {
		return ""
	}

	return strings.ToUpper(strings.TrimSpace(jsonString(tx["category"])))
}

func FormatReportPhone(number string) string {
	number = domain.NormalizeSimNumber(number)
	if len(number) != 10 {
		return "для номера +" + strings.TrimSpace(number)
	}

	return fmt.Sprintf(
		"для номера +7 %s %s-%s-%s",
		number[0:3],
		number[3:6],
		number[6:8],
		number[8:10],
	)
}

func FormatReportShortDate(value time.Time) string {
	value = value.In(reportLocation())
	return value.Format("02.01.06")
}

func FormatReportTitle(periodStart, periodEnd, createdAt time.Time) string {
	createdAt = createdAt.In(reportLocation())
	return fmt.Sprintf(
		"детализация с %s по %s создана %s в %s",
		FormatReportShortDate(periodStart),
		FormatReportShortDate(periodEnd),
		FormatReportShortDate(createdAt),
		createdAt.Format("15:04"),
	)
}

func FormatReportSpamLine(periodStart, periodEnd time.Time, blocked int) string {
	return fmt.Sprintf(
		"с %s по %s заблокировано спам звонков %d",
		FormatReportShortDate(periodStart),
		FormatReportShortDate(periodEnd),
		blocked,
	)
}

func FormatReportOpeningBalanceLine(date time.Time, balance float64) string {
	date = date.In(reportLocation())
	return fmt.Sprintf(
		"баланс на %d %s <span class=\"fc1\">%s</span>",
		date.Day(),
		russianMonthGenitive(date.Month()),
		formatReportBalanceAmount(balance),
	)
}

func FormatReportOpeningBalanceLineV2(date time.Time, balance float64) string {
	date = date.In(reportLocation())
	return fmt.Sprintf(
		"на %d %s <span class=\"fc1\">%s</span>",
		date.Day(),
		russianMonthGenitive(date.Month()),
		formatReportBalanceAmount(balance),
	)
}

func FormatReportPaymentsLine(value float64) string {
	return fmt.Sprintf(
		"платежи и переводы<span class=\"_ _3\"> </span><span class=\"fc1\">%s</span>",
		FormatReportCategorySpent(value),
	)
}

func FormatReportOtherLine(value float64) string {
	return fmt.Sprintf(
		"другое<span class=\"_ _4\"> </span><span class=\"fc1\">%s</span>",
		FormatReportCategorySpent(value),
	)
}

func FormatReportRefillLine(value float64) string {
	return fmt.Sprintf(
		"личный баланс<span class=\"_ _5\"> </span><span class=\"fc1\">%s</span>",
		FormatReportRefillAmount(value),
	)
}

func FormatReportPaymentsLineV2(value float64) string {
	return fmt.Sprintf(
		"платежи и переводы<span class=\"_ _0\"> </span><span class=\"fc1\">%s</span>",
		FormatReportCategorySpent(value),
	)
}

func FormatReportOtherLineV2(value float64) string {
	return fmt.Sprintf(
		"другое<span class=\"_\"> </span><span class=\"fc1 ws4\">%s</span>",
		FormatReportCategorySpent(value),
	)
}

func FormatReportRefillLineV2(value float64) string {
	return fmt.Sprintf(
		"личный баланс<span class=\"_ _1\"> </span><span class=\"fc1\">%s</span>",
		FormatReportRefillAmount(value),
	)
}

func FormatReportSpent(value float64) string {
	return "-" + formatReportAmount(value) + " ₽"
}

func FormatReportPaid(value float64) string {
	return "+" + formatReportAmount(value) + " ₽"
}

func FormatReportBalance(value float64) string {
	return formatReportBalanceAmount(value)
}

func FormatReportRefillAmount(value float64) string {
	return formatReportAmount(value) + " ₽"
}

func FormatReportCategorySpent(value float64) string {
	if value == 0 {
		return "0,00 ₽"
	}

	return "-" + formatReportAmount(value) + " ₽"
}

func formatReportBalanceAmount(value float64) string {
	value = domain.RoundMoney(value)
	if value < 0 {
		return "-" + formatReportAmount(value) + " ₽"
	}

	return formatReportAmount(value) + " ₽"
}

func formatReportAmount(value float64) string {
	value = math.Abs(domain.RoundMoney(value))
	intPart := int64(value)
	frac := int64(math.Round((value - float64(intPart)) * 100))
	if frac == 100 {
		intPart++
		frac = 0
	}

	intText := formatReportInteger(intPart)
	return fmt.Sprintf("%s,%02d", intText, frac)
}

func formatReportInteger(value int64) string {
	digits := fmt.Sprintf("%d", value)
	if len(digits) <= 3 {
		return digits
	}

	var parts []string
	for len(digits) > 3 {
		parts = append([]string{digits[len(digits)-3:]}, parts...)
		digits = digits[:len(digits)-3]
	}
	if digits != "" {
		parts = append([]string{digits}, parts...)
	}

	return strings.Join(parts, " ")
}

func reportLocation() *time.Location {
	return time.FixedZone("MSK", 3*60*60)
}

func russianMonthGenitive(month time.Month) string {
	switch month {
	case time.January:
		return "января"
	case time.February:
		return "февраля"
	case time.March:
		return "марта"
	case time.April:
		return "апреля"
	case time.May:
		return "мая"
	case time.June:
		return "июня"
	case time.July:
		return "июля"
	case time.August:
		return "августа"
	case time.September:
		return "сентября"
	case time.October:
		return "октября"
	case time.November:
		return "ноября"
	case time.December:
		return "декабря"
	default:
		return ""
	}
}

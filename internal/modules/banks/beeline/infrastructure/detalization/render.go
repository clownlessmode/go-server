package detalization

import (
	"strings"
	"time"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

const (
	templatePhoneLine          = "для номера +7 962 984-45-93"
	templateTitleLine          = "детализация с 17.02.26 по 16.03.26 создана 16.03.26 в 05:01"
	templateOpeningBalanceLine = "баланс на 17 февраля <span class=\"fc1\">0,00 ₽</span>"
	templateSpamLine           = "с 17.02.26 по 16.03.26 заблокировано спам звонков 0"
	templateSpentLine          = "-99 589,25 ₽"
	templatePaidLine           = "+99 591,00 ₽"
	templateBalance            = "1,75 ₽"
	templatePaymentsLine       = "платежи и переводы<span class=\"_ _3\"> </span><span class=\"fc1\">-98 840,94 ₽</span>"
	templateOtherLine          = "другое<span class=\"_ _4\"> </span><span class=\"fc1\">-748,31 ₽</span>"
	templateRefillLine         = "личный баланс<span class=\"_ _5\"> </span><span class=\"fc1\">99 591,00 ₽</span>"
)

type ReportParams struct {
	Phone            string
	PeriodStart      time.Time
	PeriodEnd        time.Time
	CreatedAt        time.Time
	Finance          detaildomain.ReportFinance
	DetalizationData map[string]any
	SpamBlocked      int
}

func renderFirstPageHTML(templateBody []byte, params ReportParams) []byte {
	if isFirstPageV2Template(templateBody) {
		return renderFirstPageHTMLV2(templateBody, params)
	}

	html := string(templateBody)
	html = injectLayoutCSS(html, amountFitCSS)
	replacements := map[string]string{
		templatePhoneLine:          detaildomain.FormatReportPhone(params.Phone),
		templateTitleLine:          detaildomain.FormatReportTitle(params.PeriodStart, params.PeriodEnd, params.CreatedAt),
		templateOpeningBalanceLine: detaildomain.FormatReportOpeningBalanceLine(params.PeriodStart, params.Finance.OpeningBalance),
		templateSpamLine:           detaildomain.FormatReportSpamLine(params.PeriodStart, params.PeriodEnd, params.SpamBlocked),
		templateSpentLine:          detaildomain.FormatReportSpent(params.Finance.Spent),
		templatePaidLine:           detaildomain.FormatReportPaid(params.Finance.Paid),
		templateBalance:            detaildomain.FormatReportBalance(params.Finance.Balance),
		templatePaymentsLine: detaildomain.FormatReportPaymentsLine(
			params.Finance.PaymentsTransfers,
		),
		templateOtherLine: detaildomain.FormatReportOtherLine(
			params.Finance.OtherSpent,
		),
		templateRefillLine: detaildomain.FormatReportRefillLine(
			params.Finance.Paid,
		),
	}

	for oldValue, newValue := range replacements {
		html = strings.ReplaceAll(html, oldValue, newValue)
	}

	return []byte(applyFirstPageAmountLines(html))
}

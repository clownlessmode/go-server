package detalization

import (
	"strings"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

const (
	templateV2PhoneLine          = "для номера +7 960 917-71-31"
	templateV2TitleLine          = "детализация с 17.05.26 по 23.05.26 создана 24.05.26 в 12:29"
	templateV2OpeningBalanceLine = "на 17 мая <span class=\"fc1\">6,76 ₽</span>"
	templateV2SpamLine           = "с 17.05.26 по 23.05.26 заблокировано спам звонков 0"
	templateV2SpentLine          = "-71,20 ₽"
	templateV2PaidLine           = "+100,00 ₽"
	templateV2Balance            = "35,56 ₽"
	templateV2PaymentsLine       = "платежи и переводы<span class=\"_ _0\"> </span><span class=\"fc1\">-70,00 ₽</span>"
	templateV2OtherLine          = "другое<span class=\"_\"> </span><span class=\"fc1 ws4\">-1,20 ₽</span>"
	templateV2RefillLine         = "личный баланс<span class=\"_ _1\"> </span><span class=\"fc1\">100,00 ₽</span>"
)

func isFirstPageV2Template(templateBody []byte) bool {
	return strings.Contains(string(templateBody), templateV2OpeningBalanceLine)
}

func renderFirstPageHTMLV2(templateBody []byte, params ReportParams) []byte {
	html := string(templateBody)
	html = injectLayoutCSS(html, amountFitCSS)
	replacements := map[string]string{
		templateV2PhoneLine:          detaildomain.FormatReportPhone(params.Phone),
		templateV2TitleLine:          detaildomain.FormatReportTitle(params.PeriodStart, params.PeriodEnd, params.CreatedAt),
		templateV2OpeningBalanceLine: detaildomain.FormatReportOpeningBalanceLineV2(params.PeriodStart, params.Finance.OpeningBalance),
		templateV2SpamLine:           detaildomain.FormatReportSpamLine(params.PeriodStart, params.PeriodEnd, params.SpamBlocked),
		templateV2SpentLine:          detaildomain.FormatReportSpent(params.Finance.Spent),
		templateV2PaidLine:           detaildomain.FormatReportPaid(params.Finance.Paid),
		templateV2Balance:            detaildomain.FormatReportBalance(params.Finance.Balance),
		templateV2PaymentsLine: detaildomain.FormatReportPaymentsLineV2(
			params.Finance.PaymentsTransfers,
		),
		templateV2OtherLine: detaildomain.FormatReportOtherLineV2(
			params.Finance.OtherSpent,
		),
		templateV2RefillLine: detaildomain.FormatReportRefillLineV2(
			params.Finance.Paid,
		),
	}

	for oldValue, newValue := range replacements {
		html = strings.ReplaceAll(html, oldValue, newValue)
	}

	return []byte(applyFirstPageV2AmountLines(html))
}

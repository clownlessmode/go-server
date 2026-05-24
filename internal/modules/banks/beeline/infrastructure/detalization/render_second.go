package detalization

import (
	"fmt"
	"strings"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

const (
	secondPageLayoutCSS      = amountFitCSS + secondPageSectionDateCSS
	templateSecondPageNumber = "2"
)

func renderSecondPageHTML(templateBody []byte, params ReportParams, page TransactionPageParams) []byte {
	html := string(templateBody)
	html = injectLayoutCSS(html, secondPageLayoutCSS)

	replacements := map[string]string{
		templatePhoneLine: detaildomain.FormatReportPhone(params.Phone),
		templateTitleLine: detaildomain.FormatReportTitle(params.PeriodStart, params.PeriodEnd, params.CreatedAt),
	}

	for oldValue, newValue := range replacements {
		html = strings.ReplaceAll(html, oldValue, newValue)
	}

	html = replaceOperationPageNumber(html, templateSecondPageNumber, fmt.Sprintf("%d", page.PageNumber))
	html = applySecondPageTransactions(html, params.DetalizationData, page.Offset, page.Limit)

	return []byte(html)
}

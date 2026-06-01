package detalization

import (
	"fmt"
	"strings"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

type operationSlot struct {
	titleY      string
	dateY       string
	descY       string
	amountClass string
	title       string
	date        string
	description string
	amount      string
}

func applyOperationSlots(
	html string,
	slots []operationSlot,
	transactions []detaildomain.ReportSecondPageTransaction,
	sectionDateY string,
	templateSectionDate string,
) string {
	if sectionDateY != "" && len(transactions) > 0 {
		html = replaceSectionDateDiv(
			html,
			sectionDateY,
			templateSectionDate,
			detaildomain.FormatReportSecondPageSectionDate(transactions[0].DateTime),
		)
	}

	for index, slot := range slots {
		title := ""
		date := ""
		description := ""
		amount := ""
		if index < len(transactions) {
			tx := transactions[index]
			title = tx.Title
			date = detaildomain.FormatReportTransactionDateTime(tx.DateTime)
			description = tx.Description
			amount = tx.Amount
		}

		html = replaceOperationTextDiv(html, "xd", "h7", "y"+slot.titleY, "fs3", "fc1", slot.title, title)
		html = replaceOperationTextDiv(html, "x1", "h6", "y"+slot.dateY, "fs5", "fc1", slot.date, date)
		html = replaceOperationTextDiv(html, "xd", "h7", "y"+slot.descY, "fs3", "fc2", slot.description, description)
		html = replaceOperationAmountDiv(html, slot.amountClass, "y"+slot.titleY, slot.amount, amount)
	}

	return html
}

func replaceSectionDateDiv(html, yClass, oldValue, newValue string) string {
	oldDiv := fmt.Sprintf(
		`<div class="t m0 xc h6 %s ff1 fs5 fc1 sc0 ls0 ws0">%s</div>`,
		yClass,
		oldValue,
	)
	newDiv := fmt.Sprintf(
		`<div class="t m0 section-date-fit h6 %s ff1 fs5 fc1 sc0 ls0 ws0">%s</div>`,
		yClass,
		escapeOperationHTML(newValue),
	)

	return strings.Replace(html, oldDiv, newDiv, 1)
}

func replaceOperationAmountDiv(html, xClass, yClass, oldValue, newValue string) string {
	oldDiv := fmt.Sprintf(
		`<div class="t m0 %s h7 %s ff1 fs3 fc1 sc0 ls0 ws0">%s</div>`,
		xClass,
		yClass,
		oldValue,
	)
	newDiv := fmt.Sprintf(
		`<div class="t m0 amount-fit h7 %s ff1 fs3 fc1 sc0 ls0 ws0">%s</div>`,
		yClass,
		escapeOperationHTML(newValue),
	)

	return strings.Replace(html, oldDiv, newDiv, 1)
}

func replaceOperationTextDiv(html, xClass, hClass, yClass, fsClass, fcClass, oldValue, newValue string) string {
	oldDiv := fmt.Sprintf(
		`<div class="t m0 %s %s %s ff1 %s %s sc0 ls0 ws0">%s</div>`,
		xClass,
		hClass,
		yClass,
		fsClass,
		fcClass,
		oldValue,
	)
	newDiv := fmt.Sprintf(
		`<div class="t m0 %s %s %s ff1 %s %s sc0 ls0 ws0">%s</div>`,
		xClass,
		hClass,
		yClass,
		fsClass,
		fcClass,
		escapeOperationHTML(newValue),
	)

	return strings.Replace(html, oldDiv, newDiv, 1)
}

func replaceOperationPageNumber(html, oldValue, newValue string) string {
	oldDiv := fmt.Sprintf(
		`<div class="t m0 x5 h4 y12 ff2 fs3 fc0 sc0 ls0 ws0">%s</div>`,
		oldValue,
	)
	newDiv := fmt.Sprintf(
		`<div class="t m0 x5 h4 y12 ff2 fs3 fc0 sc0 ls0 ws0">%s</div>`,
		escapeOperationHTML(newValue),
	)

	return strings.Replace(html, oldDiv, newDiv, 1)
}

func escapeOperationHTML(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(value)
}

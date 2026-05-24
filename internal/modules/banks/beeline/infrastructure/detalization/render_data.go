package detalization

import (
	"fmt"
	"strings"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

const dataPageDocumentWrapper = `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="utf-8">
<title>beeline detalization data page</title>
<style>
@font-face{
font-family:'Beeline Sans';
font-style:normal;
font-weight:400;
font-display:swap;
src:url('beelinesans-regular.woff2') format('woff2'),
url('beelinesans-regular.ttf') format('truetype');
}
html,body{margin:0;padding:0;background:white}
.page-logo svg{display:block;width:100%%;height:100%%}
.page-sheet{
width:595pt;
height:842pt;
margin:0 auto;
overflow:hidden;
position:relative;
background:white;
display:flex;
justify-content:center;
}
.page-content{
width:793.33px;
height:1122.67px;
flex-shrink:0;
font-family:'Beeline Sans',sans-serif;
}
.page-content,.page-content *{
font-family:'Beeline Sans',sans-serif!important;
}
.data-page-operation-divider{
align-self:stretch;
width:100%%;
height:0.5px;
flex-shrink:0;
background-image:url("data:image/svg+xml,%%3Csvg xmlns='http://www.w3.org/2000/svg' width='1' height='0.5' viewBox='0 0 1 0.5' shape-rendering='crispEdges'%%3E%%3Crect width='0.5' height='0.5' fill='%%23FAFAFC'/%%3E%%3C/svg%%3E");
background-repeat:repeat-x;
background-size:1px 0.5px;
-webkit-print-color-adjust:exact;
print-color-adjust:exact;
}
@media print{
@page{size:595pt 842pt;margin:0}
html,body{width:595pt;height:842pt;overflow:hidden;margin:0 auto}
.page-sheet{width:595pt;height:842pt;margin:0}
}
@media screen{
body{
min-height:100vh;
display:flex;
justify-content:center;
align-items:flex-start;
background:#9e9e9e;
}
.page-sheet{
margin:24px auto;
box-shadow:0 0 8px rgba(0,0,0,.2);
}
}
</style>
</head>
<body>
<div class="page-sheet">
<div class="page-content">
%s
</div>
</div>
</body>
</html>`

const (
	templateDataPagePhoneLine  = "для номера +7 962 985-25-47"
	templateDataPageTitleLine  = "детализация с 21.02.26 по 20.03.26 создана 20.03.26 в 03:50"
	dataPageNumberMarker       = "<!--DATA_PAGE_NUMBER-->"
	dataPageTransactionsMarker = "<!--DATA_PAGE_TRANSACTIONS-->"
	dataPageBannerMarker       = "<!--DATA_PAGE_BANNER-->"
)

const dataPageBannerHTML = `<div class="data-page-banner" style="align-self: stretch; margin-top: 27.2px; flex-shrink: 0;">
<img src="last-banner-3x.png" alt="" style="width: 100%; height: auto; display: block; border-radius: 12px;">
</div>`

const dataPageDateBadgeHTML = `<div style="align-self: stretch; height: 40px; padding-top: 15px; padding-bottom: 9.30px; padding-right: 10px; flex-direction: column; justify-content: flex-start; align-items: flex-start; gap: 10px; display: flex">
<div style="min-width: 107px; height: 16px; padding-left: 8px; padding-right: 8px; background: #F0F3F5; border-radius: 6px; justify-content: center; align-items: center; gap: 10px; display: inline-flex">
<div style="color: black; font-size: 12.09px; font-family: Beeline Sans; font-weight: 400; word-wrap: break-word">%s</div>
</div>
</div>`

const dataPageOperationDividerHTML = `<div class="data-page-operation-divider"></div>`

const dataPageOperationHTML = `<div style="align-self: stretch; padding-top: 11px; flex-direction: column; justify-content: flex-start; align-items: flex-start; gap: 4px; display: flex">
<div style="align-self: stretch; justify-content: flex-start; align-items: flex-start; gap: 14px; display: inline-flex">
<div style="height: 14px; justify-content: flex-start; align-items: center; gap: 10px; display: flex; flex-shrink: 0">
<div style="color: black; font-size: 12.03px; font-family: Beeline Sans; font-weight: 400; word-wrap: break-word">%s</div>
</div>
<div style="width: 553.97px; justify-content: space-between; align-items: flex-start; display: flex; flex-shrink: 0">
<div style="flex: 1 1 0; flex-direction: column; justify-content: flex-start; align-items: flex-start; gap: 8px; display: inline-flex">
<div style="color: black; font-size: 13.35px; font-family: Beeline Sans; font-weight: 400; word-wrap: break-word">%s</div>
<div style="color: #8F939C; font-size: 13.37px; font-family: Beeline Sans; font-weight: 400; word-wrap: break-word">%s</div>
</div>
<div style="color: black; font-size: 13.22px; font-family: Beeline Sans; font-weight: 400; letter-spacing: 0.13px; word-wrap: break-word">%s</div>
</div>
</div>
` + dataPageOperationDividerHTML + `
</div>`

func renderDataPageHTML(templateBody []byte, params ReportParams, page TransactionPageParams) []byte {
	html := string(templateBody)

	replacements := []struct {
		oldValue string
		newValue string
	}{
		{templateDataPagePhoneLine, detaildomain.FormatReportPhone(params.Phone)},
		{templateDataPageTitleLine, detaildomain.FormatReportTitle(params.PeriodStart, params.PeriodEnd, params.CreatedAt)},
		{dataPageNumberMarker, fmt.Sprintf("%d", page.PageNumber)},
		{dataPageTransactionsMarker, renderDataPageTransactions(params.DetalizationData, page)},
		{dataPageBannerMarker, renderDataPageBanner(page)},
	}

	for _, replacement := range replacements {
		html = strings.Replace(html, replacement.oldValue, replacement.newValue, 1)
	}

	html = injectDataPageLogo(html)
	return []byte(fmt.Sprintf(dataPageDocumentWrapper, html))
}

func renderDataPagePreviewHTML(templateBody []byte) []byte {
	return renderDataPageHTML(templateBody, ReportParams{}, TransactionPageParams{
		Offset:     0,
		Limit:      0,
		PageNumber: 4,
		ShowBanner: true,
		BannerOnly: true,
	})
}

func renderDataPageBanner(page TransactionPageParams) string {
	if !page.ShowBanner {
		return ""
	}

	return dataPageBannerHTML
}

func renderDataPageTransactions(data map[string]any, page TransactionPageParams) string {
	if page.BannerOnly || page.Limit <= 0 {
		return ""
	}

	items := buildDataPageItems(data, page.Offset, page.Limit)
	if len(items) == 0 {
		return ""
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		switch item.Kind {
		case dataPageItemBadge:
			parts = append(parts, fmt.Sprintf(
				dataPageDateBadgeHTML,
				escapeOperationHTML(detaildomain.FormatReportSecondPageSectionDate(item.DateTime)),
			))
		case dataPageItemOperation:
			tx := item.Tx
			parts = append(parts, fmt.Sprintf(
				dataPageOperationHTML,
				escapeOperationHTML(detaildomain.FormatReportTransactionDateTime(tx.DateTime)),
				escapeOperationHTML(tx.Title),
				escapeOperationHTML(tx.Description),
				escapeOperationHTML(tx.Amount),
			))
		}
	}

	return strings.Join(parts, "\n")
}

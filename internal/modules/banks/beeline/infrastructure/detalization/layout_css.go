package detalization

import "strings"

const amountFitCSS = `<style>
.amount-fit{width:400px!important;right:72px!important;left:auto!important;text-align:right!important;transform-origin:100% 100%!important;-webkit-transform-origin:100% 100%!important}
.amount-fit.ya{text-align:end!important}
@media print{
.amount-fit{width:355.555556pt!important;right:64pt!important;left:auto!important;text-align:right!important;transform-origin:100% 100%!important;-webkit-transform-origin:100% 100%!important}
.amount-fit.ya{text-align:end!important}
}
</style>`

const secondPageSectionDateCSS = `<style>
.section-date-cover{left:180px!important;width:272px!important;background:#fff!important;z-index:1!important;-webkit-print-color-adjust:exact;print-color-adjust:exact}
.section-date-badge{left:72px!important;width:auto!important;right:auto!important;z-index:2!important;display:flex!important;align-items:center!important;justify-content:flex-start!important;white-space:nowrap!important;background:transparent!important}
.section-date-badge span{display:inline-flex!important;align-items:center!important;justify-content:center!important;min-width:107px;height:16px;padding:0 8px;background:#F0F3F5!important;border-radius:6px;color:#000!important;font-size:12.09px!important;font-family:Beeline Sans,sans-serif!important;line-height:16px!important;-webkit-print-color-adjust:exact;print-color-adjust:exact}
@media print{
.section-date-cover{left:160pt!important;width:241.777778pt!important;background:#fff!important;z-index:1!important;-webkit-print-color-adjust:exact;print-color-adjust:exact}
.section-date-badge{left:64pt!important;z-index:2!important}
.section-date-badge span{min-width:95.111111pt!important;height:14.222222pt!important;padding:0 7.111111pt!important;border-radius:5.333333pt!important;font-size:10.746667pt!important;line-height:14.222222pt!important}
}
</style>`

func injectLayoutCSS(html, css string) string {
	if strings.Contains(html, "</head>") {
		return strings.Replace(html, "</head>", css+"</head>", 1)
	}

	return css + html
}

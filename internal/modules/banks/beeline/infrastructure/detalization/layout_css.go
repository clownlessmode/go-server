package detalization

import "strings"

const amountFitCSS = `<style>
.amount-fit{width:400px!important;right:72px!important;left:auto!important;text-align:right!important;transform-origin:100% 100%!important;-webkit-transform-origin:100% 100%!important}
@media print{
.amount-fit{width:355.555556pt!important;right:64pt!important;left:auto!important;text-align:right!important;transform-origin:100% 100%!important;-webkit-transform-origin:100% 100%!important}
}
</style>`

const secondPageSectionDateCSS = `<style>
.section-date-fit{width:308px!important;left:72px!important;right:auto!important;text-align:center!important;transform-origin:0 100%!important;-webkit-transform-origin:0 100%!important}
@media print{
.section-date-fit{width:273.777778pt!important;left:64pt!important;right:auto!important;text-align:center!important;transform-origin:0 100%!important;-webkit-transform-origin:0 100%!important}
}
</style>`

func injectLayoutCSS(html, css string) string {
	if strings.Contains(html, "</head>") {
		return strings.Replace(html, "</head>", css+"</head>", 1)
	}

	return css + html
}

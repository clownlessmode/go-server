package detalization

import (
	_ "embed"
	"regexp"
	"strings"
)

//go:embed templates/logo.svg
var pageLogoSVG []byte

// Logo sits on .pf (visible page), not inside scaled .pc.
// Screen coords use the 892.5px page width; print coords map to 595pt A4.
const pageLogoCSS = `<style>
.page-logo{position:absolute;top:48px;right:84px;width:46px;height:48px;z-index:10;pointer-events:none}
.page-logo svg{display:block;width:100%;height:100%}
@media print{
.page-logo{top:32pt;right:56pt;width:30.666667pt;height:32pt}
}
</style>`

const dataPageLogoPlaceholder = `<div class="page-logo" style="width: 35.90px; height: 35px; left: 680.17px; top: 45.08px; position: absolute; overflow: hidden"></div>`

var pageLogoInjectionPattern = regexp.MustCompile(`</div><div class="pi"`)

func injectPageLogo(htmlBody []byte) []byte {
	html := injectLayoutCSS(string(htmlBody), pageLogoCSS)
	logoHTML := `<div class="page-logo">` + strings.TrimSpace(string(pageLogoSVG)) + `</div>`

	return []byte(pageLogoInjectionPattern.ReplaceAllString(html, `</div>`+logoHTML+`<div class="pi"`))
}

func injectDataPageLogo(html string) string {
	logoHTML := strings.TrimSpace(string(pageLogoSVG))
	return strings.Replace(
		html,
		dataPageLogoPlaceholder,
		`<div class="page-logo" style="width: 35.90px; height: 35px; left: 680.17px; top: 45.08px; position: absolute; overflow: hidden">`+logoHTML+`</div>`,
		1,
	)
}

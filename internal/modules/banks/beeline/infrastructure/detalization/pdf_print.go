package detalization

import (
	"regexp"
	"strings"
)

var pdf2htmlEXScriptPattern = regexp.MustCompile(`(?s)<script>.*?</script>`)

const pdf2htmlEXPrintCSS = `<style>
#sidebar{display:none!important}
.pc{display:block!important;visibility:visible!important}
.pf{display:block!important;visibility:visible!important}
</style>`

func preparePDF2HtmlEXForPrint(htmlBody []byte) []byte {
	html := pdf2htmlEXScriptPattern.ReplaceAllString(string(htmlBody), "")
	if html == "" {
		return htmlBody
	}

	if idx := strings.Index(html, "</head>"); idx >= 0 {
		html = html[:idx] + pdf2htmlEXPrintCSS + html[idx:]
	} else {
		html = pdf2htmlEXPrintCSS + html
	}

	return []byte(html)
}

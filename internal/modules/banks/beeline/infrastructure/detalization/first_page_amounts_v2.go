package detalization

import "regexp"

var (
	firstPageV2AmountY9Pattern = regexp.MustCompile(
		`<div class="t m0 x1 h2 y9 ff1 fs0 fc0 sc0 ls0 ws4">(.*?)<span class="fc1">([^<]*)</span></div>`,
	)
	firstPageV2AmountYaPattern = regexp.MustCompile(
		`<div class="t m0 x1 h2 ya ff1 fs0 fc0 sc0 ls0 ws2">(.*?)<span class="fc1 ws4">([^<]*)</span></div>`,
	)
	firstPageV2AmountYcPattern = regexp.MustCompile(
		`<div class="t m0 x1 h2 yc ff1 fs0 fc0 sc0 ls0 ws4">(.*?)<span class="fc1">([^<]*)</span></div>`,
	)
)

func applyFirstPageV2AmountLines(html string) string {
	html = firstPageV2AmountY9Pattern.ReplaceAllString(
		html,
		`<div class="t m0 x1 h2 y9 ff1 fs0 fc0 sc0 ls0 ws4">$1</div><div class="t m0 amount-fit h2 y9 ff1 fs0 fc1 sc0 ls0 ws4">$2</div>`,
	)
	html = firstPageV2AmountYaPattern.ReplaceAllString(
		html,
		`<div class="t m0 x1 h2 ya ff1 fs0 fc0 sc0 ls0 ws2">$1</div><div class="t m0 amount-fit h2 ya ff1 fs0 fc1 sc0 ls0 ws4">$2</div>`,
	)
	html = firstPageV2AmountYcPattern.ReplaceAllString(
		html,
		`<div class="t m0 x1 h2 yc ff1 fs0 fc0 sc0 ls0 ws4">$1</div><div class="t m0 amount-fit h2 yc ff1 fs0 fc1 sc0 ls0 ws4">$2</div>`,
	)

	return html
}

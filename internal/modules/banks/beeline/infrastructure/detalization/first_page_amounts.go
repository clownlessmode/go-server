package detalization

import "regexp"

var firstPageAmountLinePattern = regexp.MustCompile(
	`<div class="t m0 x1 h1 (y(?:9|a|c)) ff1 fs0 fc0 sc0 ls0 ws0">(.*?)<span class="fc1">([^<]*)</span></div>`,
)

func applyFirstPageAmountLines(html string) string {
	return firstPageAmountLinePattern.ReplaceAllString(
		html,
		`<div class="t m0 x1 h1 $1 ff1 fs0 fc0 sc0 ls0 ws0">$2</div><div class="t m0 amount-fit h1 $1 ff1 fs0 fc1 sc0 ls0 ws0">$3</div>`,
	)
}

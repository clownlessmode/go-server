package detalization

import (
	"os"
	"strings"
	"testing"
	"time"

	detaildomain "project/internal/modules/banks/beeline/detalization"
)

func TestRenderFirstPageAmountLines(t *testing.T) {
	templateBody, err := os.ReadFile("templates/first-page.html")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	loc := time.FixedZone("MSK", 3*60*60)
	html := string(renderFirstPageHTML(templateBody, ReportParams{
		Phone:       "79629844593",
		PeriodStart: time.Date(2026, 2, 17, 0, 0, 0, 0, loc),
		PeriodEnd:   time.Date(2026, 3, 16, 0, 0, 0, 0, loc),
		CreatedAt:   time.Date(2026, 3, 16, 5, 1, 0, 0, loc),
		Finance: detaildomain.ReportFinance{
			Spent:             14999.46,
			Paid:              13100,
			Balance:           -1893.90,
			OpeningBalance:    -12.5,
			PaymentsTransfers: 14999.46,
			OtherSpent:        1.2,
		},
	}))

	checks := []string{
		`баланс на 17 февраля <span class="fc1">-12,50 ₽</span>`,
		`<div class="t m0 amount-fit h1 y9 ff1 fs0 fc1 sc0 ls0 ws0">-14 999,46 ₽</div>`,
		`<div class="t m0 amount-fit h1 ya ff1 fs0 fc1 sc0 ls0 ws0">-1,20 ₽</div>`,
		`<div class="t m0 amount-fit h1 yc ff1 fs0 fc1 sc0 ls0 ws0">13 100,00 ₽</div>`,
		"платежи и переводы<span class=\"_ _3\"> </span></div>",
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Fatalf("rendered html missing %q", check)
		}
	}

	if strings.Contains(html, `<span class="fc1">-1 893,90 ₽</span>`) {
		t.Fatal("opening balance line must not use end balance")
	}
}

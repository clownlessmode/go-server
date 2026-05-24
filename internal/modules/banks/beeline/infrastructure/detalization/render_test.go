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

	runRenderFirstPageHTMLTest(t, templateBody)
}

func runRenderFirstPageHTMLTest(t *testing.T, templateBody []byte) {
	t.Helper()

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

func runRenderFirstPageHTMLV2Test(t *testing.T, templateBody []byte) {
	t.Helper()

	loc := time.FixedZone("MSK", 3*60*60)
	html := string(renderFirstPageHTML(templateBody, ReportParams{
		Phone:       "79629844593",
		PeriodStart: time.Date(2026, 5, 17, 0, 0, 0, 0, loc),
		PeriodEnd:   time.Date(2026, 5, 23, 0, 0, 0, 0, loc),
		CreatedAt:   time.Date(2026, 5, 24, 12, 29, 0, 0, loc),
		Finance: detaildomain.ReportFinance{
			Spent:             71.20,
			Paid:              100,
			Balance:           35.56,
			OpeningBalance:    6.76,
			PaymentsTransfers: 70,
			OtherSpent:        1.2,
		},
	}))

	checks := []string{
		"для номера +7 962 984-45-93",
		"детализация с 17.05.26 по 23.05.26 создана 24.05.26 в 12:29",
		`на 17 мая <span class="fc1">6,76 ₽</span>`,
		"-71,20 ₽",
		"+100,00 ₽",
		"35,56 ₽",
		"с 17.05.26 по 23.05.26 заблокировано спам звонков 0",
		`<div class="t m0 amount-fit h2 y9 ff1 fs0 fc1 sc0 ls0 ws4">-70,00 ₽</div>`,
		`<div class="t m0 amount-fit h2 ya ff1 fs0 fc1 sc0 ls0 ws4">-1,20 ₽</div>`,
		`<div class="t m0 amount-fit h2 yc ff1 fs0 fc1 sc0 ls0 ws4">100,00 ₽</div>`,
		"платежи и переводы<span class=\"_ _0\"> </span></div>",
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Fatalf("rendered v2 html missing %q", check)
		}
	}

	if strings.Contains(html, templateV2PhoneLine) {
		t.Fatal("v2 phone placeholder was not replaced")
	}
}

package detalization

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestRenderSecondPageTransactions(t *testing.T) {
	templateBody, err := os.ReadFile("templates/second-page.html")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	loc := time.FixedZone("MSK", 3*60*60)
	data := map[string]any{
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-03-16T04:55:00",
				"name":     "плати с билайн: перевод на баланс билайн",
				"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -400.93, "name": "основной баланс"}},
			},
			map[string]any{
				"dateTime": "2026-03-16T04:55:00",
				"name":     "sms free8464",
				"volume":   1.0,
				"unit":     "шт",
				"balances": []any{map[string]any{"code": "coreBalance", "changeValue": 0.0, "name": "основной баланс"}},
			},
		},
	}

	html := string(renderSecondPageHTML(templateBody, ReportParams{
		Phone:            "79629844593",
		PeriodStart:      time.Date(2026, 2, 17, 0, 0, 0, 0, loc),
		PeriodEnd:        time.Date(2026, 3, 16, 0, 0, 0, 0, loc),
		CreatedAt:        time.Date(2026, 3, 16, 5, 1, 0, 0, loc),
		DetalizationData: data,
	}, TransactionPageParams{
		Offset:     0,
		Limit:      13,
		PageNumber: 2,
	}))

	checks := []string{
		"16 марта 2026 г.",
		"16 мар. 2026 04:55",
		"плати с билайн: перевод на баланс билайн",
		"-400,93 ₽",
		"sms free8464",
		"1 шт (основной баланс)",
		"amount-fit",
		"section-date-fit",
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Fatalf("rendered html missing %q", check)
		}
	}
}

func TestRenderDataPagePreviewHTML(t *testing.T) {
	templateBody, err := os.ReadFile("templates/data-page.html")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	html := string(renderDataPagePreviewHTML(templateBody))

	checks := []string{
		"<!DOCTYPE html>",
		"size:595pt 842pt",
		"beelinesans-regular.woff2",
		"font-family:'Beeline Sans'",
		"page-sheet",
		"page-content",
		"page-logo",
		"data-page-footer",
		"data-page-banner",
		"data-page-operation-divider",
		"last-banner-3x.png",
		`<svg width="40" height="40"`,
		"data-page-transactions",
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Fatalf("rendered html missing %q", check)
		}
	}
	if strings.Contains(html, "background: #313946") {
		t.Fatalf("black square placeholder still present")
	}
}

func TestRenderDataPageTransactions(t *testing.T) {
	templateBody, err := os.ReadFile("templates/data-page.html")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	loc := time.FixedZone("MSK", 3*60*60)
	data := map[string]any{
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-03-20T01:36:00",
				"name":     "пополнение баланса",
				"balances": []any{map[string]any{"code": "coreBalance", "changeValue": 300.0}},
			},
			map[string]any{
				"dateTime": "2026-03-19T19:35:00",
				"name":     "компенсация затрат на пополнение баланса",
				"balances": []any{map[string]any{"code": "coreBalance", "changeValue": -4.56}},
			},
		},
	}

	html := string(renderDataPageHTML(templateBody, ReportParams{
		Phone:            "79629852547",
		PeriodStart:      time.Date(2026, 2, 21, 0, 0, 0, 0, loc),
		PeriodEnd:        time.Date(2026, 3, 20, 0, 0, 0, 0, loc),
		CreatedAt:        time.Date(2026, 3, 20, 3, 50, 0, 0, loc),
		DetalizationData: data,
	}, TransactionPageParams{
		Offset:     0,
		Limit:      2,
		PageNumber: 4,
		ShowBanner: false,
	}))

	checks := []string{
		"для номера +7 962 985-25-47",
		"детализация с 21.02.26 по 20.03.26 создана 20.03.26 в 03:50",
		"20 мар. 2026 01:36",
		"пополнение баланса",
		"+300,00 ₽",
		"19 марта 2026 г.",
		"19 мар. 2026 19:35",
		"компенсация затрат на пополнение баланса",
		"-4,56 ₽",
		">4</div>",
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Fatalf("rendered html missing %q", check)
		}
	}
}

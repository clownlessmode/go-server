package detalization_test

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/detalization"
)

func TestFormatReportPhone(t *testing.T) {
	got := detalization.FormatReportPhone("9609177131")
	want := "для номера +7 960 917-71-31"
	if got != want {
		t.Fatalf("phone = %q, want %q", got, want)
	}
}

func TestFormatReportMoney(t *testing.T) {
	if got := detalization.FormatReportSpent(99589.25); got != "-99 589,25 ₽" {
		t.Fatalf("spent = %q", got)
	}
	if got := detalization.FormatReportPaid(99591); got != "+99 591,00 ₽" {
		t.Fatalf("paid = %q", got)
	}
	if got := detalization.FormatReportBalance(1.75); got != "1,75 ₽" {
		t.Fatalf("balance = %q", got)
	}
	if got := detalization.FormatReportBalance(-12.5); got != "-12,50 ₽" {
		t.Fatalf("negative balance = %q", got)
	}
}

func TestFormatReportOpeningBalanceLine(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	date := time.Date(2026, 4, 24, 0, 0, 0, 0, loc)

	got := detalization.FormatReportOpeningBalanceLine(date, -12.5)
	want := `баланс на 24 апреля <span class="fc1">-12,50 ₽</span>`
	if got != want {
		t.Fatalf("opening balance = %q, want %q", got, want)
	}
}

func TestFormatReportOpeningBalanceLineV2(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	date := time.Date(2026, 5, 17, 0, 0, 0, 0, loc)

	got := detalization.FormatReportOpeningBalanceLineV2(date, 6.76)
	want := `на 17 мая <span class="fc1">6,76 ₽</span>`
	if got != want {
		t.Fatalf("opening balance v2 = %q, want %q", got, want)
	}
}

func TestFormatReportTitle(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	start := time.Date(2026, 4, 24, 0, 0, 0, 0, loc)
	end := time.Date(2026, 5, 23, 23, 59, 59, 0, loc)
	created := time.Date(2026, 5, 23, 12, 27, 0, 0, loc)

	got := detalization.FormatReportTitle(start, end, created)
	want := "детализация с 24.04.26 по 23.05.26 создана 23.05.26 в 12:27"
	if got != want {
		t.Fatalf("title = %q, want %q", got, want)
	}
}

func TestFinanceTotals(t *testing.T) {
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"startValue": 0,
				"endValue":   1.75,
			},
		},
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-04-24T10:00:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 0, "startValue": -12.5},
				},
			},
			map[string]any{
				"dateTime": "2026-04-25T10:00:00",
				"category": "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": -98840.94},
				},
			},
			map[string]any{
				"dateTime": "2026-04-26T10:00:00",
				"category": "INTERNET",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": -748.31},
				},
			},
			map[string]any{
				"dateTime": "2026-04-27T10:00:00",
				"category": "refill",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 99591.0},
				},
			},
		},
	}

	totals, ok := detalization.FinanceTotals(data)
	if !ok {
		t.Fatal("expected totals")
	}
	if totals.Spent != 99589.25 || totals.Paid != 99591 || totals.Balance != 1.75 || totals.OpeningBalance != -12.5 {
		t.Fatalf("unexpected totals: %+v", totals)
	}
	if totals.PaymentsTransfers != 98840.94 || totals.OtherSpent != 748.31 {
		t.Fatalf("unexpected category totals: %+v", totals)
	}
}

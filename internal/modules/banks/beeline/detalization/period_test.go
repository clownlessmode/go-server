package detalization_test

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/detalization"
)

func TestReportPeriodForModeDay(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	snapshotStart := time.Date(2026, 4, 24, 0, 0, 0, 0, loc)
	snapshotEnd := time.Date(2026, 5, 23, 23, 59, 59, 0, loc)

	start, end := detalization.ReportPeriodForMode("day", snapshotStart, snapshotEnd)

	wantStart := time.Date(2026, 5, 23, 0, 0, 0, 0, loc)
	wantEnd := time.Date(2026, 5, 23, 23, 59, 59, int(time.Second-time.Nanosecond), loc)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("day period = %s..%s, want %s..%s", start, end, wantStart, wantEnd)
	}
}

func TestReportPeriodForModeWeek(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	snapshotStart := time.Date(2026, 4, 24, 0, 0, 0, 0, loc)
	snapshotEnd := time.Date(2026, 5, 23, 23, 59, 59, 0, loc)

	start, end := detalization.ReportPeriodForMode("week", snapshotStart, snapshotEnd)

	wantStart := time.Date(2026, 5, 17, 0, 0, 0, 0, loc)
	wantEnd := time.Date(2026, 5, 23, 23, 59, 59, int(time.Second-time.Nanosecond), loc)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("week period = %s..%s, want %s..%s", start, end, wantStart, wantEnd)
	}
}

func TestTrimViewToPeriodKeepsOpeningBalance(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"startValue": 0,
				"endValue":   100,
			},
		},
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-05-10T10:00:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 0, "startValue": 10.0, "endValue": 10.0},
				},
			},
			map[string]any{
				"dateTime": "2026-05-15T10:00:00",
				"category": "INTERNET",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": -5.0},
				},
			},
			map[string]any{
				"dateTime": "2026-05-20T10:00:00",
				"category": "refill",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 95.0},
				},
			},
		},
	}

	periodStart := time.Date(2026, 5, 17, 0, 0, 0, 0, loc)
	periodEnd := time.Date(2026, 5, 23, 23, 59, 59, 0, loc)

	view, finalBalance, err := detalization.TrimViewToPeriod(data, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("TrimViewToPeriod: %v", err)
	}

	transactions, ok := view["transactions"].([]any)
	if !ok || len(transactions) != 1 {
		t.Fatalf("expected 1 transaction in week view, got %d", len(transactions))
	}

	totals, ok := detalization.FinanceTotals(view)
	if !ok {
		t.Fatal("expected finance totals")
	}
	if totals.OpeningBalance != 5 {
		t.Fatalf("opening balance = %.2f, want 5.00", totals.OpeningBalance)
	}
	if totals.Paid != 95 || totals.Spent != 0 || totals.Balance != 100 || finalBalance != 100 {
		t.Fatalf("unexpected totals: %+v final=%.2f", totals, finalBalance)
	}
}

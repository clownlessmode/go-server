package detalization_test

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/detalization"
	"project/internal/modules/banks/beeline/domain"
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
					map[string]any{"code": "coreBalance", "changeValue": -5.0, "startValue": 10.0, "endValue": 5.0},
				},
			},
			map[string]any{
				"dateTime": "2026-05-20T10:00:00",
				"category": "refill",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 95.0, "startValue": 5.0, "endValue": 100.0},
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

	totals, ok := detalization.FinanceTotalsForPeriod(view, periodStart)
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

func TestShortPeriodViewIncludesTodayPaymentFromSnapshot(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	now := time.Date(2026, 6, 11, 16, 0, 0, 0, loc)

	payment := domain.Payment{
		ID:        "pay-today",
		Direction: domain.PaymentDirectionIncoming,
		Amount:    10000,
		Total:     10000,
		PaidAt:    time.Date(2026, 6, 11, 19, 23, 0, 0, loc),
	}

	baseData := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"startValue": 100.0,
				"endValue":   100.0,
			},
		},
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-05-20T10:00:00",
				"category": "INTERNET",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": -5.0},
				},
			},
		},
	}

	view, _, err := detalization.BuildView(baseData, []domain.Payment{payment}, nil, nil, "9053099079", now.UTC())
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}

	weekStart := time.Date(2026, 6, 5, 0, 0, 0, 0, loc).UTC()
	weekEnd := time.Date(2026, 6, 11, 23, 59, 59, 0, loc).UTC()

	trimmed, _, err := detalization.TrimViewToPeriod(view, weekStart, weekEnd)
	if err != nil {
		t.Fatalf("TrimViewToPeriod: %v", err)
	}

	if detalization.CountPaymentTransactionsOnDay(trimmed, now.UTC()) != 1 {
		t.Fatalf("expected 1 payment transaction today in week view")
	}
}

func TestTrimViewToPeriodSortsNewestFirst(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"startValue": 0.0,
				"endValue":   100.0,
			},
		},
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-06-10T10:00:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 10.0},
				},
			},
			map[string]any{
				"dateTime": "2026-06-11T19:23:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 20.0},
				},
			},
			map[string]any{
				"dateTime": "2026-06-11T19:36:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": -5.0},
				},
			},
			map[string]any{
				"dateTime": "2026-06-11T19:37:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": -3.0},
				},
			},
		},
	}

	periodStart := time.Date(2026, 6, 5, 0, 0, 0, 0, loc).UTC()
	periodEnd := time.Date(2026, 6, 11, 23, 59, 59, 0, loc).UTC()

	view, _, err := detalization.TrimViewToPeriod(data, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("TrimViewToPeriod: %v", err)
	}

	transactions, ok := view["transactions"].([]any)
	if !ok || len(transactions) < 2 {
		t.Fatalf("expected trimmed transactions, got %#v", view["transactions"])
	}

	first, _ := transactions[0].(map[string]any)["dateTime"].(string)
	second, _ := transactions[1].(map[string]any)["dateTime"].(string)
	if first <= second {
		t.Fatalf("expected newest first, got %s then %s", first, second)
	}
	if first != "2026-06-11T19:37:00" {
		t.Fatalf("first transaction = %s, want 2026-06-11T19:37:00", first)
	}
}

func TestTrimViewToPeriodPreservesReconciledEndValues(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"startValue": 1025.80,
				"endValue":   11025.80,
			},
		},
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-06-10T18:00:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": -100.0, "startValue": 1125.80, "endValue": 1025.80},
				},
			},
			map[string]any{
				"dateTime": "2026-06-11T19:23:00",
				"category": "refill",
				"name":     "пополнение баланса",
				"source":   "payment",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 10000.0, "startValue": 1025.80, "endValue": 11025.80},
				},
			},
			map[string]any{
				"dateTime": "2026-06-11T19:36:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": -1500.0, "startValue": 11025.80, "endValue": 9525.80},
				},
			},
		},
	}

	dayStart := time.Date(2026, 6, 11, 0, 0, 0, 0, loc).UTC()
	dayEnd := time.Date(2026, 6, 11, 23, 59, 59, 0, loc).UTC()

	view, finalBalance, err := detalization.TrimViewToPeriod(data, dayStart, dayEnd)
	if err != nil {
		t.Fatalf("TrimViewToPeriod: %v", err)
	}
	if finalBalance != 9525.80 {
		t.Fatalf("final balance = %.2f, want 9525.80", finalBalance)
	}

	for _, item := range view["transactions"].([]any) {
		tx := item.(map[string]any)
		if tx["dateTime"] != "2026-06-11T19:23:00" {
			continue
		}

		endValue := tx["balances"].([]any)[0].(map[string]any)["endValue"].(float64)
		if endValue != 11025.80 {
			t.Fatalf("refill endValue = %.2f, want 11025.80", endValue)
		}
		return
	}

	t.Fatal("refill transaction not found in trimmed view")
}

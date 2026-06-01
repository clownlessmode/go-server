package detalization

import (
	"testing"
	"time"
)

func serverSnapshot9063747835() map[string]any {
	return map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "bonusseconds",
				"name":       "пакет минут",
				"startValue": 450,
				"endValue":   550,
			},
			map[string]any{
				"code":       "coreBalance",
				"name":       "личный баланс",
				"startValue": 450,
				"endValue":   450,
			},
		},
		"transactions": []any{
			map[string]any{
				"category": "refill",
				"name":     "начисление пакета трафика",
				"dateTime": "2026-06-01T15:52:00",
				"balances": []any{
					map[string]any{"code": "internet", "changeValue": 1073741824},
				},
			},
			map[string]any{
				"category": "refill",
				"name":     "начисление пакета минут",
				"dateTime": "2026-06-01T15:52:00",
				"balances": []any{
					map[string]any{"code": "bonusseconds", "changeValue": 60000},
				},
			},
			map[string]any{
				"category": "SUBSCRIBER_FEE_AND_SERVICE_CHANGE_FEE",
				"name":     "абонентская плата по тарифу «ледовый»",
				"dateTime": "2026-06-01T15:52:00",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": -350,
						"startValue":  900,
						"endValue":    550,
					},
				},
			},
			map[string]any{
				"category": "refill",
				"name":     "пополнение баланса",
				"dateTime": "2026-06-01T15:50:03",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": 450,
						"startValue":  450,
						"endValue":    900,
					},
				},
			},
		},
	}
}

func TestRecalculateBalancesUsesCoreBalanceSummary(t *testing.T) {
	data := serverSnapshot9063747835()

	balance, ok := recalculateBalances(data, nil)
	if !ok {
		t.Fatal("recalculateBalances failed")
	}
	if balance != 100 {
		t.Fatalf("balance = %.2f, want 100 without configured opening", balance)
	}

	summary, ok := coreBalanceSummary(data)
	if !ok {
		t.Fatal("coreBalanceSummary failed")
	}
	if jsonString(summary["code"]) != "coreBalance" {
		t.Fatalf("summary code = %q, want coreBalance", summary["code"])
	}
	if jsonNumber(summary["endValue"]) != 100 {
		t.Fatalf("coreBalance endValue = %.2f, want 100", summary["endValue"])
	}
}

func TestRecalculateBalancesWithConfiguredOpeningZero(t *testing.T) {
	data := serverSnapshot9063747835()
	opening := 0.0

	balance, ok := recalculateBalances(data, &opening)
	if !ok {
		t.Fatal("recalculateBalances failed")
	}
	if balance != 100 {
		t.Fatalf("balance = %.2f, want 100", balance)
	}

	summary, ok := coreBalanceSummary(data)
	if !ok {
		t.Fatal("coreBalanceSummary failed")
	}
	if jsonNumber(summary["startValue"]) != 0 {
		t.Fatalf("startValue = %.2f, want 0", summary["startValue"])
	}
	if jsonNumber(summary["endValue"]) != 100 {
		t.Fatalf("endValue = %.2f, want 100", summary["endValue"])
	}
}

func TestFinanceTotalsForPeriodJune1OpeningZero(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	data := serverSnapshot9063747835()
	opening := 0.0

	if _, ok := recalculateBalances(data, &opening); !ok {
		t.Fatal("recalculateBalances failed")
	}

	periodStart := time.Date(2026, 6, 1, 0, 0, 0, 0, loc)

	totals, ok := FinanceTotalsForPeriod(data, periodStart)
	if !ok {
		t.Fatal("expected finance totals")
	}
	if totals.OpeningBalance != 0 {
		t.Fatalf("opening balance = %.2f, want 0", totals.OpeningBalance)
	}
	if totals.Balance != 100 {
		t.Fatalf("balance = %.2f, want 100", totals.Balance)
	}
}

func TestTrimViewToPeriodJune1OpeningZero(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	data := serverSnapshot9063747835()
	opening := 0.0

	if _, ok := recalculateBalances(data, &opening); !ok {
		t.Fatal("recalculateBalances failed")
	}

	periodStart := time.Date(2026, 6, 1, 0, 0, 0, 0, loc)
	periodEnd := time.Date(2026, 6, 1, 23, 59, 59, int(time.Second-time.Nanosecond), loc)

	view, finalBalance, err := TrimViewToPeriod(data, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("TrimViewToPeriod: %v", err)
	}

	totals, ok := FinanceTotalsForPeriod(view, periodStart)
	if !ok {
		t.Fatal("expected finance totals")
	}
	if totals.OpeningBalance != 0 {
		t.Fatalf("opening balance = %.2f, want 0", totals.OpeningBalance)
	}
	if totals.Paid != 450 || totals.Spent != 350 || totals.Balance != 100 || finalBalance != 100 {
		t.Fatalf("unexpected totals: %+v final=%.2f", totals, finalBalance)
	}
}

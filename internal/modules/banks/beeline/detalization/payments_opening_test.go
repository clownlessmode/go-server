package detalization

import (
	"testing"
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

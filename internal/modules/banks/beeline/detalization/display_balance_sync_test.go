package detalization

import (
	"testing"

	"project/internal/modules/banks/beeline/domain"
)

func TestSyncDisplayBalanceAlignsDetalizationEndValue(t *testing.T) {
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"name":       "личный баланс",
				"startValue": 0,
				"endValue":   41.2,
			},
		},
		"transactions": []any{
			map[string]any{
				"category": "OTHER",
				"name":     "period opening",
				"dateTime": "2026-06-01T00:00:00",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": 0,
						"startValue":  41.2,
						"endValue":    41.2,
					},
				},
			},
			map[string]any{
				"category": "refill",
				"name":     "пополнение баланса",
				"dateTime": "2026-06-02T12:00:00",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": 10000,
						"startValue":  41.2,
						"endValue":    10041.2,
					},
				},
			},
		},
	}

	api := 61.36
	synced, ok := SyncDisplayBalance(data, &api, 0, 10000, 0)
	if !ok {
		t.Fatal("SyncDisplayBalance failed")
	}
	if synced != 10061.36 {
		t.Fatalf("synced balance = %.2f, want 10061.36", synced)
	}

	summary, ok := coreBalanceSummary(data)
	if !ok {
		t.Fatal("summary missing")
	}
	if domain.RoundMoney(jsonNumber(summary["endValue"])) != 10061.36 {
		t.Fatalf("summary endValue = %.2f, want 10061.36", summary["endValue"])
	}
}

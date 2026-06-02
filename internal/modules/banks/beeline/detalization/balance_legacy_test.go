package detalization

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func TestBuildViewUsesSnapshotAnchorNotLegacyConfiguredOpening(t *testing.T) {
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"name":       "личный баланс",
				"startValue": 61.44,
				"endValue":   1075.42,
			},
		},
		"transactions": []any{
			map[string]any{
				"category": "OTHER",
				"name":     "period opening",
				"dateTime": "2026-05-01T00:00:00",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": 0,
						"startValue":  61.44,
						"endValue":    61.44,
					},
				},
			},
			map[string]any{
				"category": "refill",
				"name":     "пополнение баланса",
				"dateTime": "2026-05-20T10:00:00",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": 10000,
						"startValue":  61.44,
						"endValue":    10061.44,
					},
				},
			},
			map[string]any{
				"category": "MOBILE_COMMERCE",
				"name":     "мобильная коммерция",
				"dateTime": "2026-05-21T12:00:00",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": -9052.5,
						"startValue":  10061.44,
						"endValue":    1008.94,
					},
				},
			},
		},
	}

	_, balance, err := BuildView(data, nil, nil, nil, "9680659702", time.Date(2026, 6, 2, 12, 0, 0, 0, domain.BeelineLocation()))
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}
	if balance != 1008.94 {
		t.Fatalf("balance = %.2f, want 1008.94 from snapshot anchor", balance)
	}

	staleOpening := 51.28
	_, staleBalance, err := BuildView(data, nil, nil, &staleOpening, "9680659702", time.Date(2026, 6, 2, 12, 0, 0, 0, domain.BeelineLocation()))
	if err != nil {
		t.Fatalf("BuildView with configured opening: %v", err)
	}
	if staleBalance == balance {
		t.Fatalf("legacy configured opening must change balance (got same %.2f)", staleBalance)
	}

	// EffectiveBalance-style shortcut that ignores real Beeline refills in snapshot.
	effective := 51.28 + 10000 - 9052.5
	if effective != 998.78 {
		t.Fatalf("sanity check effective = %.2f, want 998.78", effective)
	}
	if staleBalance != 998.78 {
		t.Fatalf("stale configured opening balance = %.2f, want 998.78 (bug scenario)", staleBalance)
	}
}

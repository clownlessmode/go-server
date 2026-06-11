package detalization

import "testing"

func TestHiddenTransactionsNetChangeSumsHiddenDebits(t *testing.T) {
	data := map[string]any{
		"transactions": []any{
			map[string]any{
				"dateTime": "2026-05-22T16:20:40",
				"category": "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
				"name":     "списание за мобильную коммерцию",
				"balances": []any{
					map[string]any{"changeValue": -13845.0},
				},
			},
			map[string]any{
				"dateTime": "2026-05-23T12:00:00",
				"category": "refill",
				"name":     "пополнение баланса",
				"balances": []any{
					map[string]any{"changeValue": 5000.0},
				},
			},
		},
	}

	hiddenID := TransactionID(data["transactions"].([]any)[0].(map[string]any))
	net := HiddenTransactionsNetChange(data, []string{hiddenID})
	if net != -13845 {
		t.Fatalf("hidden net = %.2f, want -13845", net)
	}
}

func TestSyncDisplayBalanceAdjustsForHiddenTransactions(t *testing.T) {
	data := map[string]any{
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"name":       "личный баланс",
				"startValue": 0,
				"endValue":   0,
			},
		},
		"transactions": []any{
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
	hiddenDebit := -13845.0
	synced, ok := SyncDisplayBalance(data, &api, 0, 10000, hiddenDebit)
	if !ok {
		t.Fatal("SyncDisplayBalance failed")
	}

	want := 10061.36 + 13845
	if synced != want {
		t.Fatalf("synced balance = %.2f, want %.2f", synced, want)
	}
}

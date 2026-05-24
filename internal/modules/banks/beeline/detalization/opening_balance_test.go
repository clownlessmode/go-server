package detalization_test

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/detalization"
	"project/internal/modules/banks/beeline/domain"
)

func TestFinanceTotalsAfterBuildViewPreservesOpeningBalance(t *testing.T) {
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
				"dateTime": "2026-02-17T10:00:00",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 0, "startValue": -12.5},
				},
			},
			map[string]any{
				"dateTime": "2026-02-25T10:00:00",
				"category": "refill",
				"balances": []any{
					map[string]any{"code": "coreBalance", "changeValue": 99591.0},
				},
			},
		},
	}

	payments := []domain.Payment{
		{
			ID:        "p1",
			Direction: domain.PaymentDirectionIncoming,
			Amount:    100,
			PaidAt:    time.Date(2026, 2, 17, 11, 0, 0, 0, time.UTC),
		},
	}

	view, finalBalance, err := detalization.BuildView(data, payments, nil)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}

	totals, ok := detalization.FinanceTotals(view)
	if !ok {
		t.Fatal("expected totals")
	}

	if totals.Balance != finalBalance {
		t.Fatalf("balance mismatch: totals=%v final=%v", totals.Balance, finalBalance)
	}
	if totals.OpeningBalance != -12.5 {
		t.Fatalf("opening balance = %v, want -12.5", totals.OpeningBalance)
	}
}

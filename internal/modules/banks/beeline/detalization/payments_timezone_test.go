package detalization

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func TestApplyPaymentsSkipsPaymentFlowSMS(t *testing.T) {
	payment := domain.NewPaymentFlowSMSPayment(time.Date(2026, 6, 1, 12, 21, 0, 0, time.UTC))
	data := map[string]any{
		"transactions": []any{
			map[string]any{
				"category": "OTHER",
				"name":     "period opening",
				"dateTime": "2026-06-01T00:00:00",
				"balances": []any{
					map[string]any{
						"code":        "coreBalance",
						"changeValue": 0,
						"startValue":  0,
						"endValue":    0,
					},
				},
			},
		},
		"balances": []any{
			map[string]any{
				"code":       "coreBalance",
				"startValue": 0,
				"endValue":   0,
				"spentValue": 0,
				"paidValue":  0,
			},
		},
	}

	if _, ok := ApplyPayments(data, []domain.Payment{payment}, nil); !ok {
		t.Fatal("ApplyPayments failed")
	}

	transactions, ok := data["transactions"].([]any)
	if !ok || len(transactions) != 1 {
		t.Fatalf("transactions = %#v, want only opening row", data["transactions"])
	}
}

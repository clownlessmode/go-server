package detalization

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func TestPaymentTransactionDateTimeInMoscow(t *testing.T) {
	payment := domain.NewPaymentFlowSMSPayment(time.Date(2026, 6, 1, 12, 21, 0, 0, time.UTC))
	data := map[string]any{
		"transactions": []any{},
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
		t.Fatalf("transactions = %#v", data["transactions"])
	}

	tx, ok := transactions[0].(map[string]any)
	if !ok {
		t.Fatalf("transaction type = %T", transactions[0])
	}

	got, _ := tx["dateTime"].(string)
	if got != "2026-06-01T15:21:00" {
		t.Fatalf("dateTime = %q", got)
	}
}

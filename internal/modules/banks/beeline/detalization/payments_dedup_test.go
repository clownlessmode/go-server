package detalization

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func testBalanceData() map[string]any {
	return map[string]any{
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
}

func beelineRefillTransaction(dateTime string, amount float64) map[string]any {
	return map[string]any{
		"category": "refill",
		"name":     "пополнение баланса",
		"dateTime": dateTime,
		"balances": []any{
			map[string]any{
				"changeValue": amount,
				"code":        "coreBalance",
			},
		},
	}
}

func TestApplyPaymentsSkipsDuplicateRefillSameMinute(t *testing.T) {
	data := testBalanceData()
	data["transactions"] = []any{
		beelineRefillTransaction("2026-06-01T15:21:47", 450),
	}

	payment := domain.Payment{
		ID:        "pay-1",
		Direction: domain.PaymentDirectionIncoming,
		Amount:    450,
		PaidAt:    time.Date(2026, 6, 1, 15, 21, 0, 0, domain.BeelineLocation()),
	}

	balance, ok := ApplyPayments(data, []domain.Payment{payment})
	if !ok {
		t.Fatal("ApplyPayments failed")
	}
	if balance != 450 {
		t.Fatalf("balance = %.2f, want 450", balance)
	}

	transactions, _ := data["transactions"].([]any)
	if len(transactions) != 1 {
		t.Fatalf("transactions count = %d, want 1", len(transactions))
	}
}

func TestApplyPaymentsSkipsDuplicateRefillTimezoneSkew(t *testing.T) {
	data := testBalanceData()
	data["transactions"] = []any{
		beelineRefillTransaction("2026-06-01T12:21:00", 450),
	}

	payment := domain.Payment{
		ID:        "pay-1",
		Direction: domain.PaymentDirectionIncoming,
		Amount:    450,
		PaidAt:    time.Date(2026, 6, 1, 15, 21, 0, 0, domain.BeelineLocation()),
	}

	balance, ok := ApplyPayments(data, []domain.Payment{payment})
	if !ok {
		t.Fatal("ApplyPayments failed")
	}
	if balance != 450 {
		t.Fatalf("balance = %.2f, want 450", balance)
	}

	transactions, _ := data["transactions"].([]any)
	if len(transactions) != 1 {
		t.Fatalf("transactions count = %d, want 1", len(transactions))
	}
}

func TestApplyPaymentsKeepsTwoRefillsSameDayDifferentTimes(t *testing.T) {
	data := testBalanceData()
	data["transactions"] = []any{
		beelineRefillTransaction("2026-06-01T10:00:00", 450),
	}

	payment := domain.Payment{
		ID:        "pay-1",
		Direction: domain.PaymentDirectionIncoming,
		Amount:    450,
		PaidAt:    time.Date(2026, 6, 1, 18, 0, 0, 0, domain.BeelineLocation()),
	}

	balance, ok := ApplyPayments(data, []domain.Payment{payment})
	if !ok {
		t.Fatal("ApplyPayments failed")
	}
	if balance != 900 {
		t.Fatalf("balance = %.2f, want 900", balance)
	}

	transactions, _ := data["transactions"].([]any)
	if len(transactions) != 2 {
		t.Fatalf("transactions count = %d, want 2", len(transactions))
	}
}

func TestApplyPaymentsUserScenario(t *testing.T) {
	data := testBalanceData()

	incoming := domain.Payment{
		ID:        "pay-in",
		Direction: domain.PaymentDirectionIncoming,
		Amount:    450,
		PaidAt:    time.Date(2026, 6, 1, 15, 21, 0, 0, domain.BeelineLocation()),
	}
	data["transactions"] = []any{beelineRefillTransaction("2026-06-01T15:21:30", 450)}

	balance, ok := ApplyPayments(data, []domain.Payment{incoming})
	if !ok || balance != 450 {
		t.Fatalf("after refill balance = %.2f, want 450", balance)
	}

	outgoing := domain.Payment{
		ID:        "pay-out",
		Direction: domain.PaymentDirectionOutgoing,
		Amount:    350,
		Total:     350,
		PaidAt:    time.Date(2026, 6, 1, 16, 0, 0, 0, domain.BeelineLocation()),
	}
	data["transactions"] = append(data["transactions"].([]any), map[string]any{
		"category": "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
		"name":     "списание за мобильную коммерцию",
		"dateTime": "2026-06-01T16:00:15",
		"balances": []any{
			map[string]any{
				"changeValue": -350,
				"code":        "coreBalance",
			},
		},
	})

	balance, ok = ApplyPayments(data, []domain.Payment{incoming, outgoing})
	if !ok || balance != 100 {
		t.Fatalf("final balance = %.2f, want 100", balance)
	}
}

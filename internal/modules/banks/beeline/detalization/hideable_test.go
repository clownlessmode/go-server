package detalization

import (
	"testing"
	"time"

	"project/internal/modules/banks/beeline/domain"
)

func TestAnnotateTransactionIDsMarksCardTransferHideable(t *testing.T) {
	data := map[string]any{
		"transactions": []any{
			map[string]any{
				"category": "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
				"name":     "Перевод с баланса на карту",
				"dateTime": "2026-06-05T22:22:53.594",
				"balances": []any{
					map[string]any{"changeValue": -14910, "code": "coreBalance"},
				},
			},
			map[string]any{
				"category": "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
				"name":     paymentCardTransferName,
				"dateTime": "2026-06-05T22:01:28",
				"balances": []any{
					map[string]any{"changeValue": -14910, "code": "coreBalance"},
				},
			},
		},
	}

	payments := []domain.Payment{{
		ID:        "pay-14910",
		Direction: domain.PaymentDirectionOutgoing,
		Amount:    14000,
		Total:     14910,
		PaidAt:    time.Date(2026, 6, 5, 22, 1, 28, 0, domain.BeelineLocation()),
	}}

	AnnotateTransactionIDs(data, payments)

	transfer := data["transactions"].([]any)[0].(map[string]any)
	if transfer["source"] != "beeline" {
		t.Fatalf("transfer source = %v, want beeline", transfer["source"])
	}
	if transfer["hideable"] != true {
		t.Fatalf("transfer hideable = %v, want true", transfer["hideable"])
	}

	commerce := data["transactions"].([]any)[1].(map[string]any)
	if commerce["name"] != paymentCardTransferName {
		t.Fatalf("payment row name = %v, want %q", commerce["name"], paymentCardTransferName)
	}
	if commerce["source"] != "payment" {
		t.Fatalf("commerce source = %v, want payment", commerce["source"])
	}
	if commerce["hideable"] != false {
		t.Fatalf("commerce hideable = %v, want false", commerce["hideable"])
	}
}

func TestIsMobileCommerceTransactionIncludesCardTransfer(t *testing.T) {
	tx := map[string]any{
		"category": "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
		"name":     "Перевод с баланса на карту",
	}
	if isMobileCommerceTransaction(tx) {
		t.Fatal("card transfer must not be mobile commerce")
	}
	if !isCardTransferTransaction(tx) {
		t.Fatal("expected card transfer")
	}
}

func TestApplyPaymentsInjectsCardTransferWhenNotInSnapshot(t *testing.T) {
	data := testBalanceData()

	payment := domain.Payment{
		ID:        "pay-14910",
		Direction: domain.PaymentDirectionOutgoing,
		Amount:    14000,
		Total:     14910,
		PaidAt:    time.Date(2026, 6, 5, 20, 17, 0, 0, domain.BeelineLocation()),
	}

	balance, ok := ApplyPayments(data, []domain.Payment{payment}, nil)
	if !ok {
		t.Fatal("ApplyPayments failed")
	}
	if balance != -14910 {
		t.Fatalf("balance = %.2f, want -14910", balance)
	}

	transactions, _ := data["transactions"].([]any)
	if len(transactions) != 1 {
		t.Fatalf("transactions count = %d, want 1", len(transactions))
	}
	tx := transactions[0].(map[string]any)
	if tx["name"] != paymentCardTransferName {
		t.Fatalf("name = %v, want %q", tx["name"], paymentCardTransferName)
	}
}

func TestApplyPaymentsSkipsInjectWhenCardTransferAlreadyExists(t *testing.T) {
	data := testBalanceData()
	data["transactions"] = []any{
		map[string]any{
			"category": "SERVICES_PAYMENTS_AND_MOBILE_TRANSFERS",
			"name":     paymentCardTransferName,
			"dateTime": "2026-06-05T22:22:53",
			"balances": []any{map[string]any{"changeValue": -14910, "code": "coreBalance"}},
		},
	}

	payment := domain.Payment{
		ID:        "pay-14910",
		Direction: domain.PaymentDirectionOutgoing,
		Amount:    14000,
		Total:     14910,
		PaidAt:    time.Date(2026, 6, 5, 20, 17, 0, 0, domain.BeelineLocation()),
	}

	balance, ok := ApplyPayments(data, []domain.Payment{payment}, nil)
	if !ok {
		t.Fatal("ApplyPayments failed")
	}
	if balance != -14910 {
		t.Fatalf("balance = %.2f, want -14910", balance)
	}

	transactions, _ := data["transactions"].([]any)
	if len(transactions) != 1 {
		t.Fatalf("transactions count = %d, want 1", len(transactions))
	}
}

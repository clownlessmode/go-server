package domain

import (
	"testing"
	"time"
)

func TestParsePaymentDirectionBalanceReturn(t *testing.T) {
	got, err := ParsePaymentDirection("balance_return")
	if err != nil {
		t.Fatalf("ParsePaymentDirection: %v", err)
	}
	if got != PaymentDirectionBalanceReturn {
		t.Fatalf("direction = %q, want balance_return", got)
	}
}

func TestNewManualPaymentBalanceReturn(t *testing.T) {
	paidAt := time.Date(2026, 5, 23, 12, 7, 47, 0, time.UTC)
	payment, err := NewManualPayment(PaymentDirectionBalanceReturn, "", 1500, paidAt)
	if err != nil {
		t.Fatalf("NewManualPayment: %v", err)
	}
	if payment.Direction != PaymentDirectionBalanceReturn {
		t.Fatalf("direction = %q, want balance_return", payment.Direction)
	}
	if payment.Amount != 1500 || payment.Total != 1500 || payment.Commission != 0 {
		t.Fatalf("unexpected amounts: amount=%v total=%v commission=%v", payment.Amount, payment.Total, payment.Commission)
	}
	if payment.ReceiverCard != "" {
		t.Fatalf("receiverCard = %q, want empty", payment.ReceiverCard)
	}
}

func TestPaymentDirectionIsCredit(t *testing.T) {
	if !PaymentDirectionIncoming.IsCredit() {
		t.Fatal("incoming must be credit")
	}
	if !PaymentDirectionBalanceReturn.IsCredit() {
		t.Fatal("balance_return must be credit")
	}
	if PaymentDirectionOutgoing.IsCredit() {
		t.Fatal("outgoing must not be credit")
	}
}

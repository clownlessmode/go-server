package proxy

import "testing"

func TestParseBeelinePaymentBodyIgnoresSpuriousTotalAmount(t *testing.T) {
	body := []byte(`{
		"commission": 97.5,
		"totalAmount": 6097.5
	}`)

	snapshot := parseBeelinePaymentBody(body)
	if snapshot.Amount == nil {
		t.Fatal("amount is nil")
	}
	if *snapshot.Amount != 1500 {
		t.Fatalf("amount = %.2f, want 1500", *snapshot.Amount)
	}
}

func TestParseBeelinePaymentBodyUsesTotalAmountWhenConsistent(t *testing.T) {
	body := []byte(`{
		"commission": 97.5,
		"totalAmount": 1597.5
	}`)

	snapshot := parseBeelinePaymentBody(body)
	if snapshot.Amount == nil {
		t.Fatal("amount is nil")
	}
	if *snapshot.Amount != 1500 {
		t.Fatalf("amount = %.2f, want 1500", *snapshot.Amount)
	}
}

func TestParseBeelinePaymentBodyPrefersNestedData(t *testing.T) {
	body := []byte(`{
		"amount": 6000,
		"data": {
			"amount": 1500,
			"commission": 97.5
		}
	}`)

	snapshot := parseBeelinePaymentBody(body)
	if snapshot.Amount == nil {
		t.Fatal("amount is nil")
	}
	if *snapshot.Amount != 1500 {
		t.Fatalf("amount = %.2f, want 1500", *snapshot.Amount)
	}
}

func TestReconcileBeelinePaymentAmountsFixesMismatch(t *testing.T) {
	amount := 6000.0
	commission := 97.5

	snapshot := reconcileBeelinePaymentAmounts(beelinePaymentSnapshot{
		Amount:     &amount,
		Commission: &commission,
	})

	if snapshot.Amount == nil {
		t.Fatal("amount is nil")
	}
	if *snapshot.Amount != 1500 {
		t.Fatalf("amount = %.2f, want 1500", *snapshot.Amount)
	}
}

func TestBeelinePaymentSnapshotFinalizeTotal(t *testing.T) {
	amount := 6000.0
	commission := 97.5

	snapshot := (&beelinePaymentSnapshot{
		Amount:     &amount,
		Commission: &commission,
	}).finalize()

	total := snapshot.totalAmount()
	if total == nil {
		t.Fatal("total is nil")
	}
	if *total != 1597.5 {
		t.Fatalf("total = %.2f, want 1597.5", *total)
	}
}

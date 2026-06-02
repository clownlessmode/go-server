package domain

import "testing"

func TestDisplayBalanceFromAPIUsesLiveBalance(t *testing.T) {
	api := 61.92
	display := DisplayBalanceFromAPI(&api, 0, 0)
	if display == nil || *display != 61.92 {
		t.Fatalf("display = %v, want 61.92", display)
	}

	outgoing := 9052.5
	display = DisplayBalanceFromAPI(&api, outgoing, 0)
	if display == nil || *display != RoundMoney(api-outgoing) {
		t.Fatalf("display = %v, want %.2f", display, RoundMoney(api-outgoing))
	}
}

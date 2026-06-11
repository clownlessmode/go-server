package domain

import "testing"

func TestDisplayBalanceFromAPIUsesLiveBalance(t *testing.T) {
	api := 61.92
	display := DisplayBalanceFromAPI(&api, 0, 0, 0)
	if display == nil || *display != 61.92 {
		t.Fatalf("display = %v, want 61.92", display)
	}

	outgoing := 9052.5
	display = DisplayBalanceFromAPI(&api, outgoing, 0, 0)
	if display == nil || *display != RoundMoney(api-outgoing) {
		t.Fatalf("display = %v, want %.2f", display, RoundMoney(api-outgoing))
	}

	hiddenDebit := -13845.0
	display = DisplayBalanceFromAPI(&api, 0, 0, hiddenDebit)
	if display == nil || *display != RoundMoney(api-hiddenDebit) {
		t.Fatalf("display with hidden debit = %v, want %.2f", display, RoundMoney(api-hiddenDebit))
	}
}

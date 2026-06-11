package domain

// DisplayBalanceFromAPI returns the balance shown in the app:
// live Beeline balance minus configured outgoing payments, plus incoming,
// minus the net change of hidden real Beeline transactions.
func DisplayBalanceFromAPI(apiBalance *float64, outgoingTotal, incomingTotal, hiddenNetChange float64) *float64 {
	if apiBalance == nil {
		return nil
	}

	value := RoundMoney(*apiBalance - outgoingTotal + incomingTotal - hiddenNetChange)
	return &value
}

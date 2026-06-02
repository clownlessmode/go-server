package domain

func DisplayBalanceFromAPI(apiBalance *float64, outgoingTotal, incomingTotal float64) *float64 {
	if apiBalance == nil {
		return nil
	}

	value := RoundMoney(*apiBalance - outgoingTotal + incomingTotal)
	return &value
}

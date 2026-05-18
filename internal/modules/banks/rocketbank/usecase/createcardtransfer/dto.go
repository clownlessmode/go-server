package createcardtransfer

type Input struct {
	Amount              float64
	BalanceBefore       float64
	Direction           string
	Time                string
	BankID              string
	RecipientCardNumber string
}

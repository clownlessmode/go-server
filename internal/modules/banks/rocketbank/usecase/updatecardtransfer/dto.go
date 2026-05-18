package updatecardtransfer

type Input struct {
	ID                  string
	Amount              *float64
	BalanceBefore       *float64
	Direction           *string
	Time                *string
	BankID              *string
	RecipientCardNumber *string
}

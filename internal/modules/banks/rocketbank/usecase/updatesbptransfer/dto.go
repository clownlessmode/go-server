package updatesbptransfer

type Input struct {
	ID                  string
	Amount              *float64
	BalanceBefore       *float64
	Direction           *string
	Time                *string
	OperationFirstName  *string
	OperationMiddleName *string
	OperationLastName   *string
	BankID              *string
	PhoneNumber         *string
}

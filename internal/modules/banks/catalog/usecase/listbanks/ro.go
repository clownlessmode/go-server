package listbanks

import "time"

type Output struct {
	Banks []BankOutput
}

type BankOutput struct {
	ID        int64
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

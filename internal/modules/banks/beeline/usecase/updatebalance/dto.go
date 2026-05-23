package updatebalance

import "time"

type Input struct {
	Number  string
	Balance *float64
}

type Output struct {
	Number        string
	Balance       *float64
	BaseBalance   *float64
	PaymentsTotal float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

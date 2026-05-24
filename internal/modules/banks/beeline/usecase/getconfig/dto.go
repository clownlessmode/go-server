package getconfig

import "time"

type Input struct {
	Number string
}

type Output struct {
	Number        string
	Balance       *float64
	PaymentsTotal float64
	IncomingTotal float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

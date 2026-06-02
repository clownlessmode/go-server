package domain

import "time"

type DetalizationSnapshot struct {
	SimNumber       string
	PeriodStart     time.Time
	PeriodEnd       time.Time
	Data            []byte
	ComputedBalance *float64
	APIBalance      *float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

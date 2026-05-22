package domain

import "time"

const BankID int64 = 2

type Config struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

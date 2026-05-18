package domain

import "time"

const BankID int64 = 1

type Config struct {
	Balance    *float64
	ClientInfo ClientInfo
	History    []HistoryItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ClientInfo struct {
	FirstName   *string
	MiddleName  *string
	LastName    *string
	PhoneNumber *string
	CardNumber  *string
}

package grantaccess

import "time"

type Input struct {
	UserID      int64
	BankID      int64
	ExpiresAt   time.Time
	GrantReason string
}

package revokeaccess

import "time"

type Output struct {
	ID           int64
	UserID       int64
	BankID       int64
	BankCode     string
	BankName     string
	GrantedAt    time.Time
	ExpiresAt    time.Time
	GrantReason  string
	RevokedAt    *time.Time
	RevokeReason *string
	IsActive     bool
}

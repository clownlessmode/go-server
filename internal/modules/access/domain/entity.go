package domain

import "time"

type AccessGrant struct {
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
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (a *AccessGrant) IsActive(now time.Time) bool {
	return a.RevokedAt == nil && now.Before(a.ExpiresAt)
}

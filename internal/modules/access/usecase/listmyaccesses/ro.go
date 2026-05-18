package listmyaccesses

import "time"

type Output struct {
	Accesses []AccessOutput
}

type AccessOutput struct {
	ID           int64
	UserID       int64
	BankID       int64
	BankCode     string
	BankName     string
	GrantedAt    time.Time
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	RevokeReason *string
	IsActive     bool
}

package http

import "time"

type AccessResponse struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"userId"`
	BankID       int64      `json:"bankId"`
	BankCode     string     `json:"bankCode"`
	BankName     string     `json:"bankName"`
	GrantedAt    time.Time  `json:"grantedAt"`
	ExpiresAt    time.Time  `json:"expiresAt"`
	GrantReason  *string    `json:"grantReason,omitempty"`
	RevokedAt    *time.Time `json:"revokedAt,omitempty"`
	RevokeReason *string    `json:"revokeReason,omitempty"`
	IsActive     bool       `json:"isActive"`
}

type AccessErrorResponse struct {
	Error string `json:"error"`
}

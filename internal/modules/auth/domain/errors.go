package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("inactive user")
	ErrInvalidToken       = errors.New("invalid token")
	ErrRefreshRevoked     = errors.New("refresh token revoked")
)

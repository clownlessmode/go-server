package domain

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidRole  = errors.New("invalid user role")
)

package domain

import "errors"

var (
	ErrUnsupportedBank = errors.New("unsupported sms bank")
	ErrInvalidMessage  = errors.New("invalid sms message")
)

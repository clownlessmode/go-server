package domain

import "errors"

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrMessageNotFound = errors.New("message not found")
)

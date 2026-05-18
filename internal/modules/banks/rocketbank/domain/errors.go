package domain

import "errors"

var (
	ErrHistoryItemNotFound = errors.New("rocketbank history item not found")
	ErrHistoryItemExists   = errors.New("rocketbank history item already exists")
	ErrInsufficientBalance = errors.New("rocketbank insufficient balance")
)

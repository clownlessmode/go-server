package domain

import "errors"

var (
	ErrAccessNotFound      = errors.New("access not found")
	ErrAccessAlreadyExists = errors.New("access already exists")
	ErrInvalidExpiration   = errors.New("invalid access expiration")
)

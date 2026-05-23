package domain

import "errors"

var (
	ErrInvalidReceiverCard  = errors.New("invalid receiver card format")
	ErrPaymentAmountTooLow  = errors.New("payment amount is below minimum")
	ErrInvalidPaymentTime   = errors.New("invalid payment time")
	ErrInvalidPayment           = errors.New("invalid payment")
	ErrInvalidPaymentDirection  = errors.New("invalid payment direction")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidSimNumber     = errors.New("invalid sim number format")
	ErrSimNotFound          = errors.New("sim not found")
	ErrSimAlreadyExists              = errors.New("sim already exists")
	ErrDetalizationSnapshotNotFound  = errors.New("detalization snapshot not found")
)

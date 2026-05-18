package http

import "time"

type ConfigResponse struct {
	Balance    *float64           `json:"balance"`
	ClientInfo ClientInfoResponse `json:"clientInfo"`
	History    []HistoryResponse  `json:"history"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
}

type ClientInfoResponse struct {
	FirstName   *string `json:"firstName"`
	MiddleName  *string `json:"middleName"`
	LastName    *string `json:"lastName"`
	PhoneNumber *string `json:"phoneNumber"`
	CardNumber  *string `json:"cardNumber"`
}

type HistoryResponse struct {
	ID                  string  `json:"id"`
	Type                string  `json:"type"`
	Amount              float64 `json:"amount"`
	BalanceBefore       float64 `json:"balanceBefore"`
	Direction           string  `json:"direction"`
	Time                string  `json:"time"`
	OperationFirstName  string  `json:"operationFirstName,omitempty"`
	OperationMiddleName string  `json:"operationMiddleName,omitempty"`
	OperationLastName   string  `json:"operationLastName,omitempty"`
	BankID              string  `json:"bankId,omitempty"`
	PhoneNumber         string  `json:"phoneNumber,omitempty"`
	RecipientCardNumber string  `json:"recipientCardNumber,omitempty"`
}

type RocketbankErrorResponse struct {
	Error string `json:"error"`
}

package domain

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusFailed    MessageStatus = "failed"
)

type OutboundMessage struct {
	ID           string
	Address      string
	Body         string
	Bank         string
	Status       MessageStatus
	DeviceID     *string
	ErrorMessage *string
	CreatedAt    time.Time
	DeliveredAt  *time.Time
}

func NewOutboundMessage(address, body, bank string) (OutboundMessage, error) {
	address = strings.TrimSpace(address)
	body = strings.TrimSpace(body)
	if address == "" || body == "" {
		return OutboundMessage{}, ErrInvalidMessage
	}

	return OutboundMessage{
		ID:      newMessageID(),
		Address: address,
		Body:    body,
		Bank:    strings.TrimSpace(bank),
		Status:  MessageStatusPending,
	}, nil
}

func newMessageID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])

	return fmt.Sprintf("%x", b[:])
}

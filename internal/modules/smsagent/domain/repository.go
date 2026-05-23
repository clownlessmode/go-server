package domain

import "context"

type Repository interface {
	Enqueue(ctx context.Context, message OutboundMessage) (OutboundMessage, error)
	ListPending(ctx context.Context, limit int) ([]OutboundMessage, error)
	Ack(ctx context.Context, id string, status MessageStatus, deviceID, errorMessage string) error
}

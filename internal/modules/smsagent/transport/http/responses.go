package http

import (
	"time"

	agentdomain "project/internal/modules/smsagent/domain"
)

type MessageResponse struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Body      string    `json:"body"`
	Bank      string    `json:"bank,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type AckMessageRequest struct {
	Status       string `json:"status" binding:"required" example:"delivered"`
	DeviceID     string `json:"deviceId,omitempty" example:"abc123"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type AgentErrorResponse struct {
	Error string `json:"error"`
}

func messageResponses(messages []agentdomain.OutboundMessage) []MessageResponse {
	result := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		result = append(result, MessageResponse{
			ID:        message.ID,
			Address:   message.Address,
			Body:      message.Body,
			Bank:      message.Bank,
			CreatedAt: message.CreatedAt,
		})
	}

	return result
}

package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"project/internal/app/logger"
	agentdomain "project/internal/modules/smsagent/domain"
	"project/internal/modules/smsagent/usecase/ackmessage"
	"project/internal/modules/smsagent/usecase/listpending"
)

var handlerLog = logger.New("sms-agent-http")

type Handler struct {
	listPending *listpending.UseCase
	ackMessage  *ackmessage.UseCase
}

func NewHandler(listPending *listpending.UseCase, ackMessage *ackmessage.UseCase) *Handler {
	return &Handler{
		listPending: listPending,
		ackMessage:  ackMessage,
	}
}

// ListPendingMessages godoc
// @Summary Poll pending SMS for mobile agent
// @Description Returns queued SMS messages for the Android agent to deliver locally via Shizuku.
// @Tags sms agent
// @Produce json
// @Param limit query int false "Max messages" default(10)
// @Security SMSAgentKey
// @Success 200 {array} MessageResponse
// @Failure 401 {object} AgentErrorResponse
// @Failure 500 {object} AgentErrorResponse
// @Router /sms-agent/v1/messages [get]
func (h *Handler) ListPendingMessages(c *gin.Context) {
	limit := 10
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		limit = parsed
	}

	out, err := h.listPending.Execute(c.Request.Context(), listpending.Input{Limit: limit})
	if err != nil {
		handlerLog.Errorf("list pending sms failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, messageResponses(out.Messages))
}

// AckMessage godoc
// @Summary Acknowledge SMS delivery
// @Description Marks a queued SMS as delivered or failed after the agent attempts local injection.
// @Tags sms agent
// @Accept json
// @Produce json
// @Param id path string true "Message ID"
// @Param input body AckMessageRequest true "Delivery result"
// @Security SMSAgentKey
// @Success 204
// @Failure 400 {object} AgentErrorResponse
// @Failure 401 {object} AgentErrorResponse
// @Failure 404 {object} AgentErrorResponse
// @Failure 500 {object} AgentErrorResponse
// @Router /sms-agent/v1/messages/{id}/ack [post]
func (h *Handler) AckMessage(c *gin.Context) {
	var req AckMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	status := agentdomain.MessageStatus(req.Status)
	switch status {
	case agentdomain.MessageStatusDelivered, agentdomain.MessageStatusFailed:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be delivered or failed"})
		return
	}

	err := h.ackMessage.Execute(c.Request.Context(), ackmessage.Input{
		ID:           c.Param("id"),
		Status:       status,
		DeviceID:     req.DeviceID,
		ErrorMessage: req.ErrorMessage,
	})
	if err != nil {
		if errors.Is(err, agentdomain.ErrMessageNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
			return
		}
		handlerLog.Errorf("ack sms failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

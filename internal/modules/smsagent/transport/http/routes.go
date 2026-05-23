package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *Handler, apiKey string) {
	group := router.Group("/sms-agent/v1", AgentAPIKey(apiKey))
	{
		group.GET("/messages", handler.ListPendingMessages)
		group.POST("/messages/:id/ack", handler.AckMessage)
	}
}

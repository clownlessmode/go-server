package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *Handler, middlewares ...gin.HandlerFunc) {
	rocketbank := router.Group("/banks/rocketbank", middlewares...)
	{
		rocketbank.GET("/config", handler.GetConfig)
		rocketbank.PATCH("/config/balance", handler.UpdateBalance)
		rocketbank.PATCH("/config/client-info", handler.UpdateClientInfo)
		rocketbank.GET("/config/history", handler.ListHistory)
		rocketbank.DELETE("/config/history", handler.ClearHistory)
		rocketbank.POST("/config/history/card-transfer", handler.CreateCardTransfer)
		rocketbank.POST("/config/history/cash-transfer", handler.CreateCashTransfer)
		rocketbank.POST("/config/history/sbp-transfer", handler.CreateSBPTransfer)
		rocketbank.GET("/config/history/items/:id", handler.GetHistoryItem)
		rocketbank.PATCH("/config/history/items/:id/card-transfer", handler.UpdateCardTransfer)
		rocketbank.PATCH("/config/history/items/:id/cash-transfer", handler.UpdateCashTransfer)
		rocketbank.PATCH("/config/history/items/:id/sbp-transfer", handler.UpdateSBPTransfer)
		rocketbank.DELETE("/config/history/items/:id", handler.DeleteHistoryItem)
	}
}

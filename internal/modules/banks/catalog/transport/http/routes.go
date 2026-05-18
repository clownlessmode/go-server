package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *Handler, middlewares ...gin.HandlerFunc) {
	banks := router.Group("/banks", middlewares...)
	{
		banks.GET("", handler.ListBanks)
	}
}

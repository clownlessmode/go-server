package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *Handler, middlewares ...gin.HandlerFunc) {
	beeline := router.Group("/banks/beeline", middlewares...)
	{
		beeline.GET("/config", handler.GetConfig)
	}
}

package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *Handler) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
	}
}

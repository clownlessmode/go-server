package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *Handler, middlewares ...gin.HandlerFunc) {
	users := router.Group("/users", middlewares...)
	{
		users.POST("", handler.CreateUser)
		users.GET("", handler.ListUsers)
		users.GET("/:id", handler.GetUser)
		users.PUT("/:id", handler.UpdateUser)
		users.DELETE("/:id", handler.DeleteUser)
	}
}

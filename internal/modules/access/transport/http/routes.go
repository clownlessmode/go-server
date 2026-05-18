package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *Handler, authMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	accesses := router.Group("/accesses", authMiddleware)
	{
		accesses.GET("/me", handler.ListMyAccesses)
		accesses.POST("/grant", adminMiddleware, handler.GrantAccess)
		accesses.POST("/revoke", adminMiddleware, handler.RevokeAccess)
	}
}

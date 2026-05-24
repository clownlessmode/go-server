package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *Handler, middlewares ...gin.HandlerFunc) {
	beeline := router.Group("/banks/beeline", middlewares...)
	{
		beeline.GET("/sims", handler.ListSims)
		beeline.POST("/sims", handler.CreateSim)
		beeline.GET("/sims/:number", handler.GetSim)
		beeline.DELETE("/sims/:number", handler.DeleteSim)

		beeline.GET("/sims/:number/config", handler.GetConfig)

		beeline.GET("/sims/:number/detalization", handler.GetDetalization)
		beeline.DELETE("/sims/:number/detalization/transactions/:id", handler.HideDetalizationTransaction)

		beeline.GET("/sims/:number/payments", handler.ListPayments)
		beeline.POST("/sims/:number/payments", handler.CreatePayment)
		beeline.GET("/sims/:number/payments/:id", handler.GetPayment)
		beeline.PATCH("/sims/:number/payments/:id", handler.UpdatePayment)
		beeline.DELETE("/sims/:number/payments/:id", handler.DeletePayment)
	}
}

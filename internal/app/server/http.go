package server

import "github.com/gin-gonic/gin"

func NewHTTPServer() *gin.Engine {
	router := gin.Default()
	return router
}

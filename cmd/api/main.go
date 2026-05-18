package main

import (
	"project/internal/app/bootstrap"
	"project/internal/app/logger"
)

var apiLog = logger.New("api")

// @title Rebellion Banking APII
// @version 1.0
// @description API documentation for .
// @host localhost:8080
// @BasePath /
func main() {
	app := bootstrap.NewApp()
	defer app.Close()

	apiLog.Successf("server started on :8080")
	if err := app.Router.Run(":8080"); err != nil {
		apiLog.Fatalf("%v", err)
	}
}

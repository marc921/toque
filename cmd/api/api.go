package main

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"marcbrun.io/toque/pkg"
	"marcbrun.io/toque/pkg/api"
)

func main() {
	logger := pkg.NewLogger("production", "api")

	logger.Info("Starting...")
	defer logger.Info("Shutting down.")

	// Create a new Echo instance
	e := echo.New()
	e.Use(middleware.Logger())
	e.Validator = api.NewCustomValidator()

	// Create RabbitMQ publisher
	publisher, err := pkg.NewRabbitMQClient()
	if err != nil {
		logger.Fatal("failed to create RabbitMQ publisher", zap.Error(err))
	}
	defer publisher.Close()

	// Create a new handler
	handler := api.NewHandler(publisher)

	// Define a route handler
	e.POST("/", handler.Echo)
	e.POST("/test", handler.StressTest)

	go func() {
		pkg.OnSignal(func() {
			logger.Info("Shutting down the server...")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			e.Shutdown(shutdownCtx)
		})
	}()

	// Start the server
	err = e.Start(":8080")
	if err != nil {
		logger.Fatal("e.Start", zap.Error(err))
	}
}

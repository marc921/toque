package main

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"marcbrun.io/toque/pkg"
	"marcbrun.io/toque/pkg/api"
)

func main() {
	// Create a new Echo instance
	e := echo.New()
	e.Use(middleware.Logger())
	e.Validator = api.NewCustomValidator()

	// Create RabbitMQ publisher
	publisher, err := pkg.NewRabbitMQClient()
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("failed to create RabbitMQ publisher: %v", err))
	}
	defer publisher.Close()

	// Create a new handler
	handler := api.NewHandler(publisher)

	// Define a route handler
	e.POST("/", handler.Echo)
	e.POST("/test", handler.StressTest)

	go func() {
		pkg.OnSignal(func() {
			e.Logger.Info("Shutting down the server")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			e.Shutdown(shutdownCtx)
		})
	}()

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}

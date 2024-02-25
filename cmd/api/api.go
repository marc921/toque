package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"marcbrun.io/kubernetes-api/pkg"
)

func main() {
	// Create a new Echo instance
	e := echo.New()

	// Create RabbitMQ publisher
	publisher, err := pkg.NewRabbitMQClient()
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("failed to create RabbitMQ publisher: %v", err))
	}
	defer publisher.Close()

	// Define a route handler
	e.GET("/", func(c echo.Context) error {
		err = publisher.Publish(c.Request().Context(), "Hello, World!")
		if err != nil {
			return fmt.Errorf("failed to publish a message: %v", err)
		}
		return c.String(http.StatusOK, "Hello, World!")
	})

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

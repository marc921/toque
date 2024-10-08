package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"marcbrun.io/toque/pkg"
	"marcbrun.io/toque/pkg/messagebroker"
)

func main() {
	// Connect to the database.
	// connString, ok := os.LookupEnv("DATABASE_URL")
	// if !ok {
	// 	log.Fatal("DATABASE_URL environment variable is not set")
	// }

	// dbConn, err := pkg.NewPostgresConnection(ctx, connString)
	// if err != nil {
	// 	log.Fatal(fmt.Errorf("pkg.NewPostgresConnection: %w", err))
	// }
	// defer dbConn.Close(ctx)

	// queries := sqlcgen.New(dbConn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pkg.OnSignal(cancel)

	config, err := ParseConfig(ctx)
	if err != nil {
		zap.L().Fatal("ParseConfig", zap.Error(err))
	}

	logger := pkg.NewLogger(config.Env, "worker")

	logger.Info("Starting...")
	defer logger.Info("Shutting down.")

	errGrp, ctx := errgroup.WithContext(ctx)

	// Start the RabbitMQ consumer
	errGrp.Go(func() error {
		msgConsumer, err := messagebroker.NewRabbitMQConsumer(
			ctx,
			logger.With(zap.String("component", "RabbitMQConsumer")),
			"worker-input",
			"worker",
			config.RabbitMQ.URL,
		)
		if err != nil {
			return fmt.Errorf("failed to create RabbitMQ consumer: %w", err)
		}
		defer msgConsumer.Close()

		err = msgConsumer.Consume(
			ctx,
			func(msg amqp091.Delivery) error {
				// TODO: process the message
				return nil
			},
		)
		if err != nil {
			return fmt.Errorf("msgProcessor.Start: %w", err)
		}
		return nil
	})

	// Create an echo server for metrics
	e := echo.New()
	e.Use(middleware.Logger())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Start the metrics HTTP server
	errGrp.Go(func() error {
		logger.Info("starting metrics server", zap.String("address", ":9000"))
		err = e.Start(":9000")
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				// Graceful http server shutdown
				return nil
			}
			return fmt.Errorf("e.Start: %w", err)
		}
		return nil
	})

	// Graceful shutdown on context cancellation
	errGrp.Go(func() error {
		// From errgroup.WithContext, ctx is canceled the first time a function passed to errGrp.Go returns a non-nil error
		<-ctx.Done()
		logger.Info("context canceled, shutting down the server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		//nolint: contextcheck	// we do want a new, non-inherited context here since ctx is done
		err := e.Shutdown(shutdownCtx)
		if err != nil {
			return fmt.Errorf("e.Shutdown: %w", err)
		}
		return nil
	})

	err = errGrp.Wait()
	if err != nil {
		logger.Error("errGrp.Wait", zap.Error(err))
	}
}

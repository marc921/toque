package api

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/labstack/echo/v4"
	"marcbrun.io/toque/pkg/messagebroker"
)

type Handler struct {
	publisher messagebroker.Publisher
}

func NewHandler(publisher messagebroker.Publisher) *Handler {
	return &Handler{publisher: publisher}
}

type Request struct {
	Message string `json:"message" validate:"required"`
	Times   int    `json:"times" validate:"required"`
}

type Response struct {
	Message string `json:"message"`
}

func (h *Handler) Echo(c echo.Context) error {
	c.Logger().Info("Received request")
	req := new(Request)
	if err := c.Bind(req); err != nil {
		return fmt.Errorf("failed to bind a request: %w", err)
	}
	if err := c.Validate(req); err != nil {
		return fmt.Errorf("failed to validate a request: %w", err)
	}

	err := h.publisher.Publish(c.Request().Context(), req.Message)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return c.JSON(http.StatusOK, Response{Message: req.Message})
}

func (h *Handler) StressTest(c echo.Context) error {
	req := new(Request)
	if err := c.Bind(req); err != nil {
		return fmt.Errorf("failed to bind a request: %w", err)
	}
	if err := c.Validate(req); err != nil {
		return fmt.Errorf("failed to validate a request: %w", err)
	}

	errGroup, ctx := errgroup.WithContext(c.Request().Context())
	errGroup.SetLimit(500)
	start := time.Now()

	for i := 0; i < req.Times; i++ {
		i := i
		errGroup.Go(func() error {
			err := h.publisher.Publish(ctx, fmt.Sprintf("%s %d", req.Message, i))
			if err != nil {
				return fmt.Errorf("failed to publish a message: %v", err)
			}
			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("failed to publish messages: %v", err)
	}

	after := time.Since(start)
	return c.JSON(http.StatusOK, Response{Message: fmt.Sprintf("Published %d messages in %v", req.Times, after)})
}

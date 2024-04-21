package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"marcbrun.io/toque/pkg"
)

type Handler struct {
	publisher pkg.Publisher
}

func NewHandler(publisher pkg.Publisher) *Handler {
	return &Handler{publisher: publisher}
}

type Request struct {
	Message string `json:"message" validate:"required"`
}

type Response struct {
	Message string `json:"message"`
}

func (h *Handler) Echo(c echo.Context) error {
	req := new(Request)
	if err := c.Bind(req); err != nil {
		return fmt.Errorf("failed to bind a request: %v", err)
	}
	if err := c.Validate(req); err != nil {
		return fmt.Errorf("failed to validate a request: %v", err)
	}

	err := h.publisher.Publish(c.Request().Context(), req.Message)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}

	return c.JSON(http.StatusOK, Response{Message: req.Message})
}

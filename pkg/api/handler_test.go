package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"marcbrun.io/toque/pkg"
)

func TestHelloHandler(t *testing.T) {
	e := echo.New()
	e.Validator = NewCustomValidator()

	body := Request{Message: "Hi there!"}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := NewHandler(&pkg.PublisherMock{})

	require.NoError(t, handler.Echo(c))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get(echo.HeaderContentType))

	var response Response
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	require.Equal(t, "Hi there!", response.Message)
}

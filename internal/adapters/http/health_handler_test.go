package http_test

import (
	"encoding/json"
	"financial-system-pro/internal/adapters/http"
	"financial-system-pro/internal/application/services"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestHandler local para testes deste pacote seguindo Clean Arch
func setupTestHandler() (*http.Handler, *fiber.App) {
	logger := zap.NewNop()

	userService := &services.UserService{Logger: logger}
	authService := &services.AuthService{Logger: logger}
	transactionService := &services.TransactionService{Logger: logger}

	h := http.NewHandlerForTesting(
		userService,
		authService,
		transactionService,
		&services.TronService{},
		nil,
		http.NewZapLoggerAdapter(logger),
		http.NewRateLimiter(logger),
		nil,
	)

	app := fiber.New()
	return h, app
}

func TestLivenessProbe(t *testing.T) {
	h, app := setupTestHandler()
	app.Get("/live", h.LivenessProbe)

	req := httptest.NewRequest("GET", "/live", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]bool
	_ = json.NewDecoder(resp.Body).Decode(&body)
	assert.True(t, body["alive"])
}

func TestReadinessProbe(t *testing.T) {
	h, app := setupTestHandler()
	app.Get("/ready", h.ReadinessProbe)

	req := httptest.NewRequest("GET", "/ready", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHealthCheckFull_Degraded(t *testing.T) {
	logger := zap.NewNop()
	// userService nil forces degraded + database down path
	h := http.NewHandlerForTesting(
		nil,
		&services.AuthService{Logger: logger},
		&services.TransactionService{Logger: logger},
		&services.TronService{},
		nil,
		http.NewZapLoggerAdapter(logger),
		http.NewRateLimiter(logger),
		nil,
	)
	app := fiber.New()
	app.Get("/health/full", h.HealthCheckFull)
	req := httptest.NewRequest("GET", "/health/full", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	// Expect degraded -> 503
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

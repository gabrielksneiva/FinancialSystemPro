package http_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLivenessProbe testa probe de liveness
func TestLivenessProbe(t *testing.T) {
	handler, app, _, _ := setupTestHandler()

	app.Get("/live", handler.LivenessProbe)

	req := httptest.NewRequest("GET", "/live", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]bool
	_ = json.NewDecoder(resp.Body).Decode(&response)
	assert.True(t, response["alive"])
}

// TestReadinessProbe testa probe de readiness
func TestReadinessProbe(t *testing.T) {
	// Services are mocked and not nil, so should return 200
	handler, app, _, _ := setupTestHandler()

	app.Get("/ready", handler.ReadinessProbe)

	req := httptest.NewRequest("GET", "/ready", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should return 200 when services are available
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

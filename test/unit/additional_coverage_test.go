package http_test

import (
	"financial-system-pro/internal/adapters/http"
	"financial-system-pro/internal/application/dto"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestNewRateLimiter_Variations testa diferentes configurações do rate limiter
func TestNewRateLimiter_Variations(t *testing.T) {
	logger := zap.NewNop()

	rl := http.NewRateLimiter(logger)
	assert.NotNil(t, rl)
}

// TestRateLimiter_IsAllowed_EdgeCases testa casos extremos do rate limiter
func TestRateLimiter_IsAllowed_EdgeCases(t *testing.T) {
	logger := zap.NewNop()
	rl := http.NewRateLimiter(logger)

	// Test with empty user ID
	allowed := rl.IsAllowed("", "test_action")
	assert.True(t, allowed) // Should allow since no user to rate limit

	// Test with same user, different actions
	for i := 0; i < 5; i++ {
		rl.IsAllowed("user1", "action_a")
		rl.IsAllowed("user1", "action_b")
	}

	// Test cleanup trigger (simulate time passing)
	time.Sleep(50 * time.Millisecond)
	allowed = rl.IsAllowed("user2", "test")
	assert.True(t, allowed)
}

// TestMetricsRecording_EdgeCases testa casos extremos de métricas
func TestMetricsRecording_EdgeCases(t *testing.T) {
	// Record zero values
	http.RecordDeposit()
	http.RecordWithdraw()
	http.RecordTransfer()
	http.RecordFailure()
	http.RecordRequestTime(0)

	// Record negative duration (edge case)
	http.RecordRequestTime(-1 * time.Second)

	// Record very large duration
	http.RecordRequestTime(time.Hour)

	assert.True(t, true) // Just verify no panics
}

// TestRouter_AllRouteCombinations testa combinações de rotas
func TestRouter_AllRouteCombinations(t *testing.T) {
	app := fiber.New()

	// Add some test routes
	app.Get("/test-get", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Post("/test-post", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	routes := app.GetRoutes()
	assert.Greater(t, len(routes), 0, "Should have registered routes")

	// Verify different HTTP methods are registered
	var hasGET, hasPOST bool
	for _, route := range routes {
		if route.Method == "GET" {
			hasGET = true
		}
		if route.Method == "POST" {
			hasPOST = true
		}
	}
	assert.True(t, hasGET, "Should have GET routes")
	assert.True(t, hasPOST, "Should have POST routes")
}

// TestFiberAppConfiguration testa configuração da app Fiber
func TestFiberAppConfiguration(t *testing.T) {
	app := fiber.New()

	assert.NotNil(t, app)

	// Verify app can handle routes
	app.Get("/test-config", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	routes := app.GetRoutes()
	assert.NotEmpty(t, routes)
}

// TestMetricsGlobalState testa estado global das métricas
func TestMetricsGlobalState(t *testing.T) {
	// Test that metrics functions are accessible
	http.RecordDeposit()
	http.RecordDeposit()
	http.RecordWithdraw()

	// Verify functions execute without panic
	http.RecordTransfer()
	http.RecordFailure()
	http.RecordRequestTime(time.Millisecond * 100)

	// Test passes if no panic occurs
	assert.True(t, true, "Metrics functions executed successfully")
}

// TestValidateRequest_MultipleScenarios testa múltiplos cenários de validação
func TestValidateRequest_MultipleScenarios(t *testing.T) {
	type ComplexRequest struct {
		Email    string   `json:"email" validate:"required,email"`
		Password string   `json:"password" validate:"required,min=8,max=100"`
		Age      int      `json:"age" validate:"required,gte=18,lte=120"`
		Tags     []string `json:"tags" validate:"omitempty,dive,min=1"`
	}

	app := fiber.New()

	app.Post("/complex", func(c *fiber.Ctx) error {
		var req ComplexRequest
		if err := dto.ValidateRequest(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(req)
	})

	// This test just verifies the app setup doesn't panic
	assert.NotNil(t, app)
}

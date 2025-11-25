package integration

import (
	"net/http/httptest"
	"testing"

	httpAdapter "financial-system-pro/internal/adapters/http"
	"financial-system-pro/internal/shared/breaker"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func TestCircuitBreaker_StatusAndHealth(t *testing.T) {
	logger := zap.NewNop()
	bm := breaker.NewBreakerManager(logger)
	handler := httpAdapter.NewCircuitBreakerHandler(bm)

	app := fiber.New()
	app.Get("/api/circuit-breakers", handler.GetCircuitBreakerStatus)
	app.Get("/api/circuit-breakers/health", handler.GetCircuitBreakerHealth)

	// Initially no breakers -> healthy
	req1 := httptest.NewRequest(fiber.MethodGet, "/api/circuit-breakers", nil)
	resp1, _ := app.Test(req1, -1)
	if resp1.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 status listing got %d", resp1.StatusCode)
	}

	req2 := httptest.NewRequest(fiber.MethodGet, "/api/circuit-breakers/health", nil)
	resp2, _ := app.Test(req2, -1)
	if resp2.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 health got %d", resp2.StatusCode)
	}

	// Create a breaker and force it open by exceeding failures
	br := bm.GetBreaker("test-breaker")
	for i := 0; i < 6; i++ { // Consecutive failures >5 triggers open
		_, _ = br.Execute(func() (interface{}, error) { return nil, ErrForcedFailure })
	}
	// Validate state open
	if br.State().String() != "open" {
		t.Fatalf("expected breaker state open got %s", br.State().String())
	}

	// Health endpoint should now return 503
	req3 := httptest.NewRequest(fiber.MethodGet, "/api/circuit-breakers/health", nil)
	resp3, _ := app.Test(req3, -1)
	if resp3.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected 503 unhealthy got %d", resp3.StatusCode)
	}
}

// ErrForcedFailure simple error value for breaker testing
type forcedErr string

func (e forcedErr) Error() string { return string(e) }

var ErrForcedFailure = forcedErr("forced failure")

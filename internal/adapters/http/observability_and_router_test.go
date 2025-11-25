package http

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"financial-system-pro/internal/shared/breaker"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func TestMetricsEndpointSetup(t *testing.T) {
	app := fiber.New()
	SetupMetricsEndpoint(app)
	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}

func TestCircuitBreakerHandler_StatusAndHealth_HTTP(t *testing.T) {
	bm := breaker.NewBreakerManager(zap.NewNop())
	h := NewCircuitBreakerHandler(bm)
	app := fiber.New()
	app.Get("/api/circuit-breakers", h.GetCircuitBreakerStatus)
	app.Get("/api/circuit-breakers/health", h.GetCircuitBreakerHealth)

	// Create a breaker and force it open
	br := bm.GetBreaker("test")
	for i := 0; i < 6; i++ { // trigger open (consecutive failures >5)
		_, _ = br.Execute(func() (interface{}, error) { return nil, forcedErr("fail") })
	}

	// Status endpoint
	r1 := httptest.NewRequest(fiber.MethodGet, "/api/circuit-breakers", nil)
	resp1, _ := app.Test(r1, -1)
	if resp1.StatusCode != 200 {
		t.Fatalf("expected 200 status got %d", resp1.StatusCode)
	}

	// Health endpoint should be 503
	r2 := httptest.NewRequest(fiber.MethodGet, "/api/circuit-breakers/health", nil)
	resp2, _ := app.Test(r2, -1)
	if resp2.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected 503 unhealthy got %d", resp2.StatusCode)
	}
}

// forcedErr used to open breaker
type forcedErr string

func (e forcedErr) Error() string { return string(e) }

func TestHandler_AuditEndpointsPlaceholders(t *testing.T) {
	h := &Handler{logger: NewZapLoggerAdapter(zap.NewNop())}
	app := fiber.New()
	app.Get("/audit", h.GetAuditLogs)
	app.Get("/audit/stats", h.GetAuditStats)

	r1 := httptest.NewRequest(fiber.MethodGet, "/audit", nil)
	resp1, _ := app.Test(r1, -1)
	if resp1.StatusCode != fiber.StatusNotImplemented {
		t.Fatalf("expected 501 got %d", resp1.StatusCode)
	}
	r2 := httptest.NewRequest(fiber.MethodGet, "/audit/stats", nil)
	resp2, _ := app.Test(r2, -1)
	if resp2.StatusCode != fiber.StatusNotImplemented {
		t.Fatalf("expected 501 got %d", resp2.StatusCode)
	}
}

func TestHealthCheckFull_Degraded(t *testing.T) {
	h := &Handler{logger: NewZapLoggerAdapter(zap.NewNop())}
	app := fiber.New()
	app.Get("/health/full", h.HealthCheckFull)
	r := httptest.NewRequest(fiber.MethodGet, "/health/full", nil)
	resp, _ := app.Test(r, -1)
	if resp.StatusCode != fiber.StatusServiceUnavailable { // degraded due to missing services
		t.Fatalf("expected 503 degraded got %d", resp.StatusCode)
	}
	var payload map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if payload["status"] != "degraded" {
		t.Fatalf("expected status degraded got %v", payload["status"])
	}
}

func TestRouterRegistersBasicEndpoints(t *testing.T) {
	app := fiber.New()
	logger := zap.NewNop()
	bm := breaker.NewBreakerManager(logger)
	router(app, nil, nil, nil, nil, nil, logger, nil, bm)

	// Health endpoint
	r := httptest.NewRequest(fiber.MethodGet, "/health", nil)
	resp, _ := app.Test(r, -1)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 health got %d", resp.StatusCode)
	}
}

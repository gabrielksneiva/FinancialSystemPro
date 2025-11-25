package integration

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	httpAdapter "financial-system-pro/internal/adapters/http"

	"github.com/gofiber/fiber/v2"
)

func TestMetricsEndpoint_ExposesPrometheus(t *testing.T) {
	app := fiber.New()
	httpAdapter.SetupMetricsEndpoint(app)
	req := httptest.NewRequest(fiber.MethodGet, "/metrics", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	content := string(body)
	// Check for some common Prometheus exposition markers
	if !strings.Contains(content, "# HELP") || !strings.Contains(content, "# TYPE") {
		t.Fatalf("metrics body missing HELP/TYPE markers: %s", content[:200])
	}
}

package container

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegisterFiberHealthChecks(t *testing.T) {
	app := fiber.New()
	registerFiberHealthChecks(app)
	req1 := httptest.NewRequest(fiber.MethodGet, "/health", nil)
	resp1, _ := app.Test(req1)
	if resp1.StatusCode != 200 {
		t.Fatalf("health code %d", resp1.StatusCode)
	}
	req2 := httptest.NewRequest(fiber.MethodGet, "/ready", nil)
	resp2, _ := app.Test(req2)
	if resp2.StatusCode != 200 {
		t.Fatalf("ready code %d", resp2.StatusCode)
	}
}

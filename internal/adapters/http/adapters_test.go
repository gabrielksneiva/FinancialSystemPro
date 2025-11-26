package http

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func TestRateLimiterAdapter_Methods(t *testing.T) {
	rl := NewRateLimiter(zap.NewNop())
	adapter := NewRateLimiterAdapter(rl)
	if !adapter.IsAllowed("user", "deposit") {
		t.Fatalf("expected allowed")
	}
	// middleware simply calls next
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user_id", "u1"); return c.Next() })
	app.Get("/x", adapter.Middleware("deposit"), func(c *fiber.Ctx) error { return c.SendStatus(204) })
	resp, _ := app.Test(httptest.NewRequest(fiber.MethodGet, "/x", nil))
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		t.Fatalf("expected 204 got %d", resp.StatusCode)
	}
}

func TestRateLimiterAdapter_Middleware_BlockAfterLimit(t *testing.T) {
	rl := NewRateLimiter(zap.NewNop())
	adapter := NewRateLimiterAdapter(rl)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user_id", "u1"); return c.Next() })
	app.Use(adapter.Middleware("login"))
	app.Get("/login", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	for i := 0; i < 5; i++ {
		resp, _ := app.Test(httptest.NewRequest(fiber.MethodGet, "/login", nil))
		defer resp.Body.Close()
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200 got %d", resp.StatusCode)
		}
	}
	resp, _ := app.Test(httptest.NewRequest(fiber.MethodGet, "/login", nil))
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected 429 got %d", resp.StatusCode)
	}
}

func TestQueueManagerAdapter_Nil(t *testing.T) {
	if NewQueueManagerAdapter(nil) != nil {
		t.Fatalf("expected nil adapter")
	}
}

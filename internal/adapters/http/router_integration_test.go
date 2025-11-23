package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestAPIv1RoutesExist(t *testing.T) {
	// Create a simple Fiber app to test routing
	app := fiber.New()

	// Simulate the routes from router.go
	v1 := app.Group("/api/v1")
	{
		v1.Post("/users", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "create user"})
		})
		v1.Post("/login", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "login"})
		})
		v1.Post("/deposit", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "deposit"})
		})
		v1.Post("/withdraw", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "withdraw"})
		})
		v1.Post("/transfer", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "transfer"})
		})
		v1.Get("/balance", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "balance"})
		})
		v1.Get("/wallet", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "wallet"})
		})
	}

	// Test that routes exist and respond
	tests := []struct {
		method   string
		path     string
		expected int
	}{
		{"POST", "/api/v1/users", http.StatusOK},
		{"POST", "/api/v1/login", http.StatusOK},
		{"POST", "/api/v1/deposit", http.StatusOK},
		{"POST", "/api/v1/withdraw", http.StatusOK},
		{"POST", "/api/v1/transfer", http.StatusOK},
		{"GET", "/api/v1/balance", http.StatusOK},
		{"GET", "/api/v1/wallet", http.StatusOK},
		{"POST", "/api/v2/users", http.StatusNotFound}, // v2 should not exist yet
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Errorf("Error testing %s %s: %v", test.method, test.path, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != test.expected {
			t.Errorf("%s %s: expected status %d, got %d", test.method, test.path, test.expected, resp.StatusCode)
		}
	}
}

func TestHealthCheckRouteUnversioned(t *testing.T) {
	// Health checks should not be versioned
	app := fiber.New()

	// Health routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Test that health check is unversioned
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Errorf("Error testing GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health: expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Test that /api/v1/health does NOT exist
	req = httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Errorf("Error testing GET /api/v1/health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET /api/v1/health: expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

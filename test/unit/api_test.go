package http_test

import (
	"bytes"
	"encoding/json"
	"financial-system-pro/internal/adapters/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func TestNewRateLimiter(t *testing.T) {
	logger := zap.NewNop()
	rl := http.NewRateLimiter(logger)

	if rl == nil {
		t.Fatal("RateLimiter should not be nil")
	}

	// Test that rate limiter is functional by checking IsAllowed
	// Cannot access unexported fields directly
	allowed := rl.IsAllowed("test-user", "deposit")
	if !allowed {
		t.Error("First request should be allowed")
	}
}

func TestRateLimiter_IsAllowed(t *testing.T) {
	logger := zap.NewNop()
	rl := http.NewRateLimiter(logger)

	userID := "test-user-1"
	action := "deposit"

	// First requests should be allowed (limit is 20 for deposit)
	for i := 0; i < 20; i++ {
		if !rl.IsAllowed(userID, action) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 21st should be blocked
	if rl.IsAllowed(userID, action) {
		t.Error("Request 21 should be blocked by rate limiter")
	}
}

func TestRateLimiter_DifferentActions(t *testing.T) {
	logger := zap.NewNop()
	rl := http.NewRateLimiter(logger)

	userID := "test-user-2"

	// Different actions should have independent limits
	if !rl.IsAllowed(userID, "deposit") {
		t.Error("First deposit should be allowed")
	}

	if !rl.IsAllowed(userID, "withdraw") {
		t.Error("First withdraw should be allowed")
	}

	if !rl.IsAllowed(userID, "transfer") {
		t.Error("First transfer should be allowed")
	}
}

func TestRateLimiter_DifferentUsers(t *testing.T) {
	logger := zap.NewNop()
	rl := http.NewRateLimiter(logger)

	action := "login"

	// Different users should have independent limits (5 for login)
	for i := 0; i < 5; i++ {
		if !rl.IsAllowed("user1", action) {
			t.Errorf("Request %d for user1 should be allowed", i+1)
		}
	}

	// User1 should now be blocked
	if rl.IsAllowed("user1", action) {
		t.Error("User1 should be blocked after 5 requests")
	}

	// But user2 should still be allowed
	if !rl.IsAllowed("user2", action) {
		t.Error("User2's first request should be allowed")
	}
}

func TestFiberApp_JSONResponse(t *testing.T) {
	app := fiber.New()

	app.Get("/json", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "success",
			"code":    200,
		})
	})

	req := httptest.NewRequest("GET", "/json", nil)
	resp, _ := app.Test(req)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	if body["message"] != "success" {
		t.Errorf("Expected message 'success', got %v", body["message"])
	}
}

func TestFiberApp_ParseJSON(t *testing.T) {
	app := fiber.New()

	app.Post("/parse", func(c *fiber.Ctx) error {
		var req struct {
			Amount string `json:"amount"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "parse error"})
		}

		return c.JSON(fiber.Map{"received": req.Amount})
	})

	payload := map[string]string{"amount": "10.50"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/parse", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["received"] != "10.50" {
		t.Errorf("Expected '10.50', got %v", result["received"])
	}
}

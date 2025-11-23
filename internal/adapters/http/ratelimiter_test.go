package http

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// helper para executar request com middleware e user_id
func execWithRateLimiter(t *testing.T, rl *RateLimiter, action, userID string, times int) (lastStatus int) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user_id", userID); return c.Next() })
	app.Get("/test", rl.Middleware(action), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	for i := 0; i < times; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		lastStatus = resp.StatusCode
	}
	return
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(zap.NewNop())
	status := execWithRateLimiter(t, rl, "deposit", "user-1", 3)
	assert.Equal(t, fiber.StatusOK, status)
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(zap.NewNop())
	// limite deposit = 20, vamos simular 21
	status := execWithRateLimiter(t, rl, "deposit", "user-2", 21)
	assert.Equal(t, fiber.StatusTooManyRequests, status)
}

func TestRateLimiter_DefaultActionHighLimit(t *testing.T) {
	rl := NewRateLimiter(zap.NewNop())
	status := execWithRateLimiter(t, rl, "unknown_action", "user-3", 50)
	// default limit 100 -> ainda deve permitir
	assert.Equal(t, fiber.StatusOK, status)
}

func TestRateLimiter_UnauthorizedWithoutUserID(t *testing.T) {
	rl := NewRateLimiter(zap.NewNop())
	app := fiber.New()
	app.Get("/test", rl.Middleware("deposit"), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

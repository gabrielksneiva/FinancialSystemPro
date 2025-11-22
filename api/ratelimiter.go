package api

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RateLimiter implementa rate limiting em memória (simple para Railway Free)
// Para produção, usar Redis rate limiting é melhor
type RateLimiter struct {
	requests map[string][]time.Time
	limits   map[string]int // max requests
	windows  map[string]time.Duration
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limits: map[string]int{
			"deposit":  20, // 20 por minuto
			"withdraw": 15, // 15 por minuto
			"transfer": 10, // 10 por minuto
			"login":    5,  // 5 por minuto
		},
		windows: map[string]time.Duration{
			"deposit":  1 * time.Minute,
			"withdraw": 1 * time.Minute,
			"transfer": 1 * time.Minute,
			"login":    1 * time.Minute,
		},
		logger: logger,
	}

	// Cleanup goroutine (remover requests antigas)
	go rl.cleanup()

	return rl
}

// IsAllowed verifica se o user pode fazer a ação
func (rl *RateLimiter) IsAllowed(userID string, action string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, action)
	now := time.Now()

	limit, ok := rl.limits[action]
	if !ok {
		limit = 100 // padrão generoso
	}

	window, ok := rl.windows[action]
	if !ok {
		window = 1 * time.Minute
	}

	// Remover requests antigas
	cutoff := now.Add(-window)
	var recent []time.Time
	for _, t := range rl.requests[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	// Verificar limite
	if len(recent) >= limit {
		rl.logger.Warn("rate limit exceeded",
			zap.String("user_id", userID),
			zap.String("action", action),
			zap.Int("limit", limit),
			zap.Int("requests", len(recent)),
		)
		return false
	}

	// Adicionar novo request
	rl.requests[key] = append(recent, now)
	return true
}

// cleanup remove entries antigas periodicamente
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		for key, times := range rl.requests {
			cutoff := now.Add(-5 * time.Minute)
			var recent []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					recent = append(recent, t)
				}
			}

			if len(recent) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = recent
			}
		}

		rl.mu.Unlock()
	}
}

// RateLimitMiddleware retorna um middleware de rate limiting
func (rl *RateLimiter) Middleware(action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Pega userID do token JWT
		userIDLocal := c.Locals("user_id")
		if userIDLocal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user_id not found in context",
			})
		}

		userID, ok := userIDLocal.(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user_id in context",
			})
		}

		if !rl.IsAllowed(userID, action) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate limit exceeded",
				"message": fmt.Sprintf("Too many %s requests. Try again later.", action),
			})
		}

		return c.Next()
	}
}

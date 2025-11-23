package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// LoggerInterface permite mockar o logger nos testes
type LoggerInterface interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
}

// QueueManagerInterface permite mockar o gerenciador de filas
type QueueManagerInterface interface {
	EnqueueDeposit(ctx context.Context, userID, amount, callbackURL string) (string, error)
	EnqueueWithdraw(ctx context.Context, userID, amount, callbackURL string) (string, error)
	EnqueueTransfer(ctx context.Context, fromUserID, toUserID, amount, callbackURL string) (string, error)
	IsConnected() bool
}

// RateLimiterInterface permite mockar o rate limiter
type RateLimiterInterface interface {
	Middleware(action string) fiber.Handler
	IsAllowed(userID string, action string) bool
}

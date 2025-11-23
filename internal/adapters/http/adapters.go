package http

import (
	"context"

	queue "financial-system-pro/internal/infrastructure/queue"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ZapLoggerAdapter adapta *zap.Logger para LoggerInterface
type ZapLoggerAdapter struct {
	logger *zap.Logger
}

func NewZapLoggerAdapter(logger *zap.Logger) LoggerInterface {
	return &ZapLoggerAdapter{logger: logger}
}

func (z *ZapLoggerAdapter) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

func (z *ZapLoggerAdapter) Warn(msg string, fields ...zap.Field) {
	z.logger.Warn(msg, fields...)
}

func (z *ZapLoggerAdapter) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
}

func (z *ZapLoggerAdapter) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}

// QueueManagerAdapter adapta *queue.QueueManager para QueueManagerInterface
type QueueManagerAdapter struct {
	qm *queue.QueueManager
}

func NewQueueManagerAdapter(qm *queue.QueueManager) QueueManagerInterface {
	if qm == nil {
		return nil
	}
	return &QueueManagerAdapter{qm: qm}
}

func (q *QueueManagerAdapter) EnqueueDeposit(ctx context.Context, userID, amount, callbackURL string) (string, error) {
	return q.qm.EnqueueDeposit(ctx, userID, amount, callbackURL)
}

func (q *QueueManagerAdapter) EnqueueWithdraw(ctx context.Context, userID, amount, callbackURL string) (string, error) {
	return q.qm.EnqueueWithdraw(ctx, userID, amount, callbackURL)
}

func (q *QueueManagerAdapter) EnqueueTransfer(ctx context.Context, fromUserID, toUserID, amount, callbackURL string) (string, error) {
	return q.qm.EnqueueTransfer(ctx, fromUserID, toUserID, amount, callbackURL)
}

func (q *QueueManagerAdapter) IsConnected() bool {
	return q.qm.IsConnected()
}

// RateLimiterAdapter adapta *RateLimiter para RateLimiterInterface
type RateLimiterAdapter struct {
	rl *RateLimiter
}

func NewRateLimiterAdapter(rl *RateLimiter) RateLimiterInterface {
	return &RateLimiterAdapter{rl: rl}
}

func (r *RateLimiterAdapter) Middleware(action string) fiber.Handler {
	return r.rl.Middleware(action)
}

func (r *RateLimiterAdapter) IsAllowed(userID string, action string) bool {
	return r.rl.IsAllowed(userID, action)
}

package http

import (
	infraMetrics "financial-system-pro/internal/infrastructure/metrics"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Wrappers para manter compatibilidade de testes existentes
func RecordDeposit()                    { infraMetrics.RecordDeposit() }
func RecordWithdraw()                   { infraMetrics.RecordWithdraw() }
func RecordTransfer()                   { infraMetrics.RecordTransfer() }
func RecordFailure()                    { infraMetrics.RecordFailure() }
func RecordRequestTime(d time.Duration) { infraMetrics.RecordRequestTime(d) }

// GetMetrics expõe snapshot das métricas agregadas
func GetMetrics(c *fiber.Ctx) error {
	snapshot := infraMetrics.Snapshot(startTime)
	return c.JSON(snapshot)
}

// MetricsMiddleware mede tempo de resposta por requisição
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		infraMetrics.RecordRequestTime(time.Since(start))
		return err
	}
}

// GlobalMetrics compatível com testes existentes, aponta para infraMetrics.GlobalAppMetrics
var GlobalMetrics = infraMetrics.GlobalAppMetrics

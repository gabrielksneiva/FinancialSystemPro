package http

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Metrics coleta métricas da aplicação
type Metrics struct {
	LastUpdated      time.Time
	DepositCount     int64
	WithdrawCount    int64
	TransferCount    int64
	FailureCount     int64
	TotalRequestTime int64
	RequestCount     int64
	mu               sync.RWMutex
}

// GlobalMetrics é a instância global de métricas exportada para testes
var GlobalMetrics = &Metrics{
	LastUpdated: time.Now(),
}

// metrics mantém referência interna
var metrics = GlobalMetrics

// RecordDeposit incrementa contador de depósitos
func RecordDeposit() {
	atomic.AddInt64(&metrics.DepositCount, 1)
	atomic.AddInt64(&metrics.RequestCount, 1)
}

// RecordWithdraw incrementa contador de saques
func RecordWithdraw() {
	atomic.AddInt64(&metrics.WithdrawCount, 1)
	atomic.AddInt64(&metrics.RequestCount, 1)
}

// RecordTransfer incrementa contador de transferências
func RecordTransfer() {
	atomic.AddInt64(&metrics.TransferCount, 1)
	atomic.AddInt64(&metrics.RequestCount, 1)
}

// RecordFailure incrementa contador de falhas
func RecordFailure() {
	atomic.AddInt64(&metrics.FailureCount, 1)
}

// RecordRequestTime registra tempo de requisição
func RecordRequestTime(duration time.Duration) {
	atomic.AddInt64(&metrics.TotalRequestTime, int64(duration.Nanoseconds()))
}

// GetMetrics retorna snapshot das métricas
// @Summary      Métricas da aplicação
// @Description  Retorna estatísticas de uso e performance
// @Tags         System
// @Produce      json
// @Success      200  {object}  domain.MetricsResponse
// @Router       /metrics [get]
func GetMetrics(c *fiber.Ctx) error {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	deposits := atomic.LoadInt64(&metrics.DepositCount)
	withdraws := atomic.LoadInt64(&metrics.WithdrawCount)
	transfers := atomic.LoadInt64(&metrics.TransferCount)
	failures := atomic.LoadInt64(&metrics.FailureCount)
	requests := atomic.LoadInt64(&metrics.RequestCount)
	totalTime := atomic.LoadInt64(&metrics.TotalRequestTime)

	avgResponseTime := float64(0)
	if requests > 0 {
		avgResponseTime = float64(totalTime) / float64(requests) / float64(time.Millisecond)
	}

	response := fiber.Map{
		"transactions": fiber.Map{
			"deposits":  deposits,
			"withdraws": withdraws,
			"transfers": transfers,
			"failures":  failures,
			"total":     deposits + withdraws + transfers,
		},
		"api": fiber.Map{
			"total_requests":       requests,
			"avg_response_time_ms": avgResponseTime,
		},
		"system": fiber.Map{
			"uptime_seconds": int64(time.Since(startTime).Seconds()),
			"memory_mb":      m.Alloc / 1024 / 1024,
			"goroutines":     runtime.NumGoroutine(),
			"gc_runs":        m.NumGC,
		},
	}

	return c.JSON(response)
}

// MetricsMiddleware coleta métricas de cada requisição
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		RecordRequestTime(duration)

		return err
	}
}

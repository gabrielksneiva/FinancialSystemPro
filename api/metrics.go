package api

import (
	"financial-system-pro/domain"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Metrics coleta métricas da aplicação
type Metrics struct {
	// Transações
	DepositCount  int64
	WithdrawCount int64
	TransferCount int64
	FailureCount  int64

	// Performance
	TotalRequestTime int64 // nanosegundos
	RequestCount     int64

	// Sistema
	LastUpdated time.Time
	mu          sync.RWMutex
}

var metrics = &Metrics{
	LastUpdated: time.Now(),
}

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

	response := domain.MetricsResponse{
		Transactions: struct {
			Deposits  int64 `json:"deposits"`
			Withdraws int64 `json:"withdraws"`
			Transfers int64 `json:"transfers"`
			Failures  int64 `json:"failures"`
			Total     int64 `json:"total"`
		}{
			Deposits:  deposits,
			Withdraws: withdraws,
			Transfers: transfers,
			Failures:  failures,
			Total:     deposits + withdraws + transfers,
		},
		API: struct {
			TotalRequests     int64   `json:"total_requests"`
			AvgResponseTimeMs float64 `json:"avg_response_time_ms"`
		}{
			TotalRequests:     requests,
			AvgResponseTimeMs: avgResponseTime,
		},
		System: struct {
			UptimeSeconds int64  `json:"uptime_seconds"`
			MemoryMb      uint64 `json:"memory_mb"`
			Goroutines    int    `json:"goroutines"`
			GCRuns        uint32 `json:"gc_runs"`
		}{
			UptimeSeconds: int64(time.Since(startTime).Seconds()),
			MemoryMb:      m.Alloc / 1024 / 1024,
			Goroutines:    runtime.NumGoroutine(),
			GCRuns:        m.NumGC,
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

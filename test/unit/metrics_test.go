package http_test

import (
	"encoding/json"
	"financial-system-pro/internal/adapters/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetrics_RecordDeposit testa registro de depósito
func TestMetrics_RecordDeposit(t *testing.T) {
	// Reset GlobalMetrics
	atomic.StoreInt64(&http.GlobalMetrics.DepositCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.RequestCount, 0)

	http.RecordDeposit()

	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.DepositCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.RequestCount))
}

// TestMetrics_RecordWithdraw testa registro de saque
func TestMetrics_RecordWithdraw(t *testing.T) {
	// Reset http.Metrics
	atomic.StoreInt64(&http.GlobalMetrics.DepositCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.WithdrawCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.RequestCount, 0)

	http.RecordWithdraw()

	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.WithdrawCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.RequestCount))
}

// TestMetrics_RecordTransfer testa registro de transferência
func TestMetrics_RecordTransfer(t *testing.T) {
	// Reset http.Metrics
	atomic.StoreInt64(&http.GlobalMetrics.TransferCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.RequestCount, 0)

	http.RecordTransfer()

	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.TransferCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.RequestCount))
}

// TestMetrics_RecordFailure testa registro de falha
func TestMetrics_RecordFailure(t *testing.T) {
	// Reset http.Metrics
	atomic.StoreInt64(&http.GlobalMetrics.FailureCount, 0)

	http.RecordFailure()

	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.FailureCount))
}

// TestMetrics_RecordRequestTime testa registro de tempo de requisição
func TestMetrics_RecordRequestTime(t *testing.T) {
	// Reset http.Metrics
	atomic.StoreInt64(&http.GlobalMetrics.TotalRequestTime, 0)

	duration := 150 * time.Millisecond
	http.RecordRequestTime(duration)

	totalTime := atomic.LoadInt64(&http.GlobalMetrics.TotalRequestTime)
	assert.Equal(t, int64(duration.Nanoseconds()), totalTime)
}

// TestMetrics_RecordMultiple testa múltiplas operações
func TestMetrics_RecordMultiple(t *testing.T) {
	// Reset http.Metrics
	atomic.StoreInt64(&http.GlobalMetrics.DepositCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.WithdrawCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.TransferCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.RequestCount, 0)

	http.RecordDeposit()
	http.RecordDeposit()
	http.RecordWithdraw()
	http.RecordTransfer()

	assert.Equal(t, int64(2), atomic.LoadInt64(&http.GlobalMetrics.DepositCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.WithdrawCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&http.GlobalMetrics.TransferCount))
	assert.Equal(t, int64(4), atomic.LoadInt64(&http.GlobalMetrics.RequestCount))
}

// TestGetMetrics testa endpoint de métricas
func TestGetMetrics(t *testing.T) {
	// Reset http.Metrics
	atomic.StoreInt64(&http.GlobalMetrics.DepositCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.WithdrawCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.TransferCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.FailureCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.RequestCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.TotalRequestTime, 0)

	http.RecordDeposit()
	http.RecordWithdraw()
	http.RecordTransfer()
	http.RecordFailure()
	http.RecordRequestTime(150 * time.Millisecond)

	app := fiber.New()
	app.Get("/http.Metrics", http.GetMetrics)

	req := httptest.NewRequest("GET", "/http.Metrics", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Contains(t, response, "transactions")
	assert.Contains(t, response, "api")
	assert.Contains(t, response, "system")
}

// TestMetricsMiddleware testa middleware de métricas
func TestMetricsMiddleware(t *testing.T) {
	// Reset http.Metrics
	atomic.StoreInt64(&http.GlobalMetrics.TotalRequestTime, 0)

	middleware := http.MetricsMiddleware()

	app := fiber.New()
	app.Use(middleware)
	app.Get("/test", func(c *fiber.Ctx) error {
		time.Sleep(10 * time.Millisecond) // Simulate work
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1) // No timeout
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify http.Metrics were recorded
	totalTime := atomic.LoadInt64(&http.GlobalMetrics.TotalRequestTime)
	assert.Greater(t, totalTime, int64(0))
}

// TestMetrics_Concurrent testa acesso concorrente às métricas
func TestMetrics_Concurrent(t *testing.T) {
	// Reset http.Metrics
	atomic.StoreInt64(&http.GlobalMetrics.DepositCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.WithdrawCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.TransferCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.FailureCount, 0)
	atomic.StoreInt64(&http.GlobalMetrics.RequestCount, 0)

	done := make(chan bool)

	// Spawn multiple goroutines to record http.Metrics concurrently
	for i := 0; i < 10; i++ {
		go func() {
			http.RecordDeposit()
			http.RecordWithdraw()
			http.RecordTransfer()
			http.RecordFailure()
			http.RecordRequestTime(10 * time.Millisecond)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all http.Metrics were recorded
	assert.Equal(t, int64(10), atomic.LoadInt64(&http.GlobalMetrics.DepositCount))
	assert.Equal(t, int64(10), atomic.LoadInt64(&http.GlobalMetrics.WithdrawCount))
	assert.Equal(t, int64(10), atomic.LoadInt64(&http.GlobalMetrics.TransferCount))
	assert.Equal(t, int64(10), atomic.LoadInt64(&http.GlobalMetrics.FailureCount))
}

// BenchmarkMetrics_RecordDeposit benchmarks deposit recording
func BenchmarkMetrics_RecordDeposit(b *testing.B) {
	atomic.StoreInt64(&http.GlobalMetrics.DepositCount, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		http.RecordDeposit()
	}
}

// BenchmarkMetrics_RecordRequestTime benchmarks request time recording
func BenchmarkMetrics_RecordRequestTime(b *testing.B) {
	atomic.StoreInt64(&http.GlobalMetrics.TotalRequestTime, 0)
	duration := 100 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		http.RecordRequestTime(duration)
	}
}

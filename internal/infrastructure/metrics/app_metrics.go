package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// AppMetrics representa métricas internas da aplicação (não Prometheus) usadas em endpoints HTTP.
type AppMetrics struct {
	LastUpdated      time.Time
	DepositCount     int64
	WithdrawCount    int64
	TransferCount    int64
	FailureCount     int64
	TotalRequestTime int64
	RequestCount     int64
	mu               sync.RWMutex
}

// GlobalAppMetrics instância global
var GlobalAppMetrics = &AppMetrics{LastUpdated: time.Now()}

// RecordDeposit incrementa contador de depósitos
func RecordDeposit() {
	atomic.AddInt64(&GlobalAppMetrics.DepositCount, 1)
	atomic.AddInt64(&GlobalAppMetrics.RequestCount, 1)
}

// RecordWithdraw incrementa contador de saques
func RecordWithdraw() {
	atomic.AddInt64(&GlobalAppMetrics.WithdrawCount, 1)
	atomic.AddInt64(&GlobalAppMetrics.RequestCount, 1)
}

// RecordTransfer incrementa contador de transferências
func RecordTransfer() {
	atomic.AddInt64(&GlobalAppMetrics.TransferCount, 1)
	atomic.AddInt64(&GlobalAppMetrics.RequestCount, 1)
}

// RecordFailure incrementa contador de falhas
func RecordFailure() {
	atomic.AddInt64(&GlobalAppMetrics.FailureCount, 1)
}

// RecordRequestTime registra tempo de requisição
func RecordRequestTime(duration time.Duration) {
	atomic.AddInt64(&GlobalAppMetrics.TotalRequestTime, int64(duration.Nanoseconds()))
}

// Snapshot gera mapa com métricas atuais e estatísticas de sistema.
func Snapshot(start time.Time) map[string]interface{} {
	GlobalAppMetrics.mu.RLock()
	defer GlobalAppMetrics.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	deposits := atomic.LoadInt64(&GlobalAppMetrics.DepositCount)
	withdraws := atomic.LoadInt64(&GlobalAppMetrics.WithdrawCount)
	transfers := atomic.LoadInt64(&GlobalAppMetrics.TransferCount)
	failures := atomic.LoadInt64(&GlobalAppMetrics.FailureCount)
	requests := atomic.LoadInt64(&GlobalAppMetrics.RequestCount)
	totalTime := atomic.LoadInt64(&GlobalAppMetrics.TotalRequestTime)

	avgResponseTime := float64(0)
	if requests > 0 {
		avgResponseTime = float64(totalTime) / float64(requests) / float64(time.Millisecond)
	}

	return map[string]interface{}{
		"transactions": map[string]interface{}{
			"deposits":  deposits,
			"withdraws": withdraws,
			"transfers": transfers,
			"failures":  failures,
			"total":     deposits + withdraws + transfers,
		},
		"api": map[string]interface{}{
			"total_requests":       requests,
			"avg_response_time_ms": avgResponseTime,
		},
		"system": map[string]interface{}{
			"uptime_seconds": int64(time.Since(start).Seconds()),
			"memory_mb":      m.Alloc / 1024 / 1024,
			"goroutines":     runtime.NumGoroutine(),
			"gc_runs":        m.NumGC,
		},
	}
}

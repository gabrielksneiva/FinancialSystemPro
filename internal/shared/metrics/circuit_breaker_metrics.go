package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Circuit Breaker Metrics

var (
	// CircuitBreakerStateGauge rastreia o estado do circuit breaker
	// 0 = closed, 1 = half-open, 2 = open
	CircuitBreakerStateGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of circuit breaker (0=closed, 1=half-open, 2=open)",
		},
		[]string{"breaker_name"},
	)

	// CircuitBreakerTripsTotal conta quantas vezes o circuit breaker abriu
	CircuitBreakerTripsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_trips_total",
			Help: "Total number of times circuit breaker opened",
		},
		[]string{"breaker_name"},
	)

	// CircuitBreakerRequestsTotal conta todas as requests através do circuit breaker
	CircuitBreakerRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_requests_total",
			Help: "Total number of requests through circuit breaker",
		},
		[]string{"breaker_name", "result"}, // result: success, failure, rejected
	)

	// CircuitBreakerFailuresTotal conta falhas de requests
	CircuitBreakerFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_failures_total",
			Help: "Total number of failed requests",
		},
		[]string{"breaker_name"},
	)

	// CircuitBreakerConsecutiveFailuresGauge rastreia falhas consecutivas
	CircuitBreakerConsecutiveFailuresGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_consecutive_failures",
			Help: "Number of consecutive failures",
		},
		[]string{"breaker_name"},
	)
)

// RecordCircuitBreakerState registra o estado atual do circuit breaker
func RecordCircuitBreakerState(name string, state int) {
	CircuitBreakerStateGauge.WithLabelValues(name).Set(float64(state))
}

// RecordCircuitBreakerTrip registra quando um circuit breaker abre
func RecordCircuitBreakerTrip(name string) {
	CircuitBreakerTripsTotal.WithLabelValues(name).Inc()
}

// RecordCircuitBreakerRequest registra uma request através do circuit breaker
func RecordCircuitBreakerRequest(name string, result string) {
	CircuitBreakerRequestsTotal.WithLabelValues(name, result).Inc()
}

// RecordCircuitBreakerFailure registra uma falha
func RecordCircuitBreakerFailure(name string) {
	CircuitBreakerFailuresTotal.WithLabelValues(name).Inc()
}

// RecordCircuitBreakerConsecutiveFailures registra falhas consecutivas
func RecordCircuitBreakerConsecutiveFailures(name string, count int) {
	CircuitBreakerConsecutiveFailuresGauge.WithLabelValues(name).Set(float64(count))
}

package breaker

import (
	"financial-system-pro/internal/shared/metrics"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// CircuitBreaker wrapper para gobreaker com logging
type CircuitBreaker struct {
	breaker *gobreaker.CircuitBreaker
	name    string
	logger  *zap.Logger
}

// Settings configurações do circuit breaker
type Settings struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	ReadyToTrip   func(counts gobreaker.Counts) bool
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// NewCircuitBreaker cria um novo circuit breaker
func NewCircuitBreaker(settings Settings, logger *zap.Logger) *CircuitBreaker {
	gbSettings := gobreaker.Settings{
		Name:        settings.Name,
		MaxRequests: settings.MaxRequests,
		Interval:    settings.Interval,
		Timeout:     settings.Timeout,
		ReadyToTrip: settings.ReadyToTrip,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("circuit breaker state changed",
				zap.String("breaker", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)

			// Record state change in Prometheus
			var stateValue int
			switch to {
			case gobreaker.StateClosed:
				stateValue = 0
			case gobreaker.StateHalfOpen:
				stateValue = 1
			case gobreaker.StateOpen:
				stateValue = 2
				metrics.RecordCircuitBreakerTrip(name)
			}
			metrics.RecordCircuitBreakerState(name, stateValue)

			if settings.OnStateChange != nil {
				settings.OnStateChange(name, from, to)
			}
		},
	}

	return &CircuitBreaker{
		breaker: gobreaker.NewCircuitBreaker(gbSettings),
		name:    settings.Name,
		logger:  logger,
	}
}

// Execute executa uma função através do circuit breaker
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	result, err := cb.breaker.Execute(fn)

	// Record metrics
	counts := cb.breaker.Counts()
	metrics.RecordCircuitBreakerConsecutiveFailures(cb.name, int(counts.ConsecutiveFailures))

	if err != nil {
		cb.logger.Error("circuit breaker execution failed",
			zap.String("breaker", cb.name),
			zap.Error(err),
		)
		metrics.RecordCircuitBreakerRequest(cb.name, "failure")
		metrics.RecordCircuitBreakerFailure(cb.name)
	} else {
		metrics.RecordCircuitBreakerRequest(cb.name, "success")
	}

	return result, err
} // State retorna o estado atual do circuit breaker
func (cb *CircuitBreaker) State() gobreaker.State {
	return cb.breaker.State()
}

// Counts retorna as estatísticas do circuit breaker
func (cb *CircuitBreaker) Counts() gobreaker.Counts {
	return cb.breaker.Counts()
}

// DefaultReadyToTrip função padrão para abrir o circuito
// Abre após 5 falhas consecutivas ou taxa de erro > 50%
func DefaultReadyToTrip(counts gobreaker.Counts) bool {
	failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
	return counts.ConsecutiveFailures > 5 || failureRatio >= 0.5
}

// NewDefaultCircuitBreaker cria um circuit breaker com configurações padrão
func NewDefaultCircuitBreaker(name string, logger *zap.Logger) *CircuitBreaker {
	return NewCircuitBreaker(Settings{
		Name:        name,
		MaxRequests: 3,                // Máximo de requests em half-open
		Interval:    time.Minute,      // Janela de contagem
		Timeout:     30 * time.Second, // Tempo em open antes de tentar half-open
		ReadyToTrip: DefaultReadyToTrip,
	}, logger)
}

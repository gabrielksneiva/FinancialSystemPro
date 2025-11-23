package workers

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CircuitBreakerState representa o estado do circuit breaker
type CircuitBreakerState string

const (
	Closed   CircuitBreakerState = "closed"    // Normal, operando
	Open     CircuitBreakerState = "open"      // Falhas detectadas, rejeitando
	HalfOpen CircuitBreakerState = "half_open" // Testando se voltou
)

// CircuitBreaker implementa padrão circuit breaker para Redis
type CircuitBreaker struct {
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	mu              sync.RWMutex
	logger          *zap.Logger

	// Configuração
	failureThreshold int           // Falhas antes de abrir
	resetTimeout     time.Duration // Tempo antes de tentar half-open
	successThreshold int           // Sucessos para fechar novamente
}

// NewCircuitBreaker cria um novo circuit breaker
func NewCircuitBreaker(logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		state:            Closed,
		failureThreshold: 5,                // Abrir após 5 falhas
		resetTimeout:     30 * time.Second, // Tentar novamente após 30s
		successThreshold: 2,                // Fechar após 2 sucessos em half-open
		logger:           logger,
	}
}

// Call executa uma operação através do circuit breaker
func (cb *CircuitBreaker) Call(operation func() error) error {
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()

	switch state {
	case Open:
		// Verificar se deve tentar half-open
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mu.Lock()
			cb.state = HalfOpen
			cb.successCount = 0
			cb.mu.Unlock()
			cb.logger.Info("circuit breaker entering half-open state")
		} else {
			return fmt.Errorf("circuit breaker is open - redis appears to be down")
		}
		fallthrough

	case HalfOpen:
		// Tentar operação
		err := operation()
		if err == nil {
			cb.recordSuccess()
		} else {
			cb.recordFailure()
		}
		return err

	case Closed:
		// Operação normal
		err := operation()
		if err != nil {
			cb.recordFailure()
		} else {
			cb.recordSuccess()
		}
		return err
	}

	return fmt.Errorf("unknown circuit breaker state: %s", state)
}

// recordSuccess registra um sucesso
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	if cb.state == HalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = Closed
			cb.logger.Info("circuit breaker closed - redis is healthy again")
		}
	}
}

// recordFailure registra uma falha
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.successCount = 0

	if cb.state == Closed && cb.failureCount >= cb.failureThreshold {
		cb.state = Open
		cb.logger.Warn("circuit breaker opened - redis appears to be down",
			zap.Int("failure_count", cb.failureCount),
		)
	} else if cb.state == HalfOpen {
		cb.state = Open
		cb.logger.Warn("circuit breaker reopened - redis still not responding")
	}
}

// GetState retorna o estado atual
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// IsHealthy retorna se o circuit breaker permite operações
func (cb *CircuitBreaker) IsHealthy() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == Closed
}

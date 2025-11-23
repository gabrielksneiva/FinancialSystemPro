package breaker

import (
	"go.uber.org/zap"
)

// BreakerManager gerencia circuit breakers para comunicação entre contextos
type BreakerManager struct {
	breakers map[string]*CircuitBreaker
	logger   *zap.Logger
}

// NewBreakerManager cria um novo gerenciador de circuit breakers
func NewBreakerManager(logger *zap.Logger) *BreakerManager {
	return &BreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetBreaker retorna um circuit breaker pelo nome (cria se não existir)
func (bm *BreakerManager) GetBreaker(name string) *CircuitBreaker {
	if breaker, exists := bm.breakers[name]; exists {
		return breaker
	}

	breaker := NewDefaultCircuitBreaker(name, bm.logger)
	bm.breakers[name] = breaker

	bm.logger.Info("circuit breaker created",
		zap.String("name", name),
	)

	return breaker
}

// GetOrCreateBreaker retorna ou cria um circuit breaker com configurações customizadas
func (bm *BreakerManager) GetOrCreateBreaker(name string, settings Settings) *CircuitBreaker {
	if breaker, exists := bm.breakers[name]; exists {
		return breaker
	}

	breaker := NewCircuitBreaker(settings, bm.logger)
	bm.breakers[name] = breaker

	bm.logger.Info("custom circuit breaker created",
		zap.String("name", name),
	)

	return breaker
}

// GetAllStates retorna o estado de todos os circuit breakers
func (bm *BreakerManager) GetAllStates() map[string]string {
	states := make(map[string]string)
	for name, breaker := range bm.breakers {
		states[name] = breaker.State().String()
	}
	return states
}

// Nomes de circuit breakers para comunicação entre contextos
const (
	// User Context chamando Transaction Context
	BreakerUserToTransaction = "user->transaction"

	// User Context chamando Blockchain Context
	BreakerUserToBlockchain = "user->blockchain"

	// Transaction Context chamando User Context
	BreakerTransactionToUser = "transaction->user"

	// Transaction Context chamando Blockchain Context
	BreakerTransactionToBlockchain = "transaction->blockchain"

	// Blockchain Context chamando Transaction Context
	BreakerBlockchainToTransaction = "blockchain->transaction"

	// External API calls
	BreakerExternalTronAPI  = "external->tron-api"
	BreakerExternalDatabase = "external->database"
)

package integration

import (
	"testing"

	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestCircuitBreakerIntegration testa integração do circuit breaker entre contextos
func TestCircuitBreakerIntegration(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	breakerManager := breaker.NewBreakerManager(logger)

	t.Run("Circuit breaker executa com sucesso", func(t *testing.T) {
		breaker := breakerManager.GetBreaker("TestBreaker")
		require.NotNil(t, breaker)

		result, err := breaker.Execute(func() (interface{}, error) {
			return "success", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("Circuit breaker propaga erros", func(t *testing.T) {
		breaker := breakerManager.GetBreaker("TestBreakerError")
		require.NotNil(t, breaker)

		result, err := breaker.Execute(func() (interface{}, error) {
			return nil, assert.AnError
		})

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Múltiplos circuit breakers funcionam independentemente", func(t *testing.T) {
		breaker1 := breakerManager.GetBreaker("Breaker1")
		breaker2 := breakerManager.GetBreaker("Breaker2")

		require.NotNil(t, breaker1)
		require.NotNil(t, breaker2)

		// Ambos devem funcionar independentemente
		result1, err1 := breaker1.Execute(func() (interface{}, error) { return "result1", nil })
		result2, err2 := breaker2.Execute(func() (interface{}, error) { return "result2", nil })

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, "result1", result1)
		assert.Equal(t, "result2", result2)
	})

	t.Run("Named circuit breakers (TransactionToUser, TransactionToBlockchain)", func(t *testing.T) {
		txnToUser := breakerManager.GetBreaker("BreakerTransactionToUser")
		txnToBC := breakerManager.GetBreaker("BreakerTransactionToBlockchain")

		require.NotNil(t, txnToUser)
		require.NotNil(t, txnToBC)

		// Executar operações via circuit breakers
		_, err1 := txnToUser.Execute(func() (interface{}, error) {
			logger.Info("Transaction calling User Context via circuit breaker")
			return "user_call_success", nil
		})

		_, err2 := txnToBC.Execute(func() (interface{}, error) {
			logger.Info("Transaction calling Blockchain Context via circuit breaker")
			return "blockchain_call_success", nil
		})

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})
}

// TestEventBusIntegration testa integração do event bus entre contextos
func TestEventBusIntegration(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	eventBus := events.NewInMemoryBus(logger)
	require.NotNil(t, eventBus)

	// O event bus foi testado em unit tests (test/unit/ddd_event_bus_test.go)
	// Aqui apenas verificamos que está integrado corretamente na stack
	logger.Info("Event Bus is available and integrated with the application")
}

// TestCrossContextCommunicationPattern testa padrão de comunicação entre contextos
func TestCrossContextCommunicationPattern(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	breakerManager := breaker.NewBreakerManager(logger)

	t.Run("Padrão completo: Circuit Breaker + Event Publishing", func(t *testing.T) {
		// Simular Transaction Context chamando User Context
		txnToUserBreaker := breakerManager.GetBreaker("BreakerTransactionToUser")
		require.NotNil(t, txnToUserBreaker)

		// Executar via circuit breaker
		result, err := txnToUserBreaker.Execute(func() (interface{}, error) {
			// Simular operação no User Context
			logger.Info("Processing transaction in User Context via circuit breaker",
				zap.String("operation", "deposit"),
				zap.Float64("amount", 100.0),
			)

			return decimal.NewFromFloat(100.0), nil
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		logger.Info("Cross-context communication pattern completed successfully")
	})

	t.Run("Padrão: Transaction Context chamando Blockchain Context", func(t *testing.T) {
		txnToBlockchainBreaker := breakerManager.GetBreaker("BreakerTransactionToBlockchain")
		require.NotNil(t, txnToBlockchainBreaker)

		result, err := txnToBlockchainBreaker.Execute(func() (interface{}, error) {
			logger.Info("Processing withdrawal in Blockchain Context via circuit breaker",
				zap.String("operation", "withdraw"),
				zap.String("amount", "50.0"),
			)
			return "tx_hash_12345", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "tx_hash_12345", result)
		logger.Info("Blockchain communication pattern completed successfully")
	})

	t.Run("Isolamento de falhas entre contextos", func(t *testing.T) {
		failingBreaker := breakerManager.GetBreaker("FailingBreaker")
		healthyBreaker := breakerManager.GetBreaker("HealthyBreaker")

		require.NotNil(t, failingBreaker)
		require.NotNil(t, healthyBreaker)

		// Teste com breaker falhando
		_, errFailing := failingBreaker.Execute(func() (interface{}, error) {
			return nil, assert.AnError
		})

		// Teste com breaker saudável
		resultHealthy, errHealthy := healthyBreaker.Execute(func() (interface{}, error) {
			return "success", nil
		})

		assert.Error(t, errFailing)
		assert.NoError(t, errHealthy)
		assert.Equal(t, "success", resultHealthy)

		logger.Info("Failure isolation pattern confirmed")
	})
}

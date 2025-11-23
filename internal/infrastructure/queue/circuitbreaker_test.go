package workers

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

// helper para criar breaker novo com logger nop
func newBreaker() *CircuitBreaker {
	return NewCircuitBreaker(zap.NewNop())
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cb := newBreaker()
	if st := cb.GetState(); st != Closed {
		t.Fatalf("estado inicial esperado Closed, obtido %s", st)
	}

	failErr := errors.New("falha artificial")
	for i := 0; i < cb.failureThreshold; i++ {
		err := cb.Call(func() error { return failErr })
		if err == nil {
			t.Fatalf("esperava erro na chamada %d", i)
		}
	}

	if st := cb.GetState(); st != Open {
		t.Fatalf("esperado estado Open após %d falhas, obtido %s", cb.failureThreshold, st)
	}
	if cb.IsHealthy() {
		t.Fatalf("IsHealthy deve ser falso quando Open")
	}
}

func TestCircuitBreaker_HalfOpenAndClosesAfterSuccesses(t *testing.T) {
	cb := newBreaker()
	// Forçar abrir
	for i := 0; i < cb.failureThreshold; i++ {
		_ = cb.Call(func() error { return errors.New("falha") })
	}
	if cb.GetState() != Open {
		t.Fatalf("esperado estado Open depois de falhas iniciais")
	}
	// Manipular tempo para permitir transição para HalfOpen
	cb.mu.Lock()
	cb.lastFailureTime = time.Now().Add(-cb.resetTimeout - time.Second)
	cb.mu.Unlock()

	// Primeira chamada (HalfOpen -> successCount=1)
	if err := cb.Call(func() error { return nil }); err != nil {
		t.Fatalf("não esperava erro na primeira tentativa half-open: %v", err)
	}
	if cb.GetState() != HalfOpen {
		// Ainda half-open após 1 sucesso
		if cb.GetState() != Closed { // permitir debug se fechou cedo
			t.Fatalf("esperado HalfOpen após 1 sucesso, estado=%s", cb.GetState())
		}
	}
	// Segunda chamada deve fechar
	if err := cb.Call(func() error { return nil }); err != nil {
		t.Fatalf("não esperava erro na segunda tentativa half-open: %v", err)
	}
	if cb.GetState() != Closed {
		t.Fatalf("esperado Closed após %d sucessos", cb.successThreshold)
	}
	if !cb.IsHealthy() {
		t.Fatalf("IsHealthy deve ser verdadeiro quando Closed")
	}
}

func TestCircuitBreaker_ReopensOnFailureInHalfOpen(t *testing.T) {
	cb := newBreaker()
	// Abrir com falhas
	for i := 0; i < cb.failureThreshold; i++ {
		_ = cb.Call(func() error { return errors.New("falha") })
	}
	if cb.GetState() != Open {
		t.Fatalf("esperado Open antes de half-open test")
	}
	// Forçar half-open
	cb.mu.Lock()
	cb.lastFailureTime = time.Now().Add(-cb.resetTimeout - time.Second)
	cb.mu.Unlock()

	// Chamada que falha em half-open deve reabrir
	_ = cb.Call(func() error { return errors.New("falhou") })
	if cb.GetState() != Open {
		t.Fatalf("esperado Open após falha em HalfOpen, estado=%s", cb.GetState())
	}
}

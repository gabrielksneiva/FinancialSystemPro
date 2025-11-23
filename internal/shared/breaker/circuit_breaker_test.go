package breaker_test

import (
	"errors"
	"financial-system-pro/internal/shared/breaker"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDefaultReadyToTrip(t *testing.T) {
	counts := gobreaker.Counts{Requests: 10, TotalFailures: 6, ConsecutiveFailures: 6}
	assert.True(t, breaker.DefaultReadyToTrip(counts))

	counts = gobreaker.Counts{Requests: 10, TotalFailures: 4, ConsecutiveFailures: 2}
	assert.False(t, breaker.DefaultReadyToTrip(counts))
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	var transitioned bool
	logger := zap.NewNop()
	cb := breaker.NewCircuitBreaker(breaker.Settings{
		Name: "test",
		MaxRequests: 1,
		Interval: 50 * time.Millisecond,
		Timeout: 100 * time.Millisecond,
		ReadyToTrip: func(c gobreaker.Counts) bool { return c.ConsecutiveFailures >= 2 },
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) { transitioned = true },
	}, logger)

	// Primeiro sucesso mantém fechado
	_, err := cb.Execute(func() (interface{}, error) { return "ok", nil })
	assert.NoError(t, err)
	assert.Equal(t, gobreaker.StateClosed, cb.State())

	// Duas falhas devem abrir
	_, err = cb.Execute(func() (interface{}, error) { return nil, errors.New("falha") })
	assert.Error(t, err)
	_, err = cb.Execute(func() (interface{}, error) { return nil, errors.New("falha") })
	assert.Error(t, err)
	assert.Equal(t, gobreaker.StateOpen, cb.State())
	assert.True(t, transitioned)

	// Esperar timeout para half-open
	time.Sleep(120 * time.Millisecond)
	// Half-open sucesso fecha
	_, err = cb.Execute(func() (interface{}, error) { return "recover", nil })
	assert.NoError(t, err)
	assert.Equal(t, gobreaker.StateClosed, cb.State())
}

func TestBreakerManager_GetBreakerReuse(t *testing.T) {
	bm := breaker.NewBreakerManager(zap.NewNop())
	b1 := bm.GetBreaker("alpha")
	b2 := bm.GetBreaker("alpha")
	assert.Equal(t, b1, b2)

	states := bm.GetAllStates()
	assert.Equal(t, b1.State().String(), states["alpha"])
}

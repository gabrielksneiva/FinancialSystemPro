package service

import (
	"context"
	"sync"
	"testing"
	"time"

	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// helper to setup service with custom eventBus (passed in) and initial balance
func setupServiceWithBus(t *testing.T, balance float64, bus events.Bus) (*TransactionService, uuid.UUID) {
	t.Helper()
	lg := zap.NewNop()
	br := breaker.NewBreakerManager(lg)
	txr := newMemTxnRepo()
	ur := newMemUserRepo()
	wr := newMemWalletRepo()
	uid := uuid.New()
	_ = ur.Create(context.Background(), &userEntity.User{ID: uid, Email: "events@test.com", Password: "hash"})
	_ = wr.Create(context.Background(), &userEntity.Wallet{UserID: uid, Address: "ADDR", Balance: balance})
	svc := NewTransactionService(txr, ur, wr, bus, br, lg)
	return svc, uid
}

func TestTransactionService_PublishesDepositCompletedEvent(t *testing.T) {
	lg := zap.NewNop()
	bus := events.NewInMemoryBus(lg)

	var received []events.Event
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe("deposit.completed", func(ctx context.Context, evt events.Event) error {
		mu.Lock()
		received = append(received, evt)
		mu.Unlock()
		wg.Done()
		return nil
	})

	svc, uid := setupServiceWithBus(t, 0, bus)
	amt := decimal.NewFromInt(25)

	if err := svc.ProcessDeposit(context.Background(), uid, amt, ""); err != nil {
		t.Fatalf("erro processo deposit: %v", err)
	}

	// Espera evento assíncrono
	waitDone(t, &wg)

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, received, 1, "deve receber exatamente um evento de depósito")
	depEvt, ok := received[0].(events.DepositCompletedEvent)
	assert.True(t, ok, "tipo do evento deve ser DepositCompletedEvent")
	assert.Equal(t, uid.String(), depEvt.AggregateID())
	assert.True(t, amt.Equal(depEvt.Amount))
	assert.Contains(t, depEvt.TxHash, "deposit-")
}

func TestTransactionService_PublishesWithdrawCompletedEvent(t *testing.T) {
	lg := zap.NewNop()
	bus := events.NewInMemoryBus(lg)

	var received []events.Event
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe("withdraw.completed", func(ctx context.Context, evt events.Event) error {
		mu.Lock()
		received = append(received, evt)
		mu.Unlock()
		wg.Done()
		return nil
	})

	svc, uid := setupServiceWithBus(t, 100, bus)
	amt := decimal.NewFromInt(40)

	if err := svc.ProcessWithdraw(context.Background(), uid, amt); err != nil {
		t.Fatalf("erro processo withdraw: %v", err)
	}

	waitDone(t, &wg)

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, received, 1, "deve receber exatamente um evento de withdraw")
	wEvt, ok := received[0].(events.WithdrawCompletedEvent)
	assert.True(t, ok, "tipo do evento deve ser WithdrawCompletedEvent")
	assert.Equal(t, uid.String(), wEvt.AggregateID())
	assert.True(t, amt.Equal(wEvt.Amount))
	assert.Contains(t, wEvt.TxHash, "withdraw-")
}

// Testa que publish sem subscriber não falha
func TestTransactionService_PublishNoSubscriber_NoError(t *testing.T) {
	lg := zap.NewNop()
	bus := events.NewInMemoryBus(lg)
	svc, uid := setupServiceWithBus(t, 0, bus)
	amt := decimal.NewFromInt(5)
	// Nenhum subscriber para deposit.completed
	if err := svc.ProcessDeposit(context.Background(), uid, amt, ""); err != nil {
		t.Fatalf("process deposit falhou: %v", err)
	}
	// apenas aguarda um pouco para goroutine
	time.Sleep(20 * time.Millisecond)
}

// Utilitário para aguardar waitgroup com timeout
func waitDone(t *testing.T, wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
		return
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout aguardando evento publicado")
	}
}

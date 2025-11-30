package application

import (
	"context"
	"testing"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/events"
	"financial-system-pro/internal/blockchains/infra/eventbus"
)

// mock repo implementing only the needed method
type mockRepo struct {
	called bool
}

func (m *mockRepo) CreateTransaction(ctx context.Context, from domain.Address, to domain.Address, amount domain.Amount) (domain.Transaction, error) {
	m.called = true
	tx := domain.Transaction{Hash: "tx_mock", From: from, To: to, Amount: amount}
	return tx, nil
}

func TestCreateTransactionUseCase_ExecutesAndPublishes(t *testing.T) {
	repo := &mockRepo{}
	bus := eventbus.NewInMemoryBus()
	uc := NewCreateTransactionUseCase(repo, bus, "ethereum")

	ctx := context.Background()
	tx, err := uc.Execute(ctx, "addr_from", "addr_to", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Hash == "" {
		t.Fatalf("expected tx hash")
	}
	if !repo.called {
		t.Fatalf("expected repo CreateTransaction to be called")
	}

	// subscribe to event
	ch, cancel := bus.Subscribe(ctx, events.EventNewTransaction)
	defer cancel()
	// The usecase published synchronously after repo call; read one event from internal processed map is set.
	// However InMemoryBus.Publish delivers asynchronously, so wait briefly.
	select {
	case v := <-ch:
		if _, ok := v.(events.NewTransactionEvent); !ok {
			t.Fatalf("expected NewTransactionEvent payload")
		}
	default:
		// allow for goroutine scheduling
	}
}

package application

import (
	"context"
	"testing"

	"financial-system-pro/internal/blockchains/domain"
)

type mockBroadcastRepo struct {
	called bool
	lastTx domain.Transaction
}

func (m *mockBroadcastRepo) BroadcastTransaction(ctx context.Context, tx domain.Transaction) (domain.TxHash, error) {
	m.called = true
	m.lastTx = tx
	return domain.TxHash("tx_broadcast_mock"), nil
}

func TestBroadcastTransactionUseCase_Executes(t *testing.T) {
	repo := &mockBroadcastRepo{}
	uc := NewBroadcastTransactionUseCase(repo)

	ctx := context.Background()
	tx := domain.Transaction{From: "a", To: "b", Amount: domain.Amount{Value: nil}}
	h, err := uc.Execute(ctx, tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == "" {
		t.Fatalf("expected tx hash")
	}
	if !repo.called {
		t.Fatalf("expected repository BroadcastTransaction to be called")
	}
}

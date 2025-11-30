package application

import (
	"context"
	"testing"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/infra/broadcaster"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

// mem persistence implementing TxPersistence
type memTxRepo struct {
	last domain.Transaction
}

func (m *memTxRepo) CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	if tx.Hash == "" {
		tx.Hash = "mem_tx_1"
	}
	m.last = tx
	return tx, nil
}

func TestBroadcastTransaction_FullWiring(t *testing.T) {
	mem := &memTxRepo{}
	sim := rpc.NewSimulatedClient()
	broad := broadcaster.NewTransactionBroadcaster(mem, sim)
	uc := NewBroadcastTransactionUseCase(broad)

	tx := domain.Transaction{From: "a", To: "b", Amount: domain.Amount{Value: nil}}
	h, err := uc.Execute(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if h == "" {
		t.Fatalf("expected hash")
	}
}

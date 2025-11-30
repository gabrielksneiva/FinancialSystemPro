package broadcaster

import (
	"context"
	"testing"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

type memRepo struct {
	created domain.Transaction
}

func (m *memRepo) CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	if tx.Hash == "" {
		tx.Hash = "persisted_tx"
	}
	m.created = tx
	return tx, nil
}

func TestTransactionBroadcaster_Broadcasts(t *testing.T) {
	repo := &memRepo{}
	client := rpc.NewSimulatedClient()
	b := NewTransactionBroadcaster(repo, client)

	tx := domain.Transaction{From: "f", To: "t", Amount: domain.Amount{Value: nil}}
	h, err := b.BroadcastTransaction(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if h == "" {
		t.Fatalf("expected hash")
	}
}

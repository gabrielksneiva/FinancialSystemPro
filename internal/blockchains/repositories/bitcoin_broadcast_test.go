package repositories

import (
	"context"
	"fmt"
	"testing"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/infra/eventbus"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

func TestBitcoinRepository_BroadcastAndStatus(t *testing.T) {
	client := rpc.NewSimulatedClient()
	bus := eventbus.NewInMemoryBus()
	repo := NewBitcoinRepository(client, bus)
	ctx := context.Background()
	_ = repo.Connect(ctx)

	amt, _ := domain.NewAmountFromString("10")
	from, _ := domain.NewAddress("f1")
	to, _ := domain.NewAddress("t1")
	tx, _ := repo.CreateTransaction(ctx, from, to, amt)

	// broadcast returns hash
	if h, err := repo.BroadcastTransaction(ctx, tx); err != nil || h == "" {
		t.Fatalf("broadcast failed: %v", err)
	}

	// nonexistent tx status should return error
	if _, err := repo.GetTransactionStatus(ctx, domain.TxHash("no_such")); err == nil {
		t.Fatalf("expected error for nonexistent tx")
	}
}

func TestBitcoinRepository_BroadcastErrorPath(t *testing.T) {
	client := rpc.NewSimulatedClient()
	client.SubmitErr = fmt.Errorf("submit fail")
	bus := eventbus.NewInMemoryBus()
	repo := NewBitcoinRepository(client, bus)
	ctx := context.Background()
	_ = repo.Connect(ctx)

	amt, _ := domain.NewAmountFromString("1")
	from, _ := domain.NewAddress("fa")
	to, _ := domain.NewAddress("tb")
	tx := domain.Transaction{From: from, To: to, Amount: amt}
	if _, err := repo.BroadcastTransaction(ctx, tx); err == nil {
		t.Fatalf("expected broadcast error")
	}
}

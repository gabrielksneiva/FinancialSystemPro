package repositories

import (
	"context"
	"testing"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/infra/eventbus"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

func TestConcreteRepos_CommonMethods(t *testing.T) {
	clients := map[string]*rpc.SimulatedClient{}
	bus := eventbus.NewInMemoryBus()
	repos := []struct {
		name string
		repo func(*rpc.SimulatedClient, *eventbus.InMemoryBus) interface{}
	}{
		{"ethereum", func(c *rpc.SimulatedClient, b *eventbus.InMemoryBus) interface{} { return NewEthereumRepository(c, b) }},
		{"bitcoin", func(c *rpc.SimulatedClient, b *eventbus.InMemoryBus) interface{} { return NewBitcoinRepository(c, b) }},
		{"solana", func(c *rpc.SimulatedClient, b *eventbus.InMemoryBus) interface{} { return NewSolanaRepository(c, b) }},
		{"tron", func(c *rpc.SimulatedClient, b *eventbus.InMemoryBus) interface{} { return NewTronRepository(c, b) }},
	}

	for _, r := range repos {
		c := rpc.NewSimulatedClient()
		clients[r.name] = c
		repoIface := r.repo(c, bus)

		// type assertion to Blockchain-like behaviour
		switch rep := repoIface.(type) {
		case *EthereumRepository:
			runRepoTests(t, rep)
		case *BitcoinRepository:
			runRepoTests(t, rep)
		case *SolanaRepository:
			runRepoTests(t, rep)
		case *TronRepository:
			runRepoTests(t, rep)
		default:
			t.Fatalf("unknown repo type")
		}
	}
}

func runRepoTests(t *testing.T, repo interface{}) {
	ctx := context.Background()
	var r *BaseRepo
	switch v := repo.(type) {
	case *EthereumRepository:
		r = &v.BaseRepo
	case *BitcoinRepository:
		r = &v.BaseRepo
	case *SolanaRepository:
		r = &v.BaseRepo
	case *TronRepository:
		r = &v.BaseRepo
	}
	if err := r.Connect(ctx); err != nil {
		t.Fatalf("connect err: %v", err)
	}

	// balance missing: expect an error
	if _, err := r.GetBalance(ctx, "noaddr"); err == nil {
		t.Fatalf("expected error for missing balance")
	}

	// broadcast
	amt, _ := domain.NewAmountFromString("2")
	from, _ := domain.NewAddress("addr_from")
	to, _ := domain.NewAddress("addr_to")
	tx, _ := r.CreateTransaction(ctx, from, to, amt)
	if tx.Hash == "" {
		t.Fatalf("expected hash")
	}

	if _, err := r.BroadcastTransaction(ctx, tx); err != nil {
		t.Fatalf("broadcast err: %v", err)
	}

	// get tx
	got, err := r.GetTransaction(ctx, domain.TxHash(tx.Hash))
	if err != nil {
		t.Fatalf("get tx err: %v", err)
	}
	if got.Hash == "" {
		t.Fatalf("bad tx")
	}

	// latest block
	r.client.AdvanceBlock(100)
	blk, _ := r.GetLatestBlock(ctx)
	if blk.Number != 100 {
		t.Fatalf("expected block 100")
	}
}

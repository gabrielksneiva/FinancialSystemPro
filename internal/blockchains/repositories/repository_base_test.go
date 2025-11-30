package repositories

import (
	"context"
	"testing"
	"time"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/events"
	"financial-system-pro/internal/blockchains/infra/eventbus"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

func TestBaseRepo_CreateTxAndEvents(t *testing.T) {
	client := rpc.NewSimulatedClient()
	bus := eventbus.NewInMemoryBus()
	repo := NewEthereumRepository(client, bus)
	ctx := context.Background()
	if err := repo.Connect(ctx); err != nil {
		t.Fatalf("connect err: %v", err)
	}

	amt, _ := domain.NewAmountFromString("123")
	from, _ := domain.NewAddress("addr_from")
	to, _ := domain.NewAddress("addr_to")

	// subscribe to new tx events
	chTx, cancel := bus.Subscribe(ctx, "NewTransaction")
	defer cancel()

	tx, err := repo.CreateTransaction(ctx, from, to, amt)
	if err != nil {
		t.Fatalf("create tx err: %v", err)
	}
	if tx.Hash == "" {
		t.Fatalf("expected hash")
	}

	select {
	case v := <-chTx:
		if v == nil {
			t.Fatalf("expected event payload")
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting tx event")
	}

	// confirm via client
	client.ConfirmTransaction(tx.Hash, 5)
	status, _ := repo.GetTransactionStatus(ctx, domain.TxHash(tx.Hash))
	if status != domain.TxConfirmed {
		t.Fatalf("expected confirmed")
	}
}

func TestBaseRepo_WatchBlocksReceivesEvents(t *testing.T) {
	client := rpc.NewSimulatedClient()
	bus := eventbus.NewInMemoryBus()
	repo := NewEthereumRepository(client, bus)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := repo.WatchBlocks(ctx)
	if err != nil {
		t.Fatalf("watch blocks err: %v", err)
	}

	// publish a block
	bus.Publish("NewBlock", "blk1", events.NewBlockEvent{Chain: "ethereum", BlockNumber: 99})

	select {
	case v := <-ch:
		if _, ok := v.(events.NewBlockEvent); !ok {
			t.Fatalf("unexpected type")
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting block event")
	}
}

func TestBaseRepo_ConfirmPublishesEvents(t *testing.T) {
	client := rpc.NewSimulatedClient()
	bus := eventbus.NewInMemoryBus()
	repo := NewEthereumRepository(client, bus)

	// subscribe to block and tx confirm events
	chBlk, cancelBlk := bus.Subscribe(context.Background(), events.EventNewBlock)
	defer cancelBlk()
	chTx, cancelTx := bus.Subscribe(context.Background(), events.EventTransactionConfirm)
	defer cancelTx()

	// confirm a tx
	tx := domain.Transaction{Hash: "tx_1", From: "a", To: "b", Amount: domain.Amount{Value: nil}}
	// seed tx into client
	if _, err := client.SubmitTransaction(context.Background(), tx); err != nil {
		t.Fatalf("seed tx err: %v", err)
	}

	// call repo confirm which should publish events
	_ = repo.ConfirmTransaction(domain.TxHash("tx_1"), 7)

	select {
	case v := <-chBlk:
		if _, ok := v.(events.NewBlockEvent); !ok {
			t.Fatalf("expected NewBlockEvent")
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting new block event")
	}

	select {
	case v := <-chTx:
		if _, ok := v.(events.TransactionConfirmedEvent); !ok {
			t.Fatalf("expected TransactionConfirmedEvent")
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting tx confirm event")
	}
}

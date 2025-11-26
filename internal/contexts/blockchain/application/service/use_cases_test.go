package service

import (
	"context"
	"sync/atomic"
	"testing"

	app "financial-system-pro/internal/contexts/blockchain/application"
	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"
	repo "financial-system-pro/internal/contexts/blockchain/domain/repository"
	"financial-system-pro/internal/contexts/blockchain/infrastructure/gateway"
	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type fakeRepo struct{}

func (f *fakeRepo) Create(ctx context.Context, _ *entity.BlockchainTransaction) error { return nil }
func (f *fakeRepo) FindByID(ctx context.Context, _ uuid.UUID) (*entity.BlockchainTransaction, error) {
	return nil, nil
}
func (f *fakeRepo) FindByHash(ctx context.Context, _ string) (*entity.BlockchainTransaction, error) {
	return nil, nil
}
func (f *fakeRepo) FindByAddress(ctx context.Context, _ string) ([]*entity.BlockchainTransaction, error) {
	return nil, nil
}
func (f *fakeRepo) Update(ctx context.Context, _ *entity.BlockchainTransaction) error { return nil }
func (f *fakeRepo) UpdateConfirmations(ctx context.Context, _ string, _ int) error    { return nil }

var _ repo.BlockchainTransactionRepository = (*fakeRepo)(nil)

func TestUseCases_Flow(t *testing.T) {
	// Registry with ETH gateway
	reg := app.NewBlockchainRegistry(gateway.NewETHGatewayFromEnv())
	bus := events.NewInMemoryBus(zap.NewNop())
	var newTxCount int32
	var confirmedCount int32
	var newBlockCount int32
	bus.Subscribe("tx.new", func(ctx context.Context, e events.Event) error { atomic.AddInt32(&newTxCount, 1); return nil })
	bus.Subscribe("blockchain.transaction.confirmed", func(ctx context.Context, e events.Event) error { atomic.AddInt32(&confirmedCount, 1); return nil })
	bus.Subscribe("block.new", func(ctx context.Context, e events.Event) error { atomic.AddInt32(&newBlockCount, 1); return nil })

	uc := NewUseCases(reg, &fakeRepo{}, bus)

	// Balance
	w, _ := gateway.NewETHGatewayFromEnv().GenerateWallet(context.Background())
	bal, err := uc.FetchBalance(context.Background(), entity.BlockchainEthereum, w.Address)
	if err != nil || bal <= 0 {
		t.Fatalf("balance error: %v", err)
	}

	// Send tx
	hash, err := uc.SendTransaction(context.Background(), entity.BlockchainEthereum, w.Address, w.Address, 123, w.PrivateKey)
	if err != nil || len(hash) == 0 {
		t.Fatalf("send tx error: %v", err)
	}
	if atomic.LoadInt32(&newTxCount) == 0 {
		t.Fatalf("expected tx.new event")
	}

	// Status
	_, _ = uc.GetTransactionStatus(context.Background(), entity.BlockchainEthereum, hash)
	if atomic.LoadInt32(&confirmedCount) == 0 {
		t.Fatalf("expected confirmed event")
	}

	// Sync blocks
	_ = uc.SyncLatestBlocks(context.Background(), entity.BlockchainEthereum)
	if atomic.LoadInt32(&newBlockCount) == 0 {
		t.Fatalf("expected block.new event")
	}
}

package repositories

import (
	"context"

	"financial-system-pro/internal/blockchains/infra/eventbus"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

type BitcoinRepository struct {
	BaseRepo
}

func NewBitcoinRepository(client rpc.Client, bus *eventbus.InMemoryBus) *BitcoinRepository {
	return &BitcoinRepository{BaseRepo{Chain: "bitcoin", client: client, bus: bus}}
}

func (b *BitcoinRepository) Connect(ctx context.Context) error {
	return b.BaseRepo.Connect(ctx)
}

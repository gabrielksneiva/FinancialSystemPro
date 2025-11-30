package repositories

import (
	"context"

	"financial-system-pro/internal/blockchains/infra/eventbus"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

type EthereumRepository struct {
	BaseRepo
}

func NewEthereumRepository(client rpc.Client, bus *eventbus.InMemoryBus) *EthereumRepository {
	return &EthereumRepository{BaseRepo{Chain: "ethereum", client: client, bus: bus}}
}

func (e *EthereumRepository) Connect(ctx context.Context) error {
	return e.BaseRepo.Connect(ctx)
}

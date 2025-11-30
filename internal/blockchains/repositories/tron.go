package repositories

import (
	"context"

	"financial-system-pro/internal/blockchains/infra/eventbus"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

type TronRepository struct {
	BaseRepo
}

func NewTronRepository(client rpc.Client, bus *eventbus.InMemoryBus) *TronRepository {
	return &TronRepository{BaseRepo{Chain: "tron", client: client, bus: bus}}
}

func (t *TronRepository) Connect(ctx context.Context) error {
	return t.BaseRepo.Connect(ctx)
}

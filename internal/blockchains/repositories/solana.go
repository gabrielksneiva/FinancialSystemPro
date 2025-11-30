package repositories

import (
	"context"

	"financial-system-pro/internal/blockchains/infra/eventbus"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

type SolanaRepository struct {
	BaseRepo
}

func NewSolanaRepository(client rpc.Client, bus *eventbus.InMemoryBus) *SolanaRepository {
	return &SolanaRepository{BaseRepo{Chain: "solana", client: client, bus: bus}}
}

func (s *SolanaRepository) Connect(ctx context.Context) error {
	return s.BaseRepo.Connect(ctx)
}

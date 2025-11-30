package application

import (
	"context"

	"financial-system-pro/internal/blockchains/domain"
)

// BroadcasterPort is a port for broadcasting transactions
type BroadcasterPort interface {
	BroadcastTransaction(ctx context.Context, tx domain.Transaction) (domain.TxHash, error)
}

type BroadcastTransactionUseCase struct {
	repo BroadcasterPort
}

func NewBroadcastTransactionUseCase(repo BroadcasterPort) *BroadcastTransactionUseCase {
	return &BroadcastTransactionUseCase{repo: repo}
}

func (u *BroadcastTransactionUseCase) Execute(ctx context.Context, tx domain.Transaction) (domain.TxHash, error) {
	return u.repo.BroadcastTransaction(ctx, tx)
}

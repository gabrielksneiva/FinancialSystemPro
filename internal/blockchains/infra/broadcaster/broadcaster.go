package broadcaster

import (
	"context"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/infra/rpc"
)

// TxPersistence is a small port for persisting transactions.
type TxPersistence interface {
	CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
}

type TransactionBroadcaster struct {
	repo   TxPersistence
	client rpc.Client
}

func NewTransactionBroadcaster(repo TxPersistence, client rpc.Client) *TransactionBroadcaster {
	return &TransactionBroadcaster{repo: repo, client: client}
}

// BroadcastTransaction persists the tx and submits it to the RPC client.
func (b *TransactionBroadcaster) BroadcastTransaction(ctx context.Context, tx domain.Transaction) (domain.TxHash, error) {
	// persist first
	created, err := b.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return domain.TxHash(""), err
	}
	// broadcast via RPC client
	h, err := b.client.SubmitTransaction(ctx, created)
	if err != nil {
		return domain.TxHash(""), err
	}
	return domain.TxHash(h), nil
}

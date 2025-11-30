package repositories

import (
    "context"

    "financial-system-pro/internal/blockchains/domain"
)

type BlockchainRepository interface {
    Connect(ctx context.Context) error
    GetBalance(ctx context.Context, addr domain.Address) (domain.Amount, error)
    CreateTransaction(ctx context.Context, from domain.Address, to domain.Address, amount domain.Amount) (domain.Transaction, error)
    BroadcastTransaction(ctx context.Context, tx domain.Transaction) (domain.TxHash, error)
    GetTransactionStatus(ctx context.Context, hash domain.TxHash) (domain.TxStatus, error)
    GetTransaction(ctx context.Context, hash domain.TxHash) (domain.Transaction, error)
    GetLatestBlock(ctx context.Context) (domain.Block, error)
    WatchBlocks(ctx context.Context) (<-chan interface{}, error)
}

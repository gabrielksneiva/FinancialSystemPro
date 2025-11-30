package blockchain

import (
    "context"
    "time"

    "github.com/shopspring/decimal"
)

// Address represents a blockchain address.
type Address string

// TxHash represents a transaction hash.
type TxHash string

// Block represents a minimal block model used in syncs.
type Block struct {
    Number uint64
    Hash   string
    Time   time.Time
}

// Transaction minimal representation.
type Transaction struct {
    Hash          TxHash
    From          Address
    To            Address
    Value         decimal.Decimal
    Raw           string
    Confirmations int64
}

// Connector defines the common interface for blockchain connectors.
type Connector interface {
    FetchBalance(ctx context.Context, address Address) (decimal.Decimal, error)
    SendTransaction(ctx context.Context, rawTx string) (TxHash, error)
    GetTransactionStatus(ctx context.Context, hash TxHash) (string, error)
    SyncLatestBlocks(ctx context.Context, since uint64) ([]Block, error)
}

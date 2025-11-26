package domain

import (
	"context"
	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"
)

// BlockchainGatewayPort define operações genéricas para qualquer blockchain suportada.
// Migrated from legacy services package into blockchain domain context.
type BlockchainGatewayPort interface {
	GenerateWallet(ctx context.Context) (*entity.GeneratedWallet, error)
	ValidateAddress(address string) bool
	EstimateFee(ctx context.Context, fromAddress, toAddress string, amountBaseUnit int64) (*FeeQuote, error)
	Broadcast(ctx context.Context, fromAddress, toAddress string, amountBaseUnit int64, privateKey string) (TxHash, error)
	GetStatus(ctx context.Context, txHash TxHash) (*TxStatusInfo, error)
	ChainType() entity.BlockchainType

	// Novos métodos para integração multi-chain completa
	GetBalance(ctx context.Context, address string) (int64, error)
	GetTransactionHistory(ctx context.Context, address string, limit, offset int) ([]*entity.BlockchainTransaction, error)
	SubscribeNewBlocks(ctx context.Context, handler BlockEventHandler) error
	SubscribeNewTransactions(ctx context.Context, address string, handler TxEventHandler) error
}

// BlockEventHandler define callback para novos blocos
type BlockEventHandler func(blockNumber int64, blockHash string, timestamp int64) error

// TxEventHandler define callback para novas transações
type TxEventHandler func(tx *entity.BlockchainTransaction) error

type TxHash string

type FeeQuote struct {
	AmountBaseUnit int64  // valor principal em unidade mínima
	EstimatedFee   int64  // taxa estimada em base unit
	FeeAsset       string // ex: TRX, ETH, BTC
	Source         string // método ou gateway usado
}

type TxStatus string

const (
	TxStatusPending    TxStatus = "pending"
	TxStatusBroadcast  TxStatus = "broadcast"
	TxStatusInProgress TxStatus = "in_progress"
	TxStatusConfirmed  TxStatus = "confirmed"
	TxStatusFailed     TxStatus = "failed"
	TxStatusUnknown    TxStatus = "unknown"
)

type TxStatusInfo struct {
	Hash          TxHash   `json:"hash"`
	Status        TxStatus `json:"status"`
	Confirmations int64    `json:"confirmations"`
	Required      int64    `json:"required_confirmations"`
	ErrorMessage  string   `json:"error_message,omitempty"`
}

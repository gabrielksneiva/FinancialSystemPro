package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
)

// BlockchainGatewayPort define operações genéricas para qualquer blockchain suportada.
// Implementações: TronGateway (TronService), futuras EthereumGateway, BitcoinGateway, etc.
type BlockchainGatewayPort interface {
	GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error)
	ValidateAddress(address string) bool
	EstimateFee(ctx context.Context, fromAddress, toAddress string, amountBaseUnit int64) (*FeeQuote, error)
	Broadcast(ctx context.Context, fromAddress, toAddress string, amountBaseUnit int64, privateKey string) (TxHash, error)
	GetStatus(ctx context.Context, txHash TxHash) (*TxStatusInfo, error)
	ChainType() entities.BlockchainType
}

// TxHash representa um hash de transação genérico.
type TxHash string

// FeeQuote contém estimativa de custo (em unidade base).
type FeeQuote struct {
	AmountBaseUnit int64  // valor principal em unidade mínima (wei, satoshi, sun)
	EstimatedFee   int64  // taxa estimada em base unit
	FeeAsset       string // ex: TRX, ETH, BTC
	Source         string // método ou gateway usado
}

// TxStatus enum simplificado para status cross-chain.
type TxStatus string

const (
	TxStatusPending    TxStatus = "pending"
	TxStatusBroadcast  TxStatus = "broadcast"
	TxStatusInProgress TxStatus = "in_progress"
	TxStatusConfirmed  TxStatus = "confirmed"
	TxStatusFailed     TxStatus = "failed"
	TxStatusUnknown    TxStatus = "unknown"
)

// TxStatusInfo descreve estado de uma transação.
type TxStatusInfo struct {
	Hash          TxHash   `json:"hash"`
	Status        TxStatus `json:"status"`
	Confirmations int64    `json:"confirmations"`
	Required      int64    `json:"required_confirmations"`
	ErrorMessage  string   `json:"error_message,omitempty"`
}

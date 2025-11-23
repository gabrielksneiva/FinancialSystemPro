package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BlockchainNetwork representa a rede blockchain
type BlockchainNetwork string

const (
	NetworkTron BlockchainNetwork = "TRON"
)

// BlockchainTransaction representa uma transação na blockchain
type BlockchainTransaction struct {
	ID              uuid.UUID
	Network         BlockchainNetwork
	TransactionHash string
	FromAddress     string
	ToAddress       string
	Amount          decimal.Decimal
	Confirmations   int
	Status          string
	BlockNumber     int64
	GasUsed         int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ConfirmedAt     *time.Time
}

// WalletInfo representa informações de uma wallet na blockchain
type WalletInfo struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	Network          BlockchainNetwork
	Address          string
	EncryptedPrivKey string
	Balance          decimal.Decimal
	Nonce            int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewBlockchainTransaction cria uma nova transação blockchain
func NewBlockchainTransaction(network BlockchainNetwork, from, to string, amount decimal.Decimal) *BlockchainTransaction {
	return &BlockchainTransaction{
		ID:          uuid.New(),
		Network:     network,
		FromAddress: from,
		ToAddress:   to,
		Amount:      amount,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Confirm marca a transação como confirmada
func (bt *BlockchainTransaction) Confirm(txHash string, blockNumber int64, confirmations int) {
	now := time.Now()
	bt.TransactionHash = txHash
	bt.BlockNumber = blockNumber
	bt.Confirmations = confirmations
	bt.Status = "confirmed"
	bt.ConfirmedAt = &now
	bt.UpdatedAt = now
}

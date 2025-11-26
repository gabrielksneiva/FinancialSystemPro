package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BlockchainType representa o tipo de blockchain suportada
type BlockchainType string

const (
	BlockchainEthereum BlockchainType = "ethereum"
	BlockchainBitcoin  BlockchainType = "bitcoin"
	BlockchainTron     BlockchainType = "tron"
	BlockchainSolana   BlockchainType = "solana"
)

// GeneratedWallet representa uma wallet gerada para qualquer blockchain
type GeneratedWallet struct {
	Address    string
	PublicKey  string
	PrivateKey string // Encrypted ou vazio se não aplicável
	Blockchain BlockchainType
	CreatedAt  int64
	UserID     uuid.UUID
}

// BlockchainNetwork representa a rede blockchain
type BlockchainNetwork string

const (
	NetworkTron     BlockchainNetwork = "TRON"
	NetworkEthereum BlockchainNetwork = "ETHEREUM"
	NetworkBitcoin  BlockchainNetwork = "BITCOIN"
	NetworkSolana   BlockchainNetwork = "SOLANA"
)

// BlockchainTransaction representa uma transação na blockchain
type BlockchainTransaction struct {
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ConfirmedAt     *time.Time
	Network         BlockchainNetwork
	TransactionHash string
	FromAddress     string
	ToAddress       string
	Amount          decimal.Decimal
	Status          string
	Confirmations   int
	GasUsed         int64
	BlockNumber     int64
	ID              uuid.UUID
}

// WalletInfo representa informações de uma wallet na blockchain
type WalletInfo struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Network          BlockchainNetwork
	Address          string
	EncryptedPrivKey string
	Balance          decimal.Decimal
	Nonce            int64
	ID               uuid.UUID
	UserID           uuid.UUID
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

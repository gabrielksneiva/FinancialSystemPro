package entities

import "github.com/google/uuid"

// BlockchainType define os tipos de blockchain suportados
type BlockchainType string

const (
	BlockchainTRON     BlockchainType = "tron"
	BlockchainEthereum BlockchainType = "ethereum"
	BlockchainBitcoin  BlockchainType = "bitcoin"
)

// WalletInfo representa os dados de uma carteira
type WalletInfo struct {
	Address          string         `json:"address"`
	PublicKey        string         `json:"public_key"`
	PrivateKey       string         `json:"private_key,omitempty"`           // Private key em hex (apenas para geração)
	EncryptedPrivKey string         `json:"encrypted_private_key,omitempty"` // Private key criptografada para armazenamento
	Blockchain       BlockchainType `json:"blockchain"`
}

// GeneratedWallet é o resultado da geração de uma carteira
type GeneratedWallet struct {
	UserID     uuid.UUID      `json:"user_id"`
	Address    string         `json:"address"`
	PublicKey  string         `json:"public_key"`
	Blockchain BlockchainType `json:"blockchain"`
	CreatedAt  int64          `json:"created_at"`
}

// WalletManager interface para gerenciar carteiras de diferentes blockchains
type WalletManager interface {
	// GenerateWallet gera uma nova carteira e retorna address e public key
	GenerateWallet() (*WalletInfo, error)

	// ValidateAddress valida um endereço da blockchain
	ValidateAddress(address string) bool

	// GetBlockchainType retorna o tipo de blockchain que este manager gerencia
	GetBlockchainType() BlockchainType
}

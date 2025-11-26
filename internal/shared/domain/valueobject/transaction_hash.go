package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

// TransactionHash representa um hash de transação blockchain com validação
type TransactionHash struct {
	hash           string
	blockchainType string
}

var (
	// Regex para validação de transaction hashes
	tronTxHashRegex     = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	ethereumTxHashRegex = regexp.MustCompile(`^0x[a-fA-F0-9]{64}$`)
	bitcoinTxHashRegex  = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
)

// NewTransactionHash cria um novo hash de transação com validação
func NewTransactionHash(hash string, blockchainType string) (TransactionHash, error) {
	if hash == "" {
		return TransactionHash{}, errors.New("transaction hash cannot be empty")
	}

	if blockchainType == "" {
		return TransactionHash{}, errors.New("blockchain type is required")
	}

	// Normalizar blockchain type
	blockchainType = strings.ToLower(blockchainType)

	// Validar formato baseado no tipo de blockchain
	switch blockchainType {
	case "tron":
		if !tronTxHashRegex.MatchString(hash) {
			return TransactionHash{}, errors.New("invalid TRON transaction hash format")
		}
	case "ethereum":
		if !ethereumTxHashRegex.MatchString(hash) {
			return TransactionHash{}, errors.New("invalid Ethereum transaction hash format")
		}
	case "bitcoin":
		if !bitcoinTxHashRegex.MatchString(hash) {
			return TransactionHash{}, errors.New("invalid Bitcoin transaction hash format")
		}
	default:
		return TransactionHash{}, errors.New("unsupported blockchain type")
	}

	return TransactionHash{
		hash:           hash,
		blockchainType: blockchainType,
	}, nil
}

// Hash retorna o hash da transação
func (t TransactionHash) Hash() string {
	return t.hash
}

// BlockchainType retorna o tipo de blockchain
func (t TransactionHash) BlockchainType() string {
	return t.blockchainType
}

// String retorna representação em string
func (t TransactionHash) String() string {
	return t.hash
}

// Equals verifica igualdade
func (t TransactionHash) Equals(other TransactionHash) bool {
	return t.hash == other.hash && t.blockchainType == other.blockchainType
}

// ShortHash retorna versão abreviada do hash (primeiros 8 e últimos 8 caracteres)
func (t TransactionHash) ShortHash() string {
	if len(t.hash) < 16 {
		return t.hash
	}
	return t.hash[:8] + "..." + t.hash[len(t.hash)-8:]
}

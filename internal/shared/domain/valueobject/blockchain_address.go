package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

// BlockchainAddress representa um endereço de blockchain com validação
type BlockchainAddress struct {
	address        string
	blockchainType string
}

var (
	// Regex para validação de endereços
	tronAddressRegex     = regexp.MustCompile(`^T[A-Za-z1-9]{33}$`)
	ethereumAddressRegex = regexp.MustCompile(`^0x[a-fA-F0-9]{39,40}$`)
	bitcoinAddressRegex  = regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$|^bc1[a-z0-9]{39,59}$`)
)

// NewBlockchainAddress cria um novo endereço de blockchain com validação
func NewBlockchainAddress(address string, blockchainType string) (BlockchainAddress, error) {
	if address == "" {
		return BlockchainAddress{}, errors.New("address cannot be empty")
	}

	if blockchainType == "" {
		return BlockchainAddress{}, errors.New("blockchain type is required")
	}

	// Normalizar blockchain type
	blockchainType = strings.ToLower(blockchainType)

	// Validar formato baseado no tipo de blockchain
	switch blockchainType {
	case "tron":
		if !tronAddressRegex.MatchString(address) {
			return BlockchainAddress{}, errors.New("invalid TRON address format")
		}
	case "ethereum":
		if !ethereumAddressRegex.MatchString(address) {
			return BlockchainAddress{}, errors.New("invalid Ethereum address format")
		}
	case "bitcoin":
		if !bitcoinAddressRegex.MatchString(address) {
			return BlockchainAddress{}, errors.New("invalid Bitcoin address format")
		}
	default:
		return BlockchainAddress{}, errors.New("unsupported blockchain type")
	}

	return BlockchainAddress{
		address:        address,
		blockchainType: blockchainType,
	}, nil
}

// Address retorna o endereço
func (b BlockchainAddress) Address() string {
	return b.address
}

// BlockchainType retorna o tipo de blockchain
func (b BlockchainAddress) BlockchainType() string {
	return b.blockchainType
}

// String retorna representação em string
func (b BlockchainAddress) String() string {
	return b.address
}

// Equals verifica igualdade
func (b BlockchainAddress) Equals(other BlockchainAddress) bool {
	return b.address == other.address && b.blockchainType == other.blockchainType
}

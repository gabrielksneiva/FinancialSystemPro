package services

import (
	"financial-system-pro/internal/domain/entities"
	"fmt"

	"github.com/shopspring/decimal"
)

// BlockchainCategory agrupa blockchains por modelo operacional.
// account: saldo global por endereço (TRON, potencialmente Solana)
// evm: chains compatíveis EVM (Ethereum, futuros L2)
// utxo: chains baseadas em UTXO (Bitcoin, Litecoin, etc.)
type BlockchainCategory string

const (
	CategoryAccount BlockchainCategory = "account"
	CategoryEVM     BlockchainCategory = "evm"
	CategoryUTXO    BlockchainCategory = "utxo"
)

// DetermineCategory mapeia BlockchainType para categoria.
func DetermineCategory(chain entities.BlockchainType) BlockchainCategory {
	switch chain {
	case entities.BlockchainTRON:
		return CategoryAccount
	case entities.BlockchainEthereum:
		return CategoryEVM
	case entities.BlockchainBitcoin:
		return CategoryUTXO
	default:
		return "unknown"
	}
}

// ConvertAmountToBaseUnit converte decimal para unidade mínima conforme categoria.
// Mantém sem overflow (retorna erro se > int64).
func ConvertAmountToBaseUnit(chain entities.BlockchainType, amount decimal.Decimal) (int64, error) {
	var multiplier int64
	switch DetermineCategory(chain) {
	case CategoryAccount: // TRON (SUN)
		multiplier = 1_000_000
	case CategoryEVM: // Ethereum (Wei)
		// Ethereum usa 1e18; tratamos overflow antes de converter para int64
		// O código existente apenas suporta valores cabendo em int64 pós multiplicação.
		weiBig := amount.Mul(decimal.NewFromInt(1_000_000_000_000_000_000)).BigInt()
		if weiBig.BitLen() > 63 {
			return 0, fmt.Errorf("valor muito grande para conversão em wei")
		}
		return weiBig.Int64(), nil
	case CategoryUTXO: // Bitcoin (satoshis)
		multiplier = 100_000_000
	default:
		return 0, fmt.Errorf("unsupported blockchain: %s", chain)
	}
	satsBig := amount.Mul(decimal.NewFromInt(multiplier)).BigInt()
	if satsBig.BitLen() > 63 {
		// Reutiliza mesma mensagem usada nos testes (tron não testa overflow; bitcoin testa "valor muito grande")
		if DetermineCategory(chain) == CategoryUTXO {
			return 0, fmt.Errorf("valor muito grande para conversão em satoshis")
		}
		return 0, fmt.Errorf("valor muito grande para conversão")
	}
	return satsBig.Int64(), nil
}

// VaultEnvKeys retorna nomes das variáveis de ambiente de vault para chain EVM/UTXO.
// TRON usa service methods.
func VaultEnvKeys(chain entities.BlockchainType) (addrKey, privKey string) {
	switch chain {
	case entities.BlockchainEthereum:
		return "ETH_VAULT_ADDRESS", "ETH_VAULT_PRIVATE_KEY"
	case entities.BlockchainBitcoin:
		return "BTC_VAULT_ADDRESS", "BTC_VAULT_PRIVATE_KEY"
	default:
		return "", ""
	}
}

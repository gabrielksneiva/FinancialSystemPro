package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"financial-system-pro/internal/domain/entities"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// EthereumService implementa BlockchainGatewayPort com lógica simplificada (mock-friendly).
// Futuras integrações (RPC, gas oracle) substituirão as estimativas placeholder.
type EthereumService struct{}

// NewEthereumService constrói instância (sem dependências externas por enquanto).
func NewEthereumService() *EthereumService { return &EthereumService{} }

// ChainType identifica a blockchain ethereum.
func (e *EthereumService) ChainType() entities.BlockchainType { return entities.BlockchainEthereum }

// GenerateWallet gera chave secp256k1 e deriva endereço (0x + 20 bytes hex).
func (e *EthereumService) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar chave ethereum: %w", err)
	}
	addr := crypto.PubkeyToAddress(pk.PublicKey)
	pubBytes := crypto.FromECDSAPub(&pk.PublicKey)
	wallet := &entities.GeneratedWallet{
		Address:    addr.Hex(),
		PublicKey:  hex.EncodeToString(pubBytes),
		Blockchain: entities.BlockchainEthereum,
		CreatedAt:  time.Now().Unix(),
	}
	return wallet, nil
}

// ValidateAddress valida formato hex 0x + 40 chars.
func (e *EthereumService) ValidateAddress(a string) bool {
	if len(a) != 42 {
		return false
	}
	if a[:2] != "0x" {
		return false
	}
	// verificação simples hex
	for _, c := range a[2:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// EstimateFee retorna FeeQuote simplificada usando gas base 21k e gasPrice fictício 1 gwei.
// amountBaseUnit esperado em wei.
func (e *EthereumService) EstimateFee(ctx context.Context, from, to string, amountBaseUnit int64) (*FeeQuote, error) {
	// 1 gwei = 1e9 wei
	gasLimit := int64(21000)
	gasPrice := int64(1_000_000_000) // 1 gwei
	fee := gasLimit * gasPrice       // wei
	return &FeeQuote{AmountBaseUnit: amountBaseUnit, EstimatedFee: fee, FeeAsset: "ETH", Source: "eth_simple"}, nil
}

// Broadcast simula envio gerando hash determinístico dos parâmetros.
func (e *EthereumService) Broadcast(ctx context.Context, from, to string, amountBaseUnit int64, privateKey string) (TxHash, error) {
	if !e.ValidateAddress(from) || !e.ValidateAddress(to) {
		return "", fmt.Errorf("endereço inválido")
	}
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d|%s|eth", from, to, amountBaseUnit, privateKey)))
	return TxHash(hex.EncodeToString(h[:])), nil
}

// GetStatus sempre retorna confirmado (mock) para hash conhecido.
func (e *EthereumService) GetStatus(ctx context.Context, hash TxHash) (*TxStatusInfo, error) {
	if len(hash) == 0 {
		return &TxStatusInfo{Hash: hash, Status: TxStatusUnknown}, nil
	}
	return &TxStatusInfo{Hash: hash, Status: TxStatusConfirmed}, nil
}

// Utilitário opcional: converter ETH (decimal) para Wei.
func EthToWei(amount int64) *big.Int {
	// amount em ETH inteiro para simplificação
	wei := big.NewInt(amount)
	return wei.Mul(wei, big.NewInt(1_000_000_000_000_000_000))
}

var _ BlockchainGatewayPort = (*EthereumService)(nil)

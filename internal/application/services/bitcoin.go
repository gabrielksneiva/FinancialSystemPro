package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"financial-system-pro/internal/domain/entities"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/ripemd160"
)

// BitcoinService implementa BlockchainGatewayPort com lógica simplificada para testes.
// NOTA: Esta implementação NÃO envia transações reais; é adequada para TDD e arquitetura limpa.
type BitcoinService struct{}

// NewBitcoinService constrói instância sem dependências externas.
func NewBitcoinService() *BitcoinService { return &BitcoinService{} }

// ChainType identifica a blockchain bitcoin.
func (b *BitcoinService) ChainType() entities.BlockchainType { return entities.BlockchainBitcoin }

// GenerateWallet gera chave secp256k1 e endereço P2PKH (legacy) base58 começando com '1'.
func (b *BitcoinService) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar chave bitcoin: %w", err)
	}
	pubBytes := crypto.FromECDSAPub(&pk.PublicKey)
	addr := b.deriveP2PKH(pubBytes)
	wallet := &entities.GeneratedWallet{
		Address:    addr,
		PublicKey:  hex.EncodeToString(pubBytes),
		Blockchain: entities.BlockchainBitcoin,
		CreatedAt:  time.Now().Unix(),
	}
	return wallet, nil
}

// deriveP2PKH cria endereço legacy mainnet (versão 0x00) a partir da public key.
func (b *BitcoinService) deriveP2PKH(pub []byte) string {
	sha := sha256.Sum256(pub)
	rip := ripemd160.New()
	_, _ = rip.Write(sha[:])
	pkHash := rip.Sum(nil) // 20 bytes
	versioned := append([]byte{0x00}, pkHash...)
	// checksum: primeiro 4 bytes de double SHA256
	first := sha256.Sum256(versioned)
	second := sha256.Sum256(first[:])
	checksum := second[:4]
	full := append(versioned, checksum...)
	return base58.Encode(full)
}

// ValidateAddress valida formato simples base58 P2PKH/P2SH.
func (b *BitcoinService) ValidateAddress(a string) bool {
	if len(a) < 26 || len(a) > 42 { // range típico
		return false
	}
	if a[0] != '1' && a[0] != '3' { // P2PKH ou P2SH
		return false
	}
	// verificar caracteres base58
	for _, c := range a {
		if !(c >= '1' && c <= '9') && !(c >= 'A' && c <= 'Z') && !(c >= 'a' && c <= 'z') {
			return false
		}
		switch c { // excluir 0 O I l
		case '0', 'O', 'I', 'l':
			return false
		}
	}
	// decodificação para validar checksum (simplificada)
	decoded := base58.Decode(a)
	if len(decoded) < 25 { // 1 + 20 + 4
		return false
	}
	payload := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	expected := second[:4]
	for i := 0; i < 4; i++ {
		if checksum[i] != expected[i] {
			return false
		}
	}
	return true
}

// EstimateFee retorna estimativa simplificada: 250 bytes * 2 sat/byte.
func (b *BitcoinService) EstimateFee(ctx context.Context, from, to string, amountBaseUnit int64) (*FeeQuote, error) {
	fee := int64(250 * 2) // 500 satoshis
	return &FeeQuote{AmountBaseUnit: amountBaseUnit, EstimatedFee: fee, FeeAsset: "BTC", Source: "btc_simple"}, nil
}

// Broadcast simula envio gerando hash determinístico.
func (b *BitcoinService) Broadcast(ctx context.Context, from, to string, amountBaseUnit int64, privateKey string) (TxHash, error) {
	if !b.ValidateAddress(from) || !b.ValidateAddress(to) {
		return "", fmt.Errorf("endereço inválido")
	}
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d|%s|btc", from, to, amountBaseUnit, privateKey)))
	return TxHash(hex.EncodeToString(h[:])), nil
}

// GetStatus retorna confirmado para hash com tamanho >= 64, caso contrário unknown.
func (b *BitcoinService) GetStatus(ctx context.Context, hash TxHash) (*TxStatusInfo, error) {
	if len(hash) >= 64 {
		return &TxStatusInfo{Hash: hash, Status: TxStatusConfirmed}, nil
	}
	return &TxStatusInfo{Hash: hash, Status: TxStatusUnknown}, nil
}

var _ BlockchainGatewayPort = (*BitcoinService)(nil)

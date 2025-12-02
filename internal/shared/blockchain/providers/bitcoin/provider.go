package bitcoin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"financial-system-pro/internal/shared/blockchain"

	"github.com/shopspring/decimal"
)

// Provider Bitcoin (stub simplificado)
type Provider struct {
	network blockchain.NetworkType
}

func NewProvider(network blockchain.NetworkType) *Provider {
	return &Provider{network: network}
}

func (p *Provider) ChainType() blockchain.ChainType     { return blockchain.ChainBitcoin }
func (p *Provider) NetworkType() blockchain.NetworkType { return p.network }
func (p *Provider) IsHealthy(ctx context.Context) bool  { return true }

// GenerateWallet gera endereço simulado (não usa chaves reais - TODO: integrar btcsuite)
func (p *Provider) GenerateWallet(ctx context.Context) (*blockchain.Wallet, error) {
	// Simular geração de chave privada
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	privHex := hex.EncodeToString(b)
	// Endereço fake bc1 + 30 hex chars
	addr := "bc1" + privHex[:30]
	return &blockchain.Wallet{
		Address:    addr,
		PublicKey:  "pub" + privHex[:33],
		PrivateKey: privHex,
		ChainType:  blockchain.ChainBitcoin,
		CreatedAt:  time.Now(),
	}, nil
}

var (
	bech32Regex = regexp.MustCompile(`^(bc1|tb1|bcrt1)[a-z0-9]{20,60}$`)
	legacyRegex = regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
)

func (p *Provider) ValidateAddress(address string) error {
	if address == "" {
		return blockchain.NewValidationError("address", "empty")
	}
	address = strings.TrimSpace(address)
	if bech32Regex.MatchString(address) || legacyRegex.MatchString(address) {
		return nil
	}
	return blockchain.NewValidationError("address", "invalid bitcoin address format")
}

func (p *Provider) ImportWallet(ctx context.Context, privateKey string) (*blockchain.Wallet, error) {
	privateKey = strings.TrimPrefix(privateKey, "0x")
	if len(privateKey) < 32 {
		return nil, blockchain.NewValidationError("privateKey", "too short")
	}
	addr := "bc1" + privateKey[:30]
	return &blockchain.Wallet{
		Address:    addr,
		PublicKey:  "pub" + privateKey[:33],
		PrivateKey: privateKey,
		ChainType:  blockchain.ChainBitcoin,
		CreatedAt:  time.Now(),
	}, nil
}

func (p *Provider) GetBalance(ctx context.Context, address string) (*blockchain.Balance, error) {
	if err := p.ValidateAddress(address); err != nil {
		return nil, err
	}
	// Stub: retornar saldo pseudo aleatório
	rnd, _ := rand.Int(rand.Reader, big.NewInt(10_000_000))
	amount := decimal.NewFromInt(rnd.Int64()).Div(decimal.NewFromInt(100_000)) // até ~100 BTC
	return &blockchain.Balance{
		Address:       address,
		Amount:        amount,
		AmountRaw:     amount.Mul(decimal.NewFromInt(100_000_000)).String(),
		Currency:      "BTC",
		Decimals:      8,
		BlockNumber:   time.Now().Unix(),
		LastUpdatedAt: time.Now(),
	}, nil
}

func (p *Provider) EstimateFee(ctx context.Context, intent *blockchain.TransactionIntent) (*blockchain.FeeEstimate, error) {
	// Simples: base em satoshis por byte
	low := decimal.NewFromFloat(0.000010)  // 1k sat/vbyte
	med := decimal.NewFromFloat(0.000020)  // 2k sat/vbyte
	high := decimal.NewFromFloat(0.000040) // 4k sat/vbyte
	return &blockchain.FeeEstimate{
		ChainType:   blockchain.ChainBitcoin,
		Low:         low,
		Medium:      med,
		High:        high,
		Currency:    "BTC",
		EstimatedAt: time.Now(),
	}, nil
}

func (p *Provider) BuildTransaction(ctx context.Context, intent *blockchain.TransactionIntent) (*blockchain.UnsignedTransaction, error) {
	if err := p.ValidateAddress(intent.From); err != nil {
		return nil, err
	}
	if err := p.ValidateAddress(intent.To); err != nil {
		return nil, err
	}
	fee := decimal.NewFromFloat(0.00002)
	return &blockchain.UnsignedTransaction{
		ChainType: blockchain.ChainBitcoin,
		From:      intent.From,
		To:        intent.To,
		Amount:    intent.Amount,
		Fee:       fee,
		Nonce:     0,
		Data:      nil,
		RawData: map[string]interface{}{
			"model": "utxo-simplified",
		},
		CreatedAt: time.Now(),
	}, nil
}

func (p *Provider) SignTransaction(ctx context.Context, tx *blockchain.UnsignedTransaction, privateKey *blockchain.PrivateKey) (*blockchain.SignedTransaction, error) {
	// Placeholder assinatura
	sig := hex.EncodeToString(privateKey.Raw)[:64]
	txHash := hex.EncodeToString([]byte(fmt.Sprintf("%s:%s:%s", tx.From, tx.To, tx.Amount.String())))[:64]
	return &blockchain.SignedTransaction{
		ChainType: blockchain.ChainBitcoin,
		RawTx:     "010203...", // stub
		TxHash:    txHash,
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Signature: sig,
		RawData:   tx.RawData,
		SignedAt:  time.Now(),
	}, nil
}

func (p *Provider) BroadcastTransaction(ctx context.Context, tx *blockchain.SignedTransaction) (*blockchain.TransactionReceipt, error) {
	return &blockchain.TransactionReceipt{
		TxHash:        tx.TxHash,
		ChainType:     blockchain.ChainBitcoin,
		Status:        blockchain.TxStatusPending,
		From:          tx.From,
		To:            tx.To,
		Amount:        tx.Amount,
		Fee:           tx.Fee,
		Confirmations: 0,
		BroadcastAt:   time.Now(),
	}, nil
}

func (p *Provider) GetTransactionStatus(ctx context.Context, txHash string) (*blockchain.TransactionStatus, error) {
	// Stub: confirmar imediatamente
	return &blockchain.TransactionStatus{TxHash: txHash, Status: blockchain.TxStatusConfirmed}, nil
}

func (p *Provider) GetTransactionHistory(ctx context.Context, address string, opts *blockchain.PaginationOptions) (*blockchain.TransactionHistory, error) {
	return &blockchain.TransactionHistory{Address: address, Transactions: []blockchain.HistoricalTransaction{}, Total: 0, HasMore: false}, nil
}

func (p *Provider) SubscribeNewBlocks(ctx context.Context, handler blockchain.BlockHandler) error {
	return blockchain.ErrNotSupported
}
func (p *Provider) SubscribeNewTransactions(ctx context.Context, filter *blockchain.TxFilter, handler blockchain.TxHandler) error {
	return blockchain.ErrNotSupported
}
func (p *Provider) UnsubscribeAll(ctx context.Context) error { return nil }

func (p *Provider) GetCapabilities() *blockchain.ProviderCapabilities {
	return &blockchain.ProviderCapabilities{
		SupportsSmartContracts:   false,
		SupportsTokens:           false,
		SupportsStaking:          false,
		SupportsSubscriptions:    false,
		SupportsMemo:             false,
		SupportsMultiSig:         true,
		RequiresGas:              false,
		NativeTokenDecimals:      8,
		MinConfirmationsRequired: 6,
		AverageBlockTime:         10 * time.Minute,
	}
}

package tron

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"financial-system-pro/internal/shared/blockchain"

	"github.com/shopspring/decimal"
)

// Provider Tron (stub)
type Provider struct{ network blockchain.NetworkType }

func NewProvider(network blockchain.NetworkType) *Provider { return &Provider{network: network} }
func (p *Provider) ChainType() blockchain.ChainType        { return blockchain.ChainTron }
func (p *Provider) NetworkType() blockchain.NetworkType    { return p.network }
func (p *Provider) IsHealthy(ctx context.Context) bool     { return true }

func (p *Provider) GenerateWallet(ctx context.Context) (*blockchain.Wallet, error) {
	b := make([]byte, 32)
	rand.Read(b)
	addr := "T" + hex.EncodeToString(b)[:33]
	return &blockchain.Wallet{Address: addr, PublicKey: hex.EncodeToString(b[:16]), PrivateKey: hex.EncodeToString(b), ChainType: blockchain.ChainTron, CreatedAt: time.Now()}, nil
}

func (p *Provider) ValidateAddress(address string) error {
	if !strings.HasPrefix(address, "T") || len(address) < 30 {
		return blockchain.NewValidationError("address", "invalid tron address")
	}
	return nil
}

func (p *Provider) ImportWallet(ctx context.Context, privateKey string) (*blockchain.Wallet, error) {
	privateKey = strings.TrimPrefix(privateKey, "0x")
	if len(privateKey) < 32 {
		return nil, blockchain.NewValidationError("privateKey", "too short")
	}
	addr := "T" + privateKey[:33]
	return &blockchain.Wallet{Address: addr, PublicKey: privateKey[:40], PrivateKey: privateKey, ChainType: blockchain.ChainTron, CreatedAt: time.Now()}, nil
}

func (p *Provider) GetBalance(ctx context.Context, address string) (*blockchain.Balance, error) {
	if err := p.ValidateAddress(address); err != nil {
		return nil, err
	}
	amount := decimal.NewFromFloat(987.654)
	return &blockchain.Balance{Address: address, Amount: amount, AmountRaw: amount.Mul(decimal.NewFromInt(1_000_000)).String(), Currency: "TRX", Decimals: 6, BlockNumber: time.Now().Unix(), LastUpdatedAt: time.Now()}, nil
}

func (p *Provider) EstimateFee(ctx context.Context, intent *blockchain.TransactionIntent) (*blockchain.FeeEstimate, error) {
	low := decimal.NewFromFloat(0.1)
	med := decimal.NewFromFloat(0.3)
	high := decimal.NewFromFloat(0.5)
	return &blockchain.FeeEstimate{ChainType: blockchain.ChainTron, Low: low, Medium: med, High: high, Currency: "TRX", EstimatedAt: time.Now()}, nil
}

func (p *Provider) BuildTransaction(ctx context.Context, intent *blockchain.TransactionIntent) (*blockchain.UnsignedTransaction, error) {
	if err := p.ValidateAddress(intent.From); err != nil {
		return nil, err
	}
	if err := p.ValidateAddress(intent.To); err != nil {
		return nil, err
	}
	fee := decimal.NewFromFloat(0.3)
	return &blockchain.UnsignedTransaction{ChainType: blockchain.ChainTron, From: intent.From, To: intent.To, Amount: intent.Amount, Fee: fee, Nonce: 0, RawData: map[string]interface{}{"note": "energy-model"}, CreatedAt: time.Now()}, nil
}

func (p *Provider) SignTransaction(ctx context.Context, tx *blockchain.UnsignedTransaction, privateKey *blockchain.PrivateKey) (*blockchain.SignedTransaction, error) {
	sig := hex.EncodeToString(privateKey.Raw)[:64]
	return &blockchain.SignedTransaction{ChainType: blockchain.ChainTron, RawTx: "tronraw...", TxHash: sig, From: tx.From, To: tx.To, Amount: tx.Amount, Fee: tx.Fee, Signature: sig, RawData: tx.RawData, SignedAt: time.Now()}, nil
}

func (p *Provider) BroadcastTransaction(ctx context.Context, tx *blockchain.SignedTransaction) (*blockchain.TransactionReceipt, error) {
	return &blockchain.TransactionReceipt{TxHash: tx.TxHash, ChainType: blockchain.ChainTron, Status: blockchain.TxStatusPending, From: tx.From, To: tx.To, Amount: tx.Amount, Fee: tx.Fee, BroadcastAt: time.Now()}, nil
}

func (p *Provider) GetTransactionStatus(ctx context.Context, txHash string) (*blockchain.TransactionStatus, error) {
	return &blockchain.TransactionStatus{TxHash: txHash, Status: blockchain.TxStatusConfirmed}, nil
}
func (p *Provider) GetTransactionHistory(ctx context.Context, address string, opts *blockchain.PaginationOptions) (*blockchain.TransactionHistory, error) {
	return &blockchain.TransactionHistory{Address: address}, nil
}
func (p *Provider) SubscribeNewBlocks(ctx context.Context, handler blockchain.BlockHandler) error {
	return blockchain.ErrNotSupported
}
func (p *Provider) SubscribeNewTransactions(ctx context.Context, filter *blockchain.TxFilter, handler blockchain.TxHandler) error {
	return blockchain.ErrNotSupported
}
func (p *Provider) UnsubscribeAll(ctx context.Context) error { return nil }

func (p *Provider) GetCapabilities() *blockchain.ProviderCapabilities {
	return &blockchain.ProviderCapabilities{SupportsSmartContracts: true, SupportsTokens: true, SupportsStaking: false, SupportsSubscriptions: false, SupportsMemo: false, SupportsMultiSig: false, RequiresGas: false, NativeTokenDecimals: 6, MinConfirmationsRequired: 1, AverageBlockTime: 3 * time.Second}
}

package solana

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"time"

	"financial-system-pro/internal/shared/blockchain"

	"github.com/shopspring/decimal"
)

// Provider Solana (stub)
type Provider struct {
	network blockchain.NetworkType
}

func NewProvider(network blockchain.NetworkType) *Provider { return &Provider{network: network} }
func (p *Provider) ChainType() blockchain.ChainType        { return blockchain.ChainSolana }
func (p *Provider) NetworkType() blockchain.NetworkType    { return p.network }
func (p *Provider) IsHealthy(ctx context.Context) bool     { return true }

// Solana address base58 regex simplificado
// Solana endereços base58, aqui simplificação: aceitar hex gerado (compat para stub) ou base58
var addrRegex = regexp.MustCompile(`^[0-9a-f]{32,64}$|^[1-9A-HJ-NP-Za-km-z]{32,44}$`)

func (p *Provider) GenerateWallet(ctx context.Context) (*blockchain.Wallet, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	// Representar public key como base64 (stub) e endereço como hex truncado
	address := hex.EncodeToString(b) // usar hex completo para satisfazer regex hex
	return &blockchain.Wallet{
		Address:    address,
		PublicKey:  base64.StdEncoding.EncodeToString(b),
		PrivateKey: hex.EncodeToString(b),
		ChainType:  blockchain.ChainSolana,
		CreatedAt:  time.Now(),
	}, nil
}

func (p *Provider) ValidateAddress(address string) error {
	if !addrRegex.MatchString(address) {
		return blockchain.NewValidationError("address", "invalid solana address format")
	}
	return nil
}

func (p *Provider) ImportWallet(ctx context.Context, privateKey string) (*blockchain.Wallet, error) {
	// privateKey esperado como hex de 32 bytes
	pkBytes, err := hex.DecodeString(privateKey)
	if err != nil || len(pkBytes) != 32 {
		return nil, blockchain.NewValidationError("privateKey", "invalid hex length")
	}
	// Gerencia igual GenerateWallet: endereço = privateKey hex
	address := privateKey
	return &blockchain.Wallet{
		Address:    address,
		PublicKey:  base64.StdEncoding.EncodeToString(pkBytes),
		PrivateKey: privateKey,
		ChainType:  blockchain.ChainSolana,
		CreatedAt:  time.Now(),
	}, nil
}

func (p *Provider) GetBalance(ctx context.Context, address string) (*blockchain.Balance, error) {
	if err := p.ValidateAddress(address); err != nil {
		return nil, err
	}
	amount := decimal.NewFromFloat(123.456) // Stub
	return &blockchain.Balance{
		Address:       address,
		Amount:        amount,
		AmountRaw:     amount.Mul(decimal.NewFromInt(1_000_000_000)).String(),
		Currency:      "SOL",
		Decimals:      9,
		BlockNumber:   time.Now().Unix(),
		LastUpdatedAt: time.Now(),
	}, nil
}

func (p *Provider) EstimateFee(ctx context.Context, intent *blockchain.TransactionIntent) (*blockchain.FeeEstimate, error) {
	low := decimal.NewFromFloat(0.000005)
	med := decimal.NewFromFloat(0.000010)
	high := decimal.NewFromFloat(0.000020)
	return &blockchain.FeeEstimate{ChainType: blockchain.ChainSolana, Low: low, Medium: med, High: high, Currency: "SOL", EstimatedAt: time.Now()}, nil
}

func (p *Provider) BuildTransaction(ctx context.Context, intent *blockchain.TransactionIntent) (*blockchain.UnsignedTransaction, error) {
	if err := p.ValidateAddress(intent.From); err != nil {
		return nil, err
	}
	if err := p.ValidateAddress(intent.To); err != nil {
		return nil, err
	}
	fee := decimal.NewFromFloat(0.00001)
	return &blockchain.UnsignedTransaction{ChainType: blockchain.ChainSolana, From: intent.From, To: intent.To, Amount: intent.Amount, Fee: fee, Nonce: 0, RawData: map[string]interface{}{"program": "system"}, CreatedAt: time.Now()}, nil
}

func (p *Provider) SignTransaction(ctx context.Context, tx *blockchain.UnsignedTransaction, privateKey *blockchain.PrivateKey) (*blockchain.SignedTransaction, error) {
	sig := hex.EncodeToString(privateKey.Raw)[:64]
	return &blockchain.SignedTransaction{ChainType: blockchain.ChainSolana, RawTx: "base64tx...", TxHash: sig, From: tx.From, To: tx.To, Amount: tx.Amount, Fee: tx.Fee, Signature: sig, RawData: tx.RawData, SignedAt: time.Now()}, nil
}

func (p *Provider) BroadcastTransaction(ctx context.Context, tx *blockchain.SignedTransaction) (*blockchain.TransactionReceipt, error) {
	return &blockchain.TransactionReceipt{TxHash: tx.TxHash, ChainType: blockchain.ChainSolana, Status: blockchain.TxStatusPending, From: tx.From, To: tx.To, Amount: tx.Amount, Fee: tx.Fee, BroadcastAt: time.Now()}, nil
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
	return &blockchain.ProviderCapabilities{SupportsSmartContracts: true, SupportsTokens: true, SupportsStaking: true, SupportsSubscriptions: false, SupportsMemo: true, SupportsMultiSig: false, RequiresGas: false, NativeTokenDecimals: 9, MinConfirmationsRequired: 1, AverageBlockTime: 400 * time.Millisecond}
}

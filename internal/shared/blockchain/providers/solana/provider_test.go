package solana_test

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"financial-system-pro/internal/shared/blockchain"
	"financial-system-pro/internal/shared/blockchain/providers/solana"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSolanaProvider_Wallet(t *testing.T) {
	p := solana.NewProvider(blockchain.NetworkDevnet)
	w, err := p.GenerateWallet(context.Background())
	require.NoError(t, err)
	assert.NoError(t, p.ValidateAddress(w.Address))
	w2, err := p.ImportWallet(context.Background(), w.PrivateKey)
	require.NoError(t, err)
	assert.Equal(t, w.Address, w2.Address)
}

func TestSolanaProvider_Balance(t *testing.T) {
	p := solana.NewProvider(blockchain.NetworkDevnet)
	w, _ := p.GenerateWallet(context.Background())
	b, err := p.GetBalance(context.Background(), w.Address)
	require.NoError(t, err)
	assert.Equal(t, "SOL", b.Currency)
}

func TestSolanaProvider_TxFlow(t *testing.T) {
	p := solana.NewProvider(blockchain.NetworkDevnet)
	wFrom, _ := p.GenerateWallet(context.Background())
	wTo, _ := p.GenerateWallet(context.Background())
	intent := blockchain.TransactionIntent{From: wFrom.Address, To: wTo.Address, Amount: decimal.NewFromFloat(2.5)}
	_, err := p.EstimateFee(context.Background(), &intent)
	require.NoError(t, err)
	unsigned, err := p.BuildTransaction(context.Background(), &intent)
	require.NoError(t, err)
	pkBytes, _ := hex.DecodeString(wFrom.PrivateKey)
	priv := blockchain.PrivateKey{Raw: pkBytes}
	signed, err := p.SignTransaction(context.Background(), unsigned, &priv)
	require.NoError(t, err)
	receipt, err := p.BroadcastTransaction(context.Background(), signed)
	require.NoError(t, err)
	assert.Equal(t, blockchain.TxStatusPending, receipt.Status)
	status, err := p.GetTransactionStatus(context.Background(), signed.TxHash)
	require.NoError(t, err)
	assert.Equal(t, blockchain.TxStatusConfirmed, status.Status)
}

func TestSolanaProvider_Capabilities(t *testing.T) {
	p := solana.NewProvider(blockchain.NetworkDevnet)
	caps := p.GetCapabilities()
	assert.True(t, caps.SupportsSmartContracts)
	assert.True(t, caps.SupportsTokens)
	assert.True(t, caps.SupportsStaking)
}

func TestSolanaProvider_History(t *testing.T) {
	p := solana.NewProvider(blockchain.NetworkDevnet)
	w, _ := p.GenerateWallet(context.Background())
	h, err := p.GetTransactionHistory(context.Background(), w.Address, &blockchain.PaginationOptions{Limit: 5})
	require.NoError(t, err)
	assert.Equal(t, w.Address, h.Address)
}

func TestSolanaProvider_Health(t *testing.T) {
	p := solana.NewProvider(blockchain.NetworkDevnet)
	assert.True(t, p.IsHealthy(context.Background()))
}

func TestSolanaProvider_Timeout(t *testing.T) {
	p := solana.NewProvider(blockchain.NetworkDevnet)
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	_ = p.IsHealthy(ctx)
}

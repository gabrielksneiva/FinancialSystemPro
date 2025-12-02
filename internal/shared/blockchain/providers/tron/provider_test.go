package tron_test

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"financial-system-pro/internal/shared/blockchain"
	"financial-system-pro/internal/shared/blockchain/providers/tron"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTronProvider_WalletValidation(t *testing.T) {
	p := tron.NewProvider(blockchain.NetworkTestnet)
	w, err := p.GenerateWallet(context.Background())
	require.NoError(t, err)
	assert.NoError(t, p.ValidateAddress(w.Address))
	err = p.ValidateAddress("X123")
	assert.Error(t, err)
}

func TestTronProvider_Balance(t *testing.T) {
	p := tron.NewProvider(blockchain.NetworkTestnet)
	w, _ := p.GenerateWallet(context.Background())
	b, err := p.GetBalance(context.Background(), w.Address)
	require.NoError(t, err)
	assert.Equal(t, "TRX", b.Currency)
}

func TestTronProvider_TxLifecycle(t *testing.T) {
	p := tron.NewProvider(blockchain.NetworkTestnet)
	wFrom, _ := p.GenerateWallet(context.Background())
	wTo, _ := p.GenerateWallet(context.Background())
	intent := blockchain.TransactionIntent{From: wFrom.Address, To: wTo.Address, Amount: decimal.NewFromFloat(10)}
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

func TestTronProvider_Capabilities(t *testing.T) {
	p := tron.NewProvider(blockchain.NetworkTestnet)
	caps := p.GetCapabilities()
	assert.True(t, caps.SupportsSmartContracts)
	assert.True(t, caps.SupportsTokens)
}

func TestTronProvider_History(t *testing.T) {
	p := tron.NewProvider(blockchain.NetworkTestnet)
	w, _ := p.GenerateWallet(context.Background())
	h, err := p.GetTransactionHistory(context.Background(), w.Address, &blockchain.PaginationOptions{Limit: 5})
	require.NoError(t, err)
	assert.Equal(t, w.Address, h.Address)
}

func TestTronProvider_HealthAndTimeout(t *testing.T) {
	p := tron.NewProvider(blockchain.NetworkTestnet)
	assert.True(t, p.IsHealthy(context.Background()))
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	_ = p.IsHealthy(ctx)
}

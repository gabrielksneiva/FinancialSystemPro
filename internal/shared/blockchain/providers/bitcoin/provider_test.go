package bitcoin_test

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"financial-system-pro/internal/shared/blockchain"
	"financial-system-pro/internal/shared/blockchain/providers/bitcoin"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBitcoinProvider_WalletAndValidation(t *testing.T) {
	p := bitcoin.NewProvider(blockchain.NetworkMainnet)
	w, err := p.GenerateWallet(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, w.Address)
	require.NoError(t, p.ValidateAddress(w.Address))
	// Import
	w2, err := p.ImportWallet(context.Background(), w.PrivateKey)
	require.NoError(t, err)
	assert.Equal(t, w.Address, w2.Address)
}

func TestBitcoinProvider_Balance(t *testing.T) {
	p := bitcoin.NewProvider(blockchain.NetworkMainnet)
	w, _ := p.GenerateWallet(context.Background())
	b, err := p.GetBalance(context.Background(), w.Address)
	require.NoError(t, err)
	assert.Equal(t, "BTC", b.Currency)
	assert.True(t, b.Amount.GreaterThan(decimal.Zero))
}

func TestBitcoinProvider_FeeAndBuildSignBroadcast(t *testing.T) {
	p := bitcoin.NewProvider(blockchain.NetworkMainnet)
	wFrom, _ := p.GenerateWallet(context.Background())
	wTo, _ := p.GenerateWallet(context.Background())
	intent := blockchain.TransactionIntent{From: wFrom.Address, To: wTo.Address, Amount: decimal.NewFromFloat(0.1234)}
	fee, err := p.EstimateFee(context.Background(), &intent)
	require.NoError(t, err)
	assert.Equal(t, blockchain.ChainBitcoin, fee.ChainType)
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

func TestBitcoinProvider_Capabilities(t *testing.T) {
	p := bitcoin.NewProvider(blockchain.NetworkTestnet)
	caps := p.GetCapabilities()
	assert.False(t, caps.SupportsSmartContracts)
	assert.True(t, caps.SupportsMultiSig)
	assert.Equal(t, 8, caps.NativeTokenDecimals)
}

func TestBitcoinProvider_ValidateInvalid(t *testing.T) {
	p := bitcoin.NewProvider(blockchain.NetworkMainnet)
	err := p.ValidateAddress("invalidXXX")
	assert.Error(t, err)
}

func TestBitcoinProvider_Health(t *testing.T) {
	p := bitcoin.NewProvider(blockchain.NetworkMainnet)
	assert.True(t, p.IsHealthy(context.Background()))
}

func TestBitcoinProvider_History(t *testing.T) {
	p := bitcoin.NewProvider(blockchain.NetworkMainnet)
	w, _ := p.GenerateWallet(context.Background())
	h, err := p.GetTransactionHistory(context.Background(), w.Address, &blockchain.PaginationOptions{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, 0, h.Total)
}

func TestBitcoinProvider_TimeoutContext(t *testing.T) {
	p := bitcoin.NewProvider(blockchain.NetworkMainnet)
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	// Even with timeout stub returns quickly; ensure no panic
	_, _ = p.IsHealthy(ctx), p.NetworkType()
}

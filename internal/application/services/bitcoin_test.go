package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitcoinService_GenerateWallet(t *testing.T) {
	svc := NewBitcoinService()
	w, err := svc.GenerateWallet(context.Background())
	require.NoError(t, err)
	require.Equal(t, entities.BlockchainBitcoin, w.Blockchain)
	require.True(t, svc.ValidateAddress(w.Address))
	require.NotEmpty(t, w.PublicKey)
}

func TestBitcoinService_ValidateAddress_Invalid(t *testing.T) {
	svc := NewBitcoinService()
	require.False(t, svc.ValidateAddress("xyz"))
	require.False(t, svc.ValidateAddress("1O0Il123")) // contém chars proibidos
}

func TestBitcoinService_EstimateFee(t *testing.T) {
	svc := NewBitcoinService()
	quote, err := svc.EstimateFee(context.Background(), "1BitcoinEaterAddressDontSendf59kuE", "1BitcoinEaterAddressDontSendf59kuF", 1500)
	require.NoError(t, err)
	require.Equal(t, int64(1500), quote.AmountBaseUnit)
	require.Equal(t, "BTC", quote.FeeAsset)
	require.Equal(t, int64(500), quote.EstimatedFee)
}

func TestBitcoinService_BroadcastAndStatus(t *testing.T) {
	svc := NewBitcoinService()
	wFrom, _ := svc.GenerateWallet(context.Background())
	wTo, _ := svc.GenerateWallet(context.Background())
	txHash, err := svc.Broadcast(context.Background(), wFrom.Address, wTo.Address, 12345, "priv")
	require.NoError(t, err)
	require.Len(t, string(txHash), 64)
	st, sErr := svc.GetStatus(context.Background(), txHash)
	require.NoError(t, sErr)
	require.Equal(t, TxStatusConfirmed, st.Status)
}

func TestBlockchainRegistry_WithBitcoin(t *testing.T) {
	btc := NewBitcoinService()
	reg := NewBlockchainRegistry(btc)
	require.True(t, reg.Has(entities.BlockchainBitcoin))
	gw, err := reg.Get(entities.BlockchainBitcoin)
	require.NoError(t, err)
	w, wErr := gw.GenerateWallet(context.Background())
	require.NoError(t, wErr)
	require.Equal(t, entities.BlockchainBitcoin, w.Blockchain)
}

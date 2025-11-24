package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEthereumService_GenerateWallet(t *testing.T) {
	svc := NewEthereumService()
	w, err := svc.GenerateWallet(context.Background())
	require.NoError(t, err)
	require.Equal(t, entities.BlockchainEthereum, w.Blockchain)
	require.True(t, svc.ValidateAddress(w.Address))
	require.NotEmpty(t, w.PublicKey)
}

func TestEthereumService_EstimateFee(t *testing.T) {
	svc := NewEthereumService()
	quote, err := svc.EstimateFee(context.Background(), "0x0000000000000000000000000000000000000000", "0x0000000000000000000000000000000000000001", 123)
	require.NoError(t, err)
	require.Equal(t, int64(123), quote.AmountBaseUnit)
	require.Equal(t, "ETH", quote.FeeAsset)
	require.Greater(t, quote.EstimatedFee, int64(0))
}

func TestEthereumService_BroadcastAndStatus(t *testing.T) {
	svc := NewEthereumService()
	from := "0x0000000000000000000000000000000000000000"
	to := "0x0000000000000000000000000000000000000001"
	txHash, err := svc.Broadcast(context.Background(), from, to, 10, "privkey")
	require.NoError(t, err)
	require.NotEmpty(t, txHash)
	st, sErr := svc.GetStatus(context.Background(), txHash)
	require.NoError(t, sErr)
	require.Equal(t, TxStatusConfirmed, st.Status)
}

func TestBlockchainRegistry_WithEthereum(t *testing.T) {
	tron := &mockTron{}
	eth := NewEthereumService()
	reg := NewBlockchainRegistry(tron, eth)
	require.True(t, reg.Has(entities.BlockchainTRON))
	require.True(t, reg.Has(entities.BlockchainEthereum))
	gwEth, err := reg.Get(entities.BlockchainEthereum)
	require.NoError(t, err)
	wallet, wErr := gwEth.GenerateWallet(context.Background())
	require.NoError(t, wErr)
	require.Equal(t, entities.BlockchainEthereum, wallet.Blockchain)
}

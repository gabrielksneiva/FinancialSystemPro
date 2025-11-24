package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockTron minimal para teste isolado sem chamadas HTTP reais.
type mockTron struct{}

func (m *mockTron) ChainType() entities.BlockchainType { return entities.BlockchainTRON }
func (m *mockTron) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	// Endereço Tron real (34 chars)
	return &entities.GeneratedWallet{Address: "TBkTz36UFssFa8Fjn8U1MKWeHg4Qq1zAiw", PublicKey: "PUB", Blockchain: entities.BlockchainTRON, CreatedAt: time.Now().Unix()}, nil
}
func (m *mockTron) ValidateAddress(a string) bool { return len(a) == 34 && a[0] == 'T' }
func (m *mockTron) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt, EstimatedFee: 25000, FeeAsset: "TRX", Source: "mock"}, nil
}
func (m *mockTron) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return TxHash("FAKE_HASH_123"), nil
}
func (m *mockTron) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed, Confirmations: 1, Required: 1}, nil
}

var _ BlockchainGatewayPort = (*mockTron)(nil)

func TestBlockchainGatewayPort_TronMock(t *testing.T) {
	gw := &mockTron{}
	ctx := context.Background()

	wallet, err := gw.GenerateWallet(ctx)
	require.NoError(t, err)
	require.Equal(t, entities.BlockchainTRON, wallet.Blockchain)
	require.True(t, gw.ValidateAddress(wallet.Address))

	fee, err := gw.EstimateFee(ctx, wallet.Address, wallet.Address, 1000)
	require.NoError(t, err)
	require.Equal(t, int64(1000), fee.AmountBaseUnit)
	require.Equal(t, "TRX", fee.FeeAsset)

	txHash, err := gw.Broadcast(ctx, wallet.Address, wallet.Address, 1000, "PRIVATE_KEY")
	require.NoError(t, err)
	require.NotEmpty(t, txHash)

	status, err := gw.GetStatus(ctx, txHash)
	require.NoError(t, err)
	require.Equal(t, TxStatusConfirmed, status.Status)
}

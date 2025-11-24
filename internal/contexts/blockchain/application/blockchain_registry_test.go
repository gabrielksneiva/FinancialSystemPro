package application

import (
	"context"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockGateway simples para teste de registro multi-chain.
type mockGateway struct{ chain entities.BlockchainType }

func (m *mockGateway) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Address: "TBkTz36UFssFa8Fjn8U1MKWeHg4Qq1zAiw", PublicKey: "PUB", Blockchain: m.chain, CreatedAt: time.Now().Unix()}, nil
}
func (m *mockGateway) ValidateAddress(a string) bool { return len(a) == 34 && a[0] == 'T' }
func (m *mockGateway) EstimateFee(ctx context.Context, from, to string, amt int64) (*services.FeeQuote, error) {
	return &services.FeeQuote{AmountBaseUnit: amt, EstimatedFee: 25000, FeeAsset: "TRX", Source: "mock"}, nil
}
func (m *mockGateway) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (services.TxHash, error) {
	return services.TxHash("HASH"), nil
}
func (m *mockGateway) GetStatus(ctx context.Context, h services.TxHash) (*services.TxStatusInfo, error) {
	return &services.TxStatusInfo{Hash: h, Status: services.TxStatusConfirmed}, nil
}
func (m *mockGateway) ChainType() entities.BlockchainType { return m.chain }

var _ services.BlockchainGatewayPort = (*mockGateway)(nil)

func TestBlockchainRegistry_RegisterAndGet(t *testing.T) {
	tronGw := &mockGateway{chain: entities.BlockchainTRON}
	reg := NewBlockchainRegistry(tronGw)
	require.True(t, reg.Has(entities.BlockchainTRON))
	gw, err := reg.Get(entities.BlockchainTRON)
	require.NoError(t, err)
	wallet, wErr := gw.GenerateWallet(context.Background())
	require.NoError(t, wErr)
	require.Equal(t, entities.BlockchainTRON, wallet.Blockchain)
}

func TestBlockchainRegistry_NotFound(t *testing.T) {
	reg := NewBlockchainRegistry()
	_, err := reg.Get(entities.BlockchainTRON)
	require.Error(t, err)
}

func TestBlockchainRegistry_Override(t *testing.T) {
	reg := NewBlockchainRegistry(&mockGateway{chain: entities.BlockchainTRON})
	// Registrar novamente substitui
	reg.Register(&mockGateway{chain: entities.BlockchainTRON})
	gw, err := reg.Get(entities.BlockchainTRON)
	require.NoError(t, err)
	status, sErr := gw.GetStatus(context.Background(), services.TxHash("X"))
	require.NoError(t, sErr)
	require.Equal(t, services.TxStatusConfirmed, status.Status)
}

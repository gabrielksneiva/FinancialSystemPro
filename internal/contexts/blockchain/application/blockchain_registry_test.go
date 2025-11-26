package application

import (
	"context"
	"financial-system-pro/internal/contexts/blockchain/domain"
	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockGateway simples para teste de registro multi-chain.
type mockGateway struct{ chain entity.BlockchainType }

func (m *mockGateway) GenerateWallet(ctx context.Context) (*entity.GeneratedWallet, error) {
	return &entity.GeneratedWallet{Address: "TBkTz36UFssFa8Fjn8U1MKWeHg4Qq1zAiw", PublicKey: "PUB", Blockchain: m.chain, CreatedAt: time.Now().Unix()}, nil
}
func (m *mockGateway) ValidateAddress(a string) bool { return len(a) == 34 && a[0] == 'T' }
func (m *mockGateway) EstimateFee(ctx context.Context, from, to string, amt int64) (*domain.FeeQuote, error) {
	return &domain.FeeQuote{AmountBaseUnit: amt, EstimatedFee: 25000, FeeAsset: "TRX", Source: "mock"}, nil
}
func (m *mockGateway) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (domain.TxHash, error) {
	return domain.TxHash("HASH"), nil
}
func (m *mockGateway) GetStatus(ctx context.Context, h domain.TxHash) (*domain.TxStatusInfo, error) {
	return &domain.TxStatusInfo{Hash: h, Status: domain.TxStatusConfirmed}, nil
}
func (m *mockGateway) ChainType() entity.BlockchainType                              { return m.chain }
func (m *mockGateway) GetBalance(ctx context.Context, address string) (int64, error) { return 0, nil }
func (m *mockGateway) GetTransactionHistory(ctx context.Context, address string, limit, offset int) ([]*entity.BlockchainTransaction, error) {
	return []*entity.BlockchainTransaction{}, nil
}
func (m *mockGateway) SubscribeNewBlocks(ctx context.Context, handler domain.BlockEventHandler) error {
	return nil
}
func (m *mockGateway) SubscribeNewTransactions(ctx context.Context, address string, handler domain.TxEventHandler) error {
	return nil
}

var _ domain.BlockchainGatewayPort = (*mockGateway)(nil)

func TestBlockchainRegistry_RegisterAndGet(t *testing.T) {
	tronGw := &mockGateway{chain: entity.BlockchainTron}
	reg := NewBlockchainRegistry(tronGw)
	require.True(t, reg.Has(entity.BlockchainTron))
	gw, err := reg.Get(entity.BlockchainTron)
	require.NoError(t, err)
	wallet, wErr := gw.GenerateWallet(context.Background())
	require.NoError(t, wErr)
	require.Equal(t, entity.BlockchainTron, wallet.Blockchain)
}

func TestBlockchainRegistry_NotFound(t *testing.T) {
	reg := NewBlockchainRegistry()
	_, err := reg.Get(entity.BlockchainTron)
	require.Error(t, err)
}

func TestBlockchainRegistry_Override(t *testing.T) {
	reg := NewBlockchainRegistry(&mockGateway{chain: entity.BlockchainTron})
	// Registrar novamente substitui
	reg.Register(&mockGateway{chain: entity.BlockchainTron})
	gw, err := reg.Get(entity.BlockchainTron)
	require.NoError(t, err)
	status, sErr := gw.GetStatus(context.Background(), domain.TxHash("X"))
	require.NoError(t, sErr)
	require.Equal(t, domain.TxStatusConfirmed, status.Status)
}

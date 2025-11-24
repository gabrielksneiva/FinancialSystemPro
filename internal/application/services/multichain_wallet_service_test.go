package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockEthGateway reutiliza EthereumService já implementado
// mockTron já definido em outro teste

func setupInMemoryDB(t *testing.T) *repositories.NewDatabase {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	// Migrar modelos necessários
	err = db.AutoMigrate(&repositories.OnChainWallet{})
	require.NoError(t, err)
	return &repositories.NewDatabase{DB: db}
}

func TestMultiChainWalletService_GeneratePersist(t *testing.T) {
	tron := &mockTron2{}
	eth := NewEthereumService()
	reg := NewBlockchainRegistry(tron, eth)
	memDB := setupInMemoryDB(t)
	repo := NewOnChainWalletRepositoryAdapter(memDB)
	svc := NewMultiChainWalletService(reg, repo)
	ctx := context.Background()
	userID := uuid.New()

	wTron, err := svc.GenerateAndPersist(ctx, userID, entities.BlockchainTRON)
	require.NoError(t, err)
	require.Equal(t, entities.BlockchainTRON, wTron.Blockchain)

	wEth, err2 := svc.GenerateAndPersist(ctx, userID, entities.BlockchainEthereum)
	require.NoError(t, err2)
	require.Equal(t, entities.BlockchainEthereum, wEth.Blockchain)

	// Regerar mesma chain deve falhar
	_, errDup := svc.GenerateAndPersist(ctx, userID, entities.BlockchainTRON)
	require.Error(t, errDup)

	// Listagem
	list, listErr := repo.ListByUser(ctx, userID)
	require.NoError(t, listErr)
	require.Len(t, list, 2)
}

// mockTron replicate (min version) if not visible in this package build
type mockTron2 struct{}

func (m *mockTron2) ChainType() entities.BlockchainType { return entities.BlockchainTRON }
func (m *mockTron2) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Address: "TBkTz36UFssFa8Fjn8U1MKWeHg4Qq1zAiw", PublicKey: "PUB", Blockchain: entities.BlockchainTRON, CreatedAt: time.Now().Unix()}, nil
}
func (m *mockTron2) ValidateAddress(a string) bool { return true }
func (m *mockTron2) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt, EstimatedFee: 1, FeeAsset: "TRX", Source: "mock"}, nil
}
func (m *mockTron2) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return TxHash("HASH"), nil
}
func (m *mockTron2) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed}, nil
}

var _ BlockchainGatewayPort = (*mockTron2)(nil)

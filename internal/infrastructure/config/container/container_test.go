package container

import (
	"financial-system-pro/internal/application/services"
	repos "financial-system-pro/internal/infrastructure/database"
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestLoadConfigEmpty(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	cfg := LoadConfig()
	if cfg.DatabaseURL != "" {
		t.Fatalf("esperava vazio")
	}
}

func TestProvideQueueManager_NoRedis(t *testing.T) {
	lg, _ := zap.NewDevelopment()
	qm := ProvideQueueManager(Config{RedisURL: ""}, lg, nil)
	if qm != nil {
		t.Fatalf("esperava nil sem REDIS_URL")
	}
}

func TestProvideBlockchainRegistry(t *testing.T) {
	tron := services.NewTronService("ADDR", "PK")
	eth := services.NewEthereumService()
	btc := services.NewBitcoinService()
	reg := ProvideBlockchainRegistry(tron, eth, btc)
	if reg == nil {
		t.Fatalf("registro nil")
	}
	if !reg.Has("tron") || !reg.Has("ethereum") || !reg.Has("bitcoin") {
		t.Fatalf("não registrou chains esperadas")
	}
}

func TestProvideUserService_NilDB(t *testing.T) {
	usr := ProvideUserService(nil, zap.NewNop(), services.NewTronWalletManager())
	if usr != nil {
		t.Fatalf("esperava nil sem DB")
	}
}

func TestProvideDatabaseConnection_Empty(t *testing.T) {
	db, err := ProvideDatabaseConnection(Config{DatabaseURL: ""})
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if db != nil {
		t.Fatalf("esperava nil")
	}
}

func TestProvideSharedDatabaseConnection_Empty(t *testing.T) {
	conn, err := ProvideSharedDatabaseConnection(Config{DatabaseURL: ""})
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if conn != nil {
		t.Fatalf("esperava nil")
	}
}

func TestProvideOnChainWalletRepository_NilDB(t *testing.T) {
	repo := ProvideOnChainWalletRepository(nil)
	if repo != nil {
		t.Fatalf("esperava nil")
	}
}

func TestLinkUserMultiChain_NoMulti(t *testing.T) {
	lg := zap.NewNop()
	dummyDB := &repos.NewDatabase{}
	usvc := services.NewUserService(dummyDB, lg, services.NewTronWalletManager())
	LinkUserMultiChain(usvc, nil)
	// Sem panic = sucesso.
}

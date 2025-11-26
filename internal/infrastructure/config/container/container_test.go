package container

import (
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

func TestProvideEventBus_NotNil(t *testing.T) {
	lg, _ := zap.NewDevelopment()
	bus := ProvideEventBus(lg)
	if bus == nil {
		t.Fatalf("esperava event bus não nil")
	}
}

func TestProvideDDDBlockchainRegistry(t *testing.T) {
	tron := ProvideTronGateway()
	eth := ProvideETHGateway()
	btc := ProvideBTCGateway()
	sol := ProvideSOLGateway()
	reg := ProvideDDDBlockchainRegistry(tron, eth, btc, sol)
	if reg == nil {
		t.Fatalf("registro nil")
	}
}

// TestProvideUserService_NilDB removed - legacy service deprecated
// TestUserService_NoPanic removed - legacy service deprecated

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

func TestProvideWalletRepository_NilConn(t *testing.T) {
	repo := ProvideWalletRepository(nil)
	if repo != nil {
		t.Fatalf("esperava nil")
	}
}

// TestUserService_NoPanic removed - testing deprecated service

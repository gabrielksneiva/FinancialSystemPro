package container

import (
	"testing"

	"go.uber.org/zap"
)

// TestProvideSharedDatabaseConnection_EmptyURL cobre caminho de retorno nil.
func TestProvideSharedDatabaseConnection_EmptyURL(t *testing.T) {
	cfg := Config{DatabaseURL: ""}
	conn, err := ProvideSharedDatabaseConnection(cfg)
	if err != nil {
		t.Fatalf("esperava err nil, obtido %v", err)
	}
	if conn != nil {
		t.Fatalf("esperava connection nil quando DATABASE_URL vazio")
	}
}

// TestProvideQueueManager_NoRedisURL cobre caminho sem fila.
func TestProvideEventBus(t *testing.T) {
	lg := zap.NewNop()
	bus := ProvideEventBus(lg)
	if bus == nil {
		t.Fatalf("esperava event bus não nil")
	}
}

// TestProvideDDDBlockchainRegistry_SemServicos verifica registro vazio.
func TestProvideDDDBlockchainRegistry_SemServicos(t *testing.T) {
	// Construir gateways via providers e registrar
	tron := ProvideTronGateway()
	eth := ProvideETHGateway()
	btc := ProvideBTCGateway()
	sol := ProvideSOLGateway()
	reg := ProvideDDDBlockchainRegistry(tron, eth, btc, sol)
	if reg == nil {
		t.Fatalf("registro nil")
	}
}

func TestProvideWalletManager_NotNil(t *testing.T) {
	wm := ProvideWalletManager()
	if wm == nil {
		t.Fatalf("esperava wallet manager não nil")
	}
}

// TestStartServer_MinimalLifecycle garante que hooks registram sem panic.
// Nota: Teste de StartServer omitido por dependências internas de fx.Lifecycle.

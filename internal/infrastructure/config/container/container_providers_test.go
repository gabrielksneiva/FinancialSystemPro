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
func TestProvideQueueManager_NoRedisURL(t *testing.T) {
	lg := zap.NewNop()
	qm := ProvideQueueManager(Config{RedisURL: ""}, lg, nil)
	if qm != nil {
		t.Fatalf("esperava qm nil sem REDIS_URL")
	}
}

// TestProvideBlockchainRegistry_SemServicos verifica registro vazio.
func TestProvideBlockchainRegistry_SemServicos(t *testing.T) {
	reg := ProvideBlockchainRegistry(nil, nil, nil)
	if reg == nil {
		t.Fatalf("registro nil")
	}
	// nenhuma chain registrada: Has deve retornar false para tipos conhecidos
	if reg.Has("tron") || reg.Has("ethereum") || reg.Has("bitcoin") {
		t.Fatalf("esperava Has false sem registro de chains")
	}
}

// TestProvideMultiChainWalletService_NilArgs retorna nil.
func TestProvideMultiChainWalletService_NilArgs(t *testing.T) {
	svc := ProvideMultiChainWalletService(nil, nil)
	if svc != nil {
		t.Fatalf("esperava nil com argumentos nil")
	}
}

// TestProvideQueueManager_WithInvalidRedisURL (não conecta, mas instancia) - usamos URL malformada.
func TestProvideQueueManager_WithInvalidRedisURL(t *testing.T) {
	lg := zap.NewNop()
	// valor fictício muito curto; QueueManager internamente só valida vazio.
	qm := ProvideQueueManager(Config{RedisURL: "redis://localhost:6379"}, lg, nil)
	// qm pode ser não nil mesmo sem interação real com redis
	if qm == nil {
		t.Fatalf("esperava qm não nil com REDIS_URL definido")
	}
	// fechar para evitar vazamentos (ignora erro)
	qm.Close()
}

// TestStartServer_MinimalLifecycle garante que hooks registram sem panic.
// Nota: Teste de StartServer omitido por dependências internas de fx.Lifecycle.

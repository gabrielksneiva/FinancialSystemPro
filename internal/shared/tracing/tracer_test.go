package tracing

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestStartSpanBasic(t *testing.T) {
	// Não inicializa provider (usa no-op), apenas garante fluxo.
	ctx := context.Background()
	ctx2, span := StartSpan(ctx, "test-instrumentation", "test-span")
	if ctx2 == nil || span == nil {
		t.Fatalf("span ou contexto nil")
	}
	span.End()
}

func TestInitTracer(t *testing.T) {
	// Tenta inicializar; se falhar por ambiente de rede, marcar como skip.
	logger := zap.NewNop()
	shutdown, err := InitTracer("test-service", logger)
	if err != nil {
		t.Skipf("falha ao inicializar tracer (aceitável em ambiente CI): %v", err)
	}
	if shutdown == nil {
		t.Fatalf("shutdown nil")
	}
}

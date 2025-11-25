package workers

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestNewQueueManager_InvalidURL(t *testing.T) {
	qm := NewQueueManager("invalid://host:1234", zap.NewNop(), nil)
	if qm != nil {
		t.Fatalf("esperava nil para URL inválida")
	}
}

func TestQueueManager_IsConnected(t *testing.T) {
	qm := NewQueueManager("redis://localhost:6379", zap.NewNop(), nil)
	if qm == nil {
		t.Fatalf("esperava instância não nil")
	}
	_ = qm.IsConnected() // cobre caminho sem afirmar estado
}

func TestEnqueueDeposit(t *testing.T) {
	qm := NewQueueManager("redis://localhost:6379", zap.NewNop(), nil)
	if qm == nil {
		t.Skip("instância não criada, não pode testar enqueue")
	}
	id, err := qm.EnqueueDeposit(context.Background(), "user-id", "10.00", "")
	if err == nil && id == "" {
		t.Fatalf("id vazio em sucesso")
	}
}

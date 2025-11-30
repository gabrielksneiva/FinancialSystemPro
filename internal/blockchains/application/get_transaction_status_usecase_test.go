package application

import (
	"context"
	"testing"

	"financial-system-pro/internal/blockchains/domain"
)

type mockStatusRepo struct {
	called bool
	status domain.TxStatus
}

func (m *mockStatusRepo) GetTransactionStatus(ctx context.Context, hash domain.TxHash) (domain.TxStatus, error) {
	m.called = true
	return m.status, nil
}

func TestGetTransactionStatusUseCase_Executes(t *testing.T) {
	repo := &mockStatusRepo{status: domain.TxConfirmed}
	uc := NewGetTransactionStatusUseCase(repo)

	ctx := context.Background()
	st, err := uc.Execute(ctx, domain.TxHash("tx_1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st != domain.TxConfirmed {
		t.Fatalf("expected confirmed status")
	}
	if !repo.called {
		t.Fatalf("expected repository GetTransactionStatus to be called")
	}
}

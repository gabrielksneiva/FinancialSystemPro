package application

import (
	"context"
	"testing"

	"financial-system-pro/internal/blockchains/domain"
)

type mockBalanceRepo struct {
	called bool
	amt    domain.Amount
}

func (m *mockBalanceRepo) GetBalance(ctx context.Context, addr domain.Address) (domain.Amount, error) {
	m.called = true
	return m.amt, nil
}

func TestGetBalanceUseCase_Executes(t *testing.T) {
	amt, _ := domain.NewAmountFromString("2500")
	repo := &mockBalanceRepo{amt: amt}
	uc := NewGetBalanceUseCase(repo)

	ctx := context.Background()
	got, err := uc.Execute(ctx, "addr1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Value.String() != "2500" {
		t.Fatalf("expected 2500 got %s", got.Value.String())
	}
	if !repo.called {
		t.Fatalf("expected repo GetBalance to be called")
	}
}

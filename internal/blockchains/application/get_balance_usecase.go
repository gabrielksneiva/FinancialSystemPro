package application

import (
	"context"
	"fmt"

	"financial-system-pro/internal/blockchains/domain"
)

// BalancePort is a port for retrieving balances
type BalancePort interface {
	GetBalance(ctx context.Context, addr domain.Address) (domain.Amount, error)
}

type GetBalanceUseCase struct {
	repo BalancePort
}

func NewGetBalanceUseCase(repo BalancePort) *GetBalanceUseCase {
	return &GetBalanceUseCase{repo: repo}
}

func (u *GetBalanceUseCase) Execute(ctx context.Context, addrStr string) (domain.Amount, error) {
	addr, err := domain.NewAddress(addrStr)
	if err != nil {
		return domain.Amount{}, fmt.Errorf("invalid address: %w", err)
	}
	return u.repo.GetBalance(ctx, addr)
}

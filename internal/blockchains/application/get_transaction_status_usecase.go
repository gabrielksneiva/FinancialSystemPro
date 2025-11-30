package application

import (
	"context"
	"fmt"

	"financial-system-pro/internal/blockchains/domain"
)

// TransactionStatusPort is a port for querying transaction status
type TransactionStatusPort interface {
	GetTransactionStatus(ctx context.Context, hash domain.TxHash) (domain.TxStatus, error)
}

type GetTransactionStatusUseCase struct {
	repo TransactionStatusPort
}

func NewGetTransactionStatusUseCase(repo TransactionStatusPort) *GetTransactionStatusUseCase {
	return &GetTransactionStatusUseCase{repo: repo}
}

func (u *GetTransactionStatusUseCase) Execute(ctx context.Context, hash domain.TxHash) (domain.TxStatus, error) {
	if hash == "" {
		return "", fmt.Errorf("empty tx hash")
	}
	return u.repo.GetTransactionStatus(ctx, hash)
}

package repository

import (
	"context"
	"financial-system-pro/internal/contexts/transaction/domain/entity"

	"github.com/google/uuid"
)

// TransactionRepository define as operações de persistência para Transaction
type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error)
	FindByHash(ctx context.Context, hash string) (*entity.Transaction, error)
	Update(ctx context.Context, tx *entity.Transaction) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TransactionStatus) error
}

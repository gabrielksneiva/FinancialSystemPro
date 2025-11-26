package transaction

import (
	"context"

	"github.com/google/uuid"
)

// TransactionRepository define persistência para agregados Transaction.
type TransactionRepository interface {
	Create(ctx context.Context, t *Transaction) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status, txHash string) error
	FindByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	FindByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*Transaction, error)
}

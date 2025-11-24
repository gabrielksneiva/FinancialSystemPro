package services

import (
	"context"
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
)

// TransactionRecordPort lida com persistência/atualização de registros de transação.
type TransactionRecordPort interface {
	Insert(ctx context.Context, tx *repositories.Transaction) error
	Update(ctx context.Context, txID uuid.UUID, updates map[string]interface{}) error
}

type transactionRecordAdapter struct{ db DatabasePort }

func NewTransactionRecordAdapter(db DatabasePort) TransactionRecordPort {
	return &transactionRecordAdapter{db: db}
}

func (a *transactionRecordAdapter) Insert(ctx context.Context, tx *repositories.Transaction) error {
	return a.db.Insert(tx)
}
func (a *transactionRecordAdapter) Update(ctx context.Context, txID uuid.UUID, updates map[string]interface{}) error {
	return a.db.UpdateTransaction(txID, updates)
}

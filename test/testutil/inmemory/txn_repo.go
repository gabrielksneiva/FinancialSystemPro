package inmemory

import (
	"context"
	"sync"

	txnEntity "financial-system-pro/internal/contexts/transaction/domain/entity"
	"financial-system-pro/internal/domain/errors"

	"github.com/google/uuid"
)

type TransactionRepository struct {
	mu           sync.RWMutex
	transactions map[uuid.UUID]*txnEntity.Transaction
	byUserID     map[uuid.UUID][]*txnEntity.Transaction
	byHash       map[string]*txnEntity.Transaction
}

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		transactions: make(map[uuid.UUID]*txnEntity.Transaction),
		byUserID:     make(map[uuid.UUID][]*txnEntity.Transaction),
		byHash:       make(map[string]*txnEntity.Transaction),
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *txnEntity.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transactions[tx.ID] = tx
	r.byUserID[tx.UserID] = append(r.byUserID[tx.UserID], tx)
	if tx.TransactionHash != "" {
		r.byHash[tx.TransactionHash] = tx
	}
	return nil
}

func (r *TransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*txnEntity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, exists := r.transactions[id]
	if !exists {
		return nil, errors.NewNotFoundError("transaction")
	}
	return tx, nil
}

func (r *TransactionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*txnEntity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	txs := r.byUserID[userID]
	if txs == nil {
		return []*txnEntity.Transaction{}, nil
	}
	return txs, nil
}

func (r *TransactionRepository) FindByHash(ctx context.Context, hash string) (*txnEntity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, exists := r.byHash[hash]
	if !exists {
		return nil, errors.NewNotFoundError("transaction")
	}
	return tx, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx *txnEntity.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.transactions[tx.ID]; !exists {
		return errors.NewNotFoundError("transaction")
	}
	r.transactions[tx.ID] = tx
	return nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status txnEntity.TransactionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, exists := r.transactions[id]
	if !exists {
		return errors.NewNotFoundError("transaction")
	}
	tx.Status = status
	return nil
}

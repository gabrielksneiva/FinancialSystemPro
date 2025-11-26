package transaction

import (
	"errors"
	"financial-system-pro/internal/contexts/transaction/domain/valueobject"
	"time"

	"github.com/google/uuid"
)

// Status represents transaction lifecycle state.
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Transaction aggregate root capturing invariants.
type Transaction struct {
	id        uuid.UUID
	fromUser  uuid.UUID
	toUser    uuid.UUID
	amount    valueobject.Money
	status    Status
	createdAt time.Time
	updatedAt time.Time
	// domain events captured during lifecycle
	events []interface{}
}

// NewTransaction creates a pending transaction.
func NewTransaction(from, to uuid.UUID, amount valueobject.Money) (*Transaction, error) {
	if from == to {
		return nil, errors.New("from and to user cannot be equal")
	}
	if amount.Amount().IsZero() {
		return nil, errors.New("amount must be > 0")
	}
	t := &Transaction{
		id:        uuid.New(),
		fromUser:  from,
		toUser:    to,
		amount:    amount,
		status:    StatusPending,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}
	return t, nil
}

func (t *Transaction) ID() uuid.UUID             { return t.id }
func (t *Transaction) Status() Status            { return t.status }
func (t *Transaction) Amount() valueobject.Money { return t.amount }
func (t *Transaction) FromUser() uuid.UUID       { return t.fromUser }
func (t *Transaction) ToUser() uuid.UUID         { return t.toUser }
func (t *Transaction) Events() []interface{}     { return append([]interface{}{}, t.events...) }

// Complete marks a transaction successful and records domain event.
func (t *Transaction) Complete(txHash string) error {
	if t.status != StatusPending {
		return errors.New("transaction not pending")
	}
	t.status = StatusCompleted
	t.updatedAt = time.Now()
	t.events = append(t.events, map[string]interface{}{ // simplified domain event struct for now
		"type":   "transaction.completed",
		"tx_id":  t.id.String(),
		"hash":   txHash,
		"amount": t.amount.Amount().String(),
	})
	return nil
}

// Fail marks transaction failed.
func (t *Transaction) Fail(reason string) error {
	if t.status != StatusPending {
		return errors.New("transaction not pending")
	}
	t.status = StatusFailed
	t.updatedAt = time.Now()
	t.events = append(t.events, map[string]interface{}{
		"type":   "transaction.failed",
		"tx_id":  t.id.String(),
		"reason": reason,
	})
	return nil
}

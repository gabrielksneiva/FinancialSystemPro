package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionType define os tipos de transação
type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"
)

// TransactionStatus define os status de transação
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

// Transaction representa uma transação financeira
type Transaction struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Type            TransactionType
	Amount          decimal.Decimal
	Status          TransactionStatus
	TransactionHash string
	FromAddress     string
	ToAddress       string
	CallbackURL     string
	ErrorMessage    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time
}

// NewTransaction cria uma nova transação
func NewTransaction(userID uuid.UUID, txType TransactionType, amount decimal.Decimal) *Transaction {
	return &Transaction{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      txType,
		Amount:    amount,
		Status:    TransactionStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Complete marca a transação como concluída
func (t *Transaction) Complete(txHash string) {
	now := time.Now()
	t.Status = TransactionStatusCompleted
	t.TransactionHash = txHash
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// Fail marca a transação como falha
func (t *Transaction) Fail(errorMsg string) {
	t.Status = TransactionStatusFailed
	t.ErrorMessage = errorMsg
	t.UpdatedAt = time.Now()
}

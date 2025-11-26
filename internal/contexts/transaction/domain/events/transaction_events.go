package events

import (
	"financial-system-pro/internal/shared/events"
	"time"

	"github.com/google/uuid"
)

// TransactionCreated evento disparado quando transação é criada
type TransactionCreated struct {
	events.BaseDomainEvent
	UserID         uuid.UUID
	Amount         float64
	Currency       string
	Type           string
	BlockchainType string
	WalletAddress  string
}

func NewTransactionCreated(txID, userID uuid.UUID, amount float64, currency, txType, blockchain, walletAddr string) *TransactionCreated {
	return &TransactionCreated{
		BaseDomainEvent: events.NewBaseDomainEvent("TransactionCreated", txID),
		UserID:          userID,
		Amount:          amount,
		Currency:        currency,
		Type:            txType,
		BlockchainType:  blockchain,
		WalletAddress:   walletAddr,
	}
}

// TransactionConfirmed evento disparado quando transação é confirmada
type TransactionConfirmed struct {
	events.BaseDomainEvent
	Hash          string
	Confirmations int
	BlockNumber   int64
	ConfirmedAt   time.Time
}

func NewTransactionConfirmed(txID uuid.UUID, hash string, confirmations int, blockNumber int64) *TransactionConfirmed {
	return &TransactionConfirmed{
		BaseDomainEvent: events.NewBaseDomainEvent("TransactionConfirmed", txID),
		Hash:            hash,
		Confirmations:   confirmations,
		BlockNumber:     blockNumber,
		ConfirmedAt:     time.Now(),
	}
}

// TransactionCompleted evento disparado quando transação é completada
type TransactionCompleted struct {
	events.BaseDomainEvent
	Hash        string
	CompletedAt time.Time
	FinalAmount float64
}

func NewTransactionCompleted(txID uuid.UUID, hash string, finalAmount float64) *TransactionCompleted {
	return &TransactionCompleted{
		BaseDomainEvent: events.NewBaseDomainEvent("TransactionCompleted", txID),
		Hash:            hash,
		CompletedAt:     time.Now(),
		FinalAmount:     finalAmount,
	}
}

// TransactionFailed evento disparado quando transação falha
type TransactionFailed struct {
	events.BaseDomainEvent
	Reason    string
	FailedAt  time.Time
	ErrorCode string
}

func NewTransactionFailed(txID uuid.UUID, reason, errorCode string) *TransactionFailed {
	return &TransactionFailed{
		BaseDomainEvent: events.NewBaseDomainEvent("TransactionFailed", txID),
		Reason:          reason,
		FailedAt:        time.Now(),
		ErrorCode:       errorCode,
	}
}

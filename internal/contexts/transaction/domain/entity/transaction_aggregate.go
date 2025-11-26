package entity

import (
	"errors"
	"financial-system-pro/internal/contexts/transaction/domain/events"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionAggregate representa o agregado raiz de transação
type TransactionAggregate struct {
	transaction  *Transaction
	domainEvents []interface{} // eventos não publicados
}

// NewTransactionAggregate cria um novo agregado de transação
func NewTransactionAggregate(userID uuid.UUID, txType TransactionType, amount decimal.Decimal, currency, blockchain, walletAddr string) (*TransactionAggregate, error) {
	if amount.IsNegative() || amount.IsZero() {
		return nil, errors.New("amount must be positive")
	}

	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	tx := NewTransaction(userID, txType, amount)
	tx.ToAddress = walletAddr

	agg := &TransactionAggregate{
		transaction:  tx,
		domainEvents: make([]interface{}, 0),
	}

	// Dispara evento de criação
	event := events.NewTransactionCreated(
		tx.ID,
		userID,
		amount.InexactFloat64(),
		currency,
		string(txType),
		blockchain,
		walletAddr,
	)
	agg.domainEvents = append(agg.domainEvents, event)

	return agg, nil
}

// Transaction retorna a entidade Transaction
func (a *TransactionAggregate) Transaction() *Transaction {
	return a.transaction
}

// DomainEvents retorna os eventos de domínio não publicados
func (a *TransactionAggregate) DomainEvents() []interface{} {
	return a.domainEvents
}

// ClearDomainEvents limpa os eventos após publicação
func (a *TransactionAggregate) ClearDomainEvents() {
	a.domainEvents = make([]interface{}, 0)
}

// Confirm confirma a transação com informações da blockchain
func (a *TransactionAggregate) Confirm(txHash string, confirmations int, blockNumber int64) error {
	if a.transaction.Status == TransactionStatusCompleted {
		return errors.New("cannot confirm already completed transaction")
	}

	if a.transaction.Status == TransactionStatusFailed {
		return errors.New("cannot confirm failed transaction")
	}

	if txHash == "" {
		return errors.New("transaction hash cannot be empty")
	}

	if confirmations < 0 {
		return errors.New("confirmations cannot be negative")
	}

	a.transaction.TransactionHash = txHash
	a.transaction.UpdatedAt = time.Now()

	// Dispara evento de confirmação
	event := events.NewTransactionConfirmed(
		a.transaction.ID,
		txHash,
		confirmations,
		blockNumber,
	)
	a.domainEvents = append(a.domainEvents, event)

	return nil
}

// Complete marca a transação como concluída
func (a *TransactionAggregate) Complete(finalAmount decimal.Decimal) error {
	if a.transaction.Status == TransactionStatusCompleted {
		return errors.New("transaction already completed")
	}

	if a.transaction.Status == TransactionStatusFailed {
		return errors.New("cannot complete failed transaction")
	}

	if a.transaction.TransactionHash == "" {
		return errors.New("transaction must be confirmed before completion")
	}

	now := time.Now()
	a.transaction.Status = TransactionStatusCompleted
	a.transaction.CompletedAt = &now
	a.transaction.UpdatedAt = now

	// Atualiza amount se diferente (pode ter taxas)
	if !finalAmount.IsZero() {
		a.transaction.Amount = finalAmount
	}

	// Dispara evento de conclusão
	event := events.NewTransactionCompleted(
		a.transaction.ID,
		a.transaction.TransactionHash,
		a.transaction.Amount.InexactFloat64(),
	)
	a.domainEvents = append(a.domainEvents, event)

	return nil
}

// Fail marca a transação como falha
func (a *TransactionAggregate) Fail(reason, errorCode string) error {
	if a.transaction.Status == TransactionStatusCompleted {
		return errors.New("cannot fail completed transaction")
	}

	a.transaction.Status = TransactionStatusFailed
	a.transaction.ErrorMessage = reason
	a.transaction.UpdatedAt = time.Now()

	// Dispara evento de falha
	event := events.NewTransactionFailed(
		a.transaction.ID,
		reason,
		errorCode,
	)
	a.domainEvents = append(a.domainEvents, event)

	return nil
}

// IsPending verifica se a transação está pendente
func (a *TransactionAggregate) IsPending() bool {
	return a.transaction.Status == TransactionStatusPending
}

// IsCompleted verifica se a transação está completa
func (a *TransactionAggregate) IsCompleted() bool {
	return a.transaction.Status == TransactionStatusCompleted
}

// IsFailed verifica se a transação falhou
func (a *TransactionAggregate) IsFailed() bool {
	return a.transaction.Status == TransactionStatusFailed
}

// CanBeConfirmed verifica se a transação pode ser confirmada
func (a *TransactionAggregate) CanBeConfirmed() bool {
	return a.transaction.Status == TransactionStatusPending
}

// CanBeCompleted verifica se a transação pode ser completada
func (a *TransactionAggregate) CanBeCompleted() bool {
	return a.transaction.Status == TransactionStatusPending && a.transaction.TransactionHash != ""
}

// GetAmount retorna o valor da transação
func (a *TransactionAggregate) GetAmount() decimal.Decimal {
	return a.transaction.Amount
}

// GetUserID retorna o ID do usuário
func (a *TransactionAggregate) GetUserID() uuid.UUID {
	return a.transaction.UserID
}

// GetType retorna o tipo da transação
func (a *TransactionAggregate) GetType() TransactionType {
	return a.transaction.Type
}

// GetStatus retorna o status da transação
func (a *TransactionAggregate) GetStatus() TransactionStatus {
	return a.transaction.Status
}

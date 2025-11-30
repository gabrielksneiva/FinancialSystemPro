package application

import (
	"context"
	"fmt"

	"financial-system-pro/internal/blockchains/domain"
	"financial-system-pro/internal/blockchains/events"
)

// TransactionRepository is a port for creating transactions
type TransactionRepository interface {
	CreateTransaction(ctx context.Context, from domain.Address, to domain.Address, amount domain.Amount) (domain.Transaction, error)
}

// EventPublisher is a minimal event publisher port
type EventPublisher interface {
	Publish(evt events.EventType, id string, payload interface{})
}

type CreateTransactionUseCase struct {
	repo  TransactionRepository
	pub   EventPublisher
	chain string
}

func NewCreateTransactionUseCase(repo TransactionRepository, pub EventPublisher, chain string) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo, pub: pub, chain: chain}
}

// Execute validates inputs, creates the transaction in the repository and publishes a NewTransaction event.
func (u *CreateTransactionUseCase) Execute(ctx context.Context, fromStr, toStr, amountStr string) (domain.Transaction, error) {
	from, err := domain.NewAddress(fromStr)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("invalid from address: %w", err)
	}
	to, err := domain.NewAddress(toStr)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("invalid to address: %w", err)
	}
	amt, err := domain.NewAmountFromString(amountStr)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("invalid amount: %w", err)
	}

	tx, err := u.repo.CreateTransaction(ctx, from, to, amt)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("create tx: %w", err)
	}

	// publish event asynchronously via publisher (publisher implementation may deliver async)
	eventID := u.chain + ":" + tx.Hash
	u.pub.Publish(events.EventNewTransaction, eventID, events.NewTransactionEvent{Chain: u.chain, TxHash: tx.Hash})

	return tx, nil
}

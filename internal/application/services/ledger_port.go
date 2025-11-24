package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// LedgerPort abstrai operações de saldo/lançamentos de conta.
type LedgerPort interface {
	Apply(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, txType string) error
	Balance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
}

// ledgerAdapter adapta DatabasePort atual (Transaction/Balance) para LedgerPort.
type ledgerAdapter struct{ db DatabasePort }

func NewLedgerAdapter(db DatabasePort) LedgerPort { return &ledgerAdapter{db: db} }

func (l *ledgerAdapter) Apply(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, txType string) error {
	return l.db.Transaction(userID, amount, txType)
}

func (l *ledgerAdapter) Balance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	return l.db.Balance(userID)
}

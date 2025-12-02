package finance

import (
	"errors"
	"time"
)

type EntryType string

const (
	EntryTypeCredit EntryType = "credit"
	EntryTypeDebit  EntryType = "debit"
)

// LedgerEntry represents an immutable append-only record for wallet balance changes.
// Invariants: id, walletID non-empty; type is credit/debit; balanceAfter >=0.

type LedgerEntry struct {
	id           string
	walletID     string
	type_        EntryType
	amount       Amount
	balanceAfter int64
	description  string
	createdAt    time.Time
}

func NewLedgerEntry(id, walletID string, t EntryType, amount Amount, balanceAfter int64, description string) (*LedgerEntry, error) {
	if id == "" {
		return nil, errors.New("ledger entry id is required")
	}
	if walletID == "" {
		return nil, errors.New("wallet id is required")
	}
	if t != EntryTypeCredit && t != EntryTypeDebit {
		return nil, errors.New("invalid entry type")
	}
	if balanceAfter < 0 {
		return nil, errors.New("balance after cannot be negative")
	}
	return &LedgerEntry{id: id, walletID: walletID, type_: t, amount: amount, balanceAfter: balanceAfter, description: description, createdAt: time.Now().UTC()}, nil
}

func (l *LedgerEntry) ID() string           { return l.id }
func (l *LedgerEntry) WalletID() string     { return l.walletID }
func (l *LedgerEntry) Type() EntryType      { return l.type_ }
func (l *LedgerEntry) Amount() Amount       { return l.amount }
func (l *LedgerEntry) BalanceAfter() int64  { return l.balanceAfter }
func (l *LedgerEntry) Description() string  { return l.description }
func (l *LedgerEntry) CreatedAt() time.Time { return l.createdAt }

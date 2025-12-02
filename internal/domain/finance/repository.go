package finance

import "context"

// AccountRepository defines persistence operations for Accounts.
type AccountRepository interface {
	Get(ctx context.Context, id string) (*Account, error)
	Save(ctx context.Context, a *Account) error
}

// WalletRepository defines persistence operations for Wallets.
type WalletRepository interface {
	Get(ctx context.Context, id string) (*Wallet, error)
	Save(ctx context.Context, w *Wallet) error
	ListByAccount(ctx context.Context, accountID string) ([]*Wallet, error)
}

// LedgerRepository defines append-only persistence for ledger entries.
type LedgerRepository interface {
	Append(ctx context.Context, e *LedgerEntry) error
	ListByWallet(ctx context.Context, walletID string) ([]*LedgerEntry, error)
	ListSince(ctx context.Context, walletID string, sinceID string) ([]*LedgerEntry, error)
}

// Transactional allows wrapping operations in a transaction boundary if supported.
type Transactional interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

package finance

import "context"

// LedgerService orchestrates balance mutations ensuring append-only ledger persistence.
// NOTE: In a real implementation concurrency control (optimistic locking) should be applied.

type LedgerService struct {
	wallets WalletRepository
	ledger  LedgerRepository
}

func NewLedgerService(w WalletRepository, l LedgerRepository) *LedgerService {
	return &LedgerService{wallets: w, ledger: l}
}

func (s *LedgerService) Credit(ctx context.Context, walletID string, amount Amount, desc string) error {
	w, err := s.wallets.Get(ctx, walletID)
	if err != nil {
		return err
	}
	if err := w.Credit(amount); err != nil {
		return err
	}
	entry, err := NewLedgerEntry("le-credit-"+walletID, walletID, EntryTypeCredit, amount, w.Balance().Value(), desc)
	if err != nil {
		return err
	}
	if err := s.wallets.Save(ctx, w); err != nil {
		return err
	}
	return s.ledger.Append(ctx, entry)
}

func (s *LedgerService) Debit(ctx context.Context, walletID string, amount Amount, desc string) error {
	w, err := s.wallets.Get(ctx, walletID)
	if err != nil {
		return err
	}
	if err := w.Debit(amount); err != nil {
		return err
	}
	entry, err := NewLedgerEntry("le-debit-"+walletID, walletID, EntryTypeDebit, amount, w.Balance().Value(), desc)
	if err != nil {
		return err
	}
	if err := s.wallets.Save(ctx, w); err != nil {
		return err
	}
	return s.ledger.Append(ctx, entry)
}

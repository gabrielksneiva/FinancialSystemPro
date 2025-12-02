package finance

import (
	"context"
	"errors"
	"testing"
)

type inMemoryAccountRepo struct{ m map[string]*Account }

func newInMemoryAccountRepo() *inMemoryAccountRepo {
	return &inMemoryAccountRepo{m: map[string]*Account{}}
}
func (r *inMemoryAccountRepo) Get(ctx context.Context, id string) (*Account, error) {
	a, ok := r.m[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return a, nil
}
func (r *inMemoryAccountRepo) Save(ctx context.Context, a *Account) error {
	r.m[a.ID()] = a
	return nil
}

type inMemoryWalletRepo struct{ m map[string]*Wallet }

func newInMemoryWalletRepo() *inMemoryWalletRepo { return &inMemoryWalletRepo{m: map[string]*Wallet{}} }
func (r *inMemoryWalletRepo) Get(ctx context.Context, id string) (*Wallet, error) {
	w, ok := r.m[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return w, nil
}
func (r *inMemoryWalletRepo) Save(ctx context.Context, w *Wallet) error { r.m[w.ID()] = w; return nil }
func (r *inMemoryWalletRepo) ListByAccount(ctx context.Context, accountID string) ([]*Wallet, error) {
	out := []*Wallet{}
	for _, w := range r.m {
		if w.AccountID() == accountID {
			out = append(out, w)
		}
	}
	return out, nil
}

type inMemoryLedgerRepo struct{ m map[string][]*LedgerEntry }

func newInMemoryLedgerRepo() *inMemoryLedgerRepo {
	return &inMemoryLedgerRepo{m: map[string][]*LedgerEntry{}}
}
func (r *inMemoryLedgerRepo) Append(ctx context.Context, e *LedgerEntry) error {
	r.m[e.WalletID()] = append(r.m[e.WalletID()], e)
	return nil
}
func (r *inMemoryLedgerRepo) ListByWallet(ctx context.Context, walletID string) ([]*LedgerEntry, error) {
	return r.m[walletID], nil
}
func (r *inMemoryLedgerRepo) ListSince(ctx context.Context, walletID string, sinceID string) ([]*LedgerEntry, error) {
	out := []*LedgerEntry{}
	for _, e := range r.m[walletID] {
		if e.ID() > sinceID {
			out = append(out, e)
		}
	}
	return out, nil
}

func TestRepositoryContracts(t *testing.T) {
	ctx := context.Background()
	accRepo := newInMemoryAccountRepo()
	walRepo := newInMemoryWalletRepo()
	ledRepo := newInMemoryLedgerRepo()

	acc, _ := NewAccount("acc-rc-1", "holder-z")
	if err := accRepo.Save(ctx, acc); err != nil {
		t.Fatalf("save account: %v", err)
	}
	fetched, err := accRepo.Get(ctx, acc.ID())
	if err != nil || fetched.ID() != acc.ID() {
		t.Fatalf("get account mismatch")
	}

	zero, _ := NewAmount(0, "USD")
	w, _ := NewWallet("wal-rc-1", acc.ID(), zero)
	if err := walRepo.Save(ctx, w); err != nil {
		t.Fatalf("save wallet: %v", err)
	}
	w2, err := walRepo.Get(ctx, w.ID())
	if err != nil || w2.ID() != w.ID() {
		t.Fatalf("get wallet mismatch")
	}
	list, _ := walRepo.ListByAccount(ctx, acc.ID())
	if len(list) != 1 {
		t.Fatalf("expected 1 wallet for account")
	}

	amt, _ := NewAmount(10, "USD")
	le, _ := NewLedgerEntry("le-rc-1", w.ID(), EntryTypeCredit, amt, 10, "init deposit")
	if err := ledRepo.Append(ctx, le); err != nil {
		t.Fatalf("append ledger: %v", err)
	}
	entries, _ := ledRepo.ListByWallet(ctx, w.ID())
	if len(entries) != 1 {
		t.Fatalf("expected 1 ledger entry")
	}
	entriesSince, _ := ledRepo.ListSince(ctx, w.ID(), "le-rc-0")
	if len(entriesSince) != 1 {
		t.Fatalf("expected entries since id")
	}
}

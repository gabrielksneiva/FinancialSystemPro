package usecases

import (
	"context"
	"testing"

	"financial-system-pro/internal/domain/finance"
	"financial-system-pro/internal/infrastructure/eventbus"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupUseCaseTest() (finance.AccountRepository, finance.WalletRepository, finance.LedgerRepository, finance.Transactional, eventbus.Bus) {
	// Use in-memory implementations
	accRepo := &inMemoryAccountRepo{m: make(map[string]*finance.Account)}
	walRepo := &inMemoryWalletRepo{m: make(map[string]*finance.Wallet)}
	ledRepo := &inMemoryLedgerRepo{m: make(map[string][]*finance.LedgerEntry)}
	txProvider := &noOpTransactional{}
	
	store := eventbus.NewInMemoryStore()
	procLog := eventbus.NewInMemoryProcessingLog()
	bus := eventbus.NewResilientBus(store, procLog, zap.NewNop())

	return accRepo, walRepo, ledRepo, txProvider, bus
}

// Mock implementations
type inMemoryAccountRepo struct{ m map[string]*finance.Account }
func (r *inMemoryAccountRepo) Get(ctx context.Context, id string) (*finance.Account, error) { return r.m[id], nil }
func (r *inMemoryAccountRepo) Save(ctx context.Context, a *finance.Account) error { r.m[a.ID()] = a; return nil }

type inMemoryWalletRepo struct{ m map[string]*finance.Wallet }
func (r *inMemoryWalletRepo) Get(ctx context.Context, id string) (*finance.Wallet, error) { return r.m[id], nil }
func (r *inMemoryWalletRepo) Save(ctx context.Context, w *finance.Wallet) error { r.m[w.ID()] = w; return nil }
func (r *inMemoryWalletRepo) ListByAccount(ctx context.Context, accountID string) ([]*finance.Wallet, error) { return nil, nil }

type inMemoryLedgerRepo struct{ m map[string][]*finance.LedgerEntry }
func (r *inMemoryLedgerRepo) Append(ctx context.Context, e *finance.LedgerEntry) error { r.m[e.WalletID()] = append(r.m[e.WalletID()], e); return nil }
func (r *inMemoryLedgerRepo) ListByWallet(ctx context.Context, walletID string) ([]*finance.LedgerEntry, error) { return r.m[walletID], nil }
func (r *inMemoryLedgerRepo) ListSince(ctx context.Context, walletID string, sinceID string) ([]*finance.LedgerEntry, error) { return nil, nil }

type noOpTransactional struct{}
func (t *noOpTransactional) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

func TestDepositUseCase(t *testing.T) {
	accRepo, walRepo, ledRepo, txProvider, bus := setupUseCaseTest()
	defer bus.Close()

	uc := NewDepositUseCase(accRepo, walRepo, ledRepo, txProvider, bus)
	ctx := context.Background()

	// Setup account and wallet
	acc, _ := finance.NewAccount("acc-1", "holder-1")
	_ = accRepo.Save(ctx, acc)

	zero, _ := finance.NewAmount(0, "USD")
	wallet, _ := finance.NewWallet("wal-1", "acc-1", zero)
	_ = walRepo.Save(ctx, wallet)

	// Execute deposit
	req := DepositRequest{
		AccountID:   "acc-1",
		WalletID:    "wal-1",
		Amount:      10000, // $100.00
		Currency:    "USD",
		Description: "Test deposit",
		ExternalID:  "ext-123",
	}

	result, err := uc.Execute(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "acc-1", result.AccountID)
	require.Equal(t, int64(10000), result.BalanceAfter)

	// Verify wallet updated
	updatedWallet, _ := walRepo.Get(ctx, "wal-1")
	require.Equal(t, int64(10000), updatedWallet.Balance().Value())

	// Verify ledger entry
	entries, _ := ledRepo.ListByWallet(ctx, "wal-1")
	require.Len(t, entries, 1)
	require.Equal(t, finance.EntryTypeCredit, entries[0].Type())
}

func TestWithdrawalUseCase(t *testing.T) {
	accRepo, walRepo, ledRepo, txProvider, bus := setupUseCaseTest()
	defer bus.Close()

	uc := NewWithdrawalUseCase(accRepo, walRepo, ledRepo, txProvider, bus)
	ctx := context.Background()

	// Setup account and wallet with balance
	acc, _ := finance.NewAccount("acc-2", "holder-2")
	_ = accRepo.Save(ctx, acc)

	initial, _ := finance.NewAmount(20000, "USD")
	wallet, _ := finance.NewWallet("wal-2", "acc-2", initial)
	_ = walRepo.Save(ctx, wallet)

	// Execute withdrawal
	req := WithdrawalRequest{
		AccountID:        "acc-2",
		WalletID:         "wal-2",
		Amount:           5000,
		Currency:         "USD",
		Description:      "Test withdrawal",
		DestinationID:    "bank-123",
		WithdrawalMethod: "bank_transfer",
	}

	result, err := uc.Execute(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(15000), result.BalanceAfter)

	// Verify wallet updated
	updatedWallet, _ := walRepo.Get(ctx, "wal-2")
	require.Equal(t, int64(15000), updatedWallet.Balance().Value())

	// Test insufficient balance
	req.Amount = 30000
	_, err = uc.Execute(ctx, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient balance")
}

func TestTransferUseCase(t *testing.T) {
	accRepo, walRepo, ledRepo, txProvider, bus := setupUseCaseTest()
	defer bus.Close()

	uc := NewTransferUseCase(accRepo, walRepo, ledRepo, txProvider, bus)
	ctx := context.Background()

	// Setup accounts and wallets
	acc1, _ := finance.NewAccount("acc-3", "holder-3")
	_ = accRepo.Save(ctx, acc1)
	acc2, _ := finance.NewAccount("acc-4", "holder-4")
	_ = accRepo.Save(ctx, acc2)

	bal1, _ := finance.NewAmount(10000, "USD")
	wallet1, _ := finance.NewWallet("wal-3", "acc-3", bal1)
	_ = walRepo.Save(ctx, wallet1)

	bal2, _ := finance.NewAmount(5000, "USD")
	wallet2, _ := finance.NewWallet("wal-4", "acc-4", bal2)
	_ = walRepo.Save(ctx, wallet2)

	// Execute transfer
	req := TransferRequest{
		FromAccountID: "acc-3",
		FromWalletID:  "wal-3",
		ToAccountID:   "acc-4",
		ToWalletID:    "wal-4",
		Amount:        3000,
		Currency:      "USD",
		Description:   "Test transfer",
	}

	result, err := uc.Execute(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(7000), result.FromBalanceAfter)
	require.Equal(t, int64(8000), result.ToBalanceAfter)

	// Verify wallets updated
	w1, _ := walRepo.Get(ctx, "wal-3")
	require.Equal(t, int64(7000), w1.Balance().Value())
	w2, _ := walRepo.Get(ctx, "wal-4")
	require.Equal(t, int64(8000), w2.Balance().Value())

	// Verify ledger entries
	entries1, _ := ledRepo.ListByWallet(ctx, "wal-3")
	require.Len(t, entries1, 1)
	require.Equal(t, finance.EntryTypeDebit, entries1[0].Type())

	entries2, _ := ledRepo.ListByWallet(ctx, "wal-4")
	require.Len(t, entries2, 1)
	require.Equal(t, finance.EntryTypeCredit, entries2[0].Type())
}

func TestTransferUseCaseValidations(t *testing.T) {
	accRepo, walRepo, ledRepo, txProvider, bus := setupUseCaseTest()
	defer bus.Close()

	uc := NewTransferUseCase(accRepo, walRepo, ledRepo, txProvider, bus)
	ctx := context.Background()

	// Test same wallet transfer
	req := TransferRequest{
		FromAccountID: "acc-1",
		FromWalletID:  "wal-1",
		ToAccountID:   "acc-1",
		ToWalletID:    "wal-1",
		Amount:        100,
		Currency:      "USD",
	}
	_, err := uc.Execute(ctx, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot transfer to same wallet")

	// Test negative amount
	req.ToWalletID = "wal-2"
	req.Amount = -100
	_, err = uc.Execute(ctx, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be positive")
}

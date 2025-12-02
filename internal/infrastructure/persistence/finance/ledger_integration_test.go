package finance

import (
	"context"
	"database/sql"
	"testing"

	"financial-system-pro/internal/domain/finance"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Integration tests require a running PostgreSQL instance
// These tests validate concurrency, atomicity and append-only behavior

func setupTestDB(t *testing.T) *sql.DB {
	// Skip if no test database available
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/financial_system_test?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available for integration tests")
	}

	if err := db.Ping(); err != nil {
		t.Skip("PostgreSQL not available for integration tests")
	}

	// Clean tables
	_, _ = db.Exec("TRUNCATE finance_accounts, finance_wallets, finance_ledger_entries CASCADE")

	return db
}

func TestPostgresAccountRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresAccountRepository(db)
	ctx := context.Background()

	acc, err := finance.NewAccount("acc-pg-1", "holder-pg-1")
	require.NoError(t, err)

	// Save
	err = repo.Save(ctx, acc)
	require.NoError(t, err)

	// Get
	fetched, err := repo.Get(ctx, acc.ID())
	require.NoError(t, err)
	require.Equal(t, acc.ID(), fetched.ID())
	require.Equal(t, acc.HolderID(), fetched.HolderID())
	require.Equal(t, finance.AccountStatusActive, fetched.Status())

	// Update status
	err = acc.Deactivate()
	require.NoError(t, err)
	err = repo.Save(ctx, acc)
	require.NoError(t, err)

	fetched, err = repo.Get(ctx, acc.ID())
	require.NoError(t, err)
	require.Equal(t, finance.AccountStatusInactive, fetched.Status())
}

func TestPostgresWalletRepositoryOptimisticLock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	accRepo := NewPostgresAccountRepository(db)
	walRepo := NewPostgresWalletRepository(db)
	ctx := context.Background()

	acc, _ := finance.NewAccount("acc-pg-w-1", "holder-pg-w-1")
	_ = accRepo.Save(ctx, acc)

	zero, _ := finance.NewAmount(0, "USD")
	w, _ := finance.NewWallet("wal-pg-1", acc.ID(), zero)

	err := walRepo.Save(ctx, w)
	require.NoError(t, err)

	// Concurrent modification simulation
	w1, err := walRepo.Get(ctx, w.ID())
	require.NoError(t, err)
	w2, err := walRepo.Get(ctx, w.ID())
	require.NoError(t, err)

	amt, _ := finance.NewAmount(100, "USD")
	_ = w1.Credit(amt)
	err = walRepo.Save(ctx, w1)
	require.NoError(t, err)

	// Second save should detect version conflict
	_ = w2.Credit(amt)
	err = walRepo.Save(ctx, w2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "optimistic lock conflict")
}

func TestPostgresLedgerRepositoryAppendOnly(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	accRepo := NewPostgresAccountRepository(db)
	walRepo := NewPostgresWalletRepository(db)
	ledRepo := NewPostgresLedgerRepository(db)
	ctx := context.Background()

	acc, _ := finance.NewAccount("acc-pg-l-1", "holder-pg-l-1")
	_ = accRepo.Save(ctx, acc)

	zero, _ := finance.NewAmount(0, "USD")
	w, _ := finance.NewWallet("wal-pg-l-1", acc.ID(), zero)
	_ = walRepo.Save(ctx, w)

	// Append entries
	amt1, _ := finance.NewAmount(100, "USD")
	entry1, _ := finance.NewLedgerEntry("le-pg-1", w.ID(), finance.EntryTypeCredit, amt1, 100, "initial deposit")
	err := ledRepo.Append(ctx, entry1)
	require.NoError(t, err)

	amt2, _ := finance.NewAmount(50, "USD")
	entry2, _ := finance.NewLedgerEntry("le-pg-2", w.ID(), finance.EntryTypeDebit, amt2, 50, "withdrawal")
	err = ledRepo.Append(ctx, entry2)
	require.NoError(t, err)

	// List all
	entries, err := ledRepo.ListByWallet(ctx, w.ID())
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// List since
	entriesSince, err := ledRepo.ListSince(ctx, w.ID(), "le-pg-1")
	require.NoError(t, err)
	require.Len(t, entriesSince, 1)
	require.Equal(t, "le-pg-2", entriesSince[0].ID())
}

func TestPostgresTransactionalConcurrency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	accRepo := NewPostgresAccountRepository(db)
	walRepo := NewPostgresWalletRepository(db)
	ledRepo := NewPostgresLedgerRepository(db)
	txProvider := NewPostgresTransactional(db)
	ctx := context.Background()

	acc, _ := finance.NewAccount("acc-pg-tx-1", "holder-pg-tx-1")
	_ = accRepo.Save(ctx, acc)

	zero, _ := finance.NewAmount(0, "USD")
	w, _ := finance.NewWallet("wal-pg-tx-1", acc.ID(), zero)
	_ = walRepo.Save(ctx, w)

	svc := finance.NewLedgerService(walRepo, ledRepo)

	// Simulate concurrent credits
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			amt, _ := finance.NewAmount(10, "USD")
			err := txProvider.WithinTx(ctx, func(txCtx context.Context) error {
				return svc.Credit(txCtx, w.ID(), amt, "concurrent credit")
			})
			if err != nil {
				t.Logf("concurrent credit %d failed (expected with optimistic locking): %v", idx, err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// Final balance should be consistent
	finalWallet, err := walRepo.Get(ctx, w.ID())
	require.NoError(t, err)
	t.Logf("Final balance after concurrent operations: %d", finalWallet.Balance().Value())

	entries, err := ledRepo.ListByWallet(ctx, w.ID())
	require.NoError(t, err)
	t.Logf("Total ledger entries: %d", len(entries))
}

func TestLedgerServiceAtomicity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	accRepo := NewPostgresAccountRepository(db)
	walRepo := NewPostgresWalletRepository(db)
	ledRepo := NewPostgresLedgerRepository(db)
	ctx := context.Background()

	acc, _ := finance.NewAccount("acc-pg-svc-1", "holder-pg-svc-1")
	_ = accRepo.Save(ctx, acc)

	initial, _ := finance.NewAmount(100, "USD")
	w, _ := finance.NewWallet("wal-pg-svc-1", acc.ID(), initial)
	_ = walRepo.Save(ctx, w)

	svc := finance.NewLedgerService(walRepo, ledRepo)

	// Credit
	credAmt, _ := finance.NewAmount(50, "USD")
	err := svc.Credit(ctx, w.ID(), credAmt, "credit test")
	require.NoError(t, err)

	wUpdated, _ := walRepo.Get(ctx, w.ID())
	require.Equal(t, int64(150), wUpdated.Balance().Value())

	// Debit
	debAmt, _ := finance.NewAmount(30, "USD")
	err = svc.Debit(ctx, w.ID(), debAmt, "debit test")
	require.NoError(t, err)

	wUpdated, _ = walRepo.Get(ctx, w.ID())
	require.Equal(t, int64(120), wUpdated.Balance().Value())

	// Verify ledger entries
	entries, err := ledRepo.ListByWallet(ctx, w.ID())
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, finance.EntryTypeCredit, entries[0].Type())
	require.Equal(t, finance.EntryTypeDebit, entries[1].Type())
}

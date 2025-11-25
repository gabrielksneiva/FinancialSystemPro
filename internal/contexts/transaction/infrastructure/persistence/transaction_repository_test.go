package persistence

import (
	"context"
	"database/sql"
	"errors"
	txEntity "financial-system-pro/internal/contexts/transaction/domain/entity"
	dbIface "financial-system-pro/internal/shared/database"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// fakeConnectionTx reutiliza implementação similar de fakeConnection.
type fakeConnectionTx struct{ execs, queries int }

func (f *fakeConnectionTx) Query(ctx context.Context, q string, args ...interface{}) (dbIface.Rows, error) {
	f.queries++
	return &fakeTxRows{}, nil
}
func (f *fakeConnectionTx) QueryRow(ctx context.Context, q string, args ...interface{}) dbIface.Row {
	f.queries++
	return &fakeTxRow{}
}
func (f *fakeConnectionTx) Exec(ctx context.Context, q string, args ...interface{}) (dbIface.Result, error) {
	f.execs++
	return &fakeTxResult{}, nil
}
func (f *fakeConnectionTx) Begin(ctx context.Context) (dbIface.Transaction, error) {
	return nil, errors.New("not supported")
}
func (f *fakeConnectionTx) BeginTx(ctx context.Context, opts *sql.TxOptions) (dbIface.Transaction, error) {
	return nil, errors.New("not supported")
}
func (f *fakeConnectionTx) Ping(ctx context.Context) error { return nil }
func (f *fakeConnectionTx) Close() error                   { return nil }
func (f *fakeConnectionTx) Stats() sql.DBStats             { return sql.DBStats{} }

type fakeTxRows struct{}

func (r *fakeTxRows) Next() bool                     { return false }
func (r *fakeTxRows) Scan(dest ...interface{}) error { return nil }
func (r *fakeTxRows) Close() error                   { return nil }
func (r *fakeTxRows) Err() error                     { return nil }

type fakeTxRow struct{}

func (r *fakeTxRow) Scan(dest ...interface{}) error { return errors.New("no rows") }

type fakeTxResult struct{}

func (r *fakeTxResult) LastInsertId() (int64, error) { return 0, nil }
func (r *fakeTxResult) RowsAffected() (int64, error) { return 1, nil }

func TestPostgresTransactionRepository_CreateAndUpdate(t *testing.T) {
	fc := &fakeConnectionTx{}
	repo := NewPostgresTransactionRepository(fc)
	tx := txEntity.NewTransaction(uuid.New(), txEntity.TransactionTypeDeposit, decimal.NewFromInt(5))
	if err := repo.Create(context.Background(), tx); err != nil {
		t.Fatalf("create err: %v", err)
	}
	if fc.execs == 0 {
		t.Fatalf("esperava Exec create")
	}
	tx.Complete("hash123")
	if err := repo.Update(context.Background(), tx); err != nil {
		t.Fatalf("update err: %v", err)
	}
	if fc.execs < 2 {
		t.Fatalf("esperava Exec update")
	}
	if err := repo.UpdateStatus(context.Background(), tx.ID, txEntity.TransactionStatusFailed); err != nil {
		t.Fatalf("update status err: %v", err)
	}
	if fc.execs < 3 {
		t.Fatalf("esperava Exec update status")
	}
	// FindByID retorna nil devido fakeRow
	out, err := repo.FindByID(context.Background(), tx.ID)
	if err == nil {
		t.Fatalf("esperava erro no rows")
	}
	if out != nil {
		t.Fatalf("esperado nil em fakeRow")
	}
	// FindByUserID increments queries
	list, err := repo.FindByUserID(context.Background(), tx.UserID)
	if err != nil {
		t.Fatalf("find user err: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("fake rows sem Next devem retornar lista vazia")
	}
	if fc.queries == 0 {
		t.Fatalf("esperava queries registradas")
	}
}

package persistence

import (
	"context"
	"database/sql"
	"errors"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	dbIface "financial-system-pro/internal/shared/database"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeConnection implementa database.Connection em memória registrando queries.
type fakeConnection struct{ execCount, queryCount int }

func (f *fakeConnection) Query(ctx context.Context, q string, args ...interface{}) (dbIface.Rows, error) {
	f.queryCount++
	return &fakeRows{}, nil
}
func (f *fakeConnection) QueryRow(ctx context.Context, q string, args ...interface{}) dbIface.Row {
	f.queryCount++
	return &fakeRow{}
}
func (f *fakeConnection) Exec(ctx context.Context, q string, args ...interface{}) (dbIface.Result, error) {
	f.execCount++
	return &fakeResult{}, nil
}
func (f *fakeConnection) Begin(ctx context.Context) (dbIface.Transaction, error) {
	return nil, errors.New("not supported")
}
func (f *fakeConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (dbIface.Transaction, error) {
	return nil, errors.New("not supported")
}
func (f *fakeConnection) Ping(ctx context.Context) error { return nil }
func (f *fakeConnection) Close() error                   { return nil }
func (f *fakeConnection) Stats() (stats sql.DBStats)     { return }

type fakeRows struct{}

func (r *fakeRows) Next() bool                     { return false }
func (r *fakeRows) Scan(dest ...interface{}) error { return nil }
func (r *fakeRows) Close() error                   { return nil }
func (r *fakeRows) Err() error                     { return nil }

type fakeRow struct{}

func (r *fakeRow) Scan(dest ...interface{}) error { return errors.New("no rows") }

type fakeResult struct{}

func (r *fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r *fakeResult) RowsAffected() (int64, error) { return 1, nil }

func TestPostgresUserRepository_CreateAndFind(t *testing.T) {
	fc := &fakeConnection{}
	repo := NewPostgresUserRepository(fc)
	u := &userEntity.User{ID: uuid.New(), Email: "a@b.com", Password: "pw", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("create err: %v", err)
	}
	if fc.execCount == 0 {
		t.Fatalf("esperava Exec chamado")
	}
	out, err := repo.FindByID(context.Background(), u.ID)
	// nossa fakeRow sempre retorna erro "no rows" simulando registro não encontrado
	if err == nil {
		t.Fatalf("esperava erro no rows")
	}
	if out != nil {
		t.Fatalf("esperado nil quando no rows")
	}
}

func TestPostgresWalletRepository_CreateAndBalanceUpdate(t *testing.T) {
	fc := &fakeConnection{}
	repo := NewPostgresWalletRepository(fc)
	w := &userEntity.Wallet{ID: uuid.New(), UserID: uuid.New(), Address: "ADDR", Balance: 10, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.Create(context.Background(), w); err != nil {
		t.Fatalf("create wallet err: %v", err)
	}
	if fc.execCount == 0 {
		t.Fatalf("esperava Exec chamado")
	}
	if err := repo.UpdateBalance(context.Background(), w.UserID, 20); err != nil {
		t.Fatalf("update balance err: %v", err)
	}
	if fc.execCount < 2 {
		t.Fatalf("esperava segundo Exec para update")
	}
	_, _ = repo.FindByUserID(context.Background(), w.UserID) // queryCount++
	if fc.queryCount == 0 {
		t.Fatalf("esperava QueryRow chamado")
	}
}

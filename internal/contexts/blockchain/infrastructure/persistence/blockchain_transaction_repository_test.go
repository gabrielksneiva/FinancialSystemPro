package persistence

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"financial-system-pro/internal/contexts/blockchain/domain/entity"
	dbIface "financial-system-pro/internal/shared/database"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// fakeConnectionBc registra chamadas e simula respostas para o repositório.
type fakeConnectionBc struct{ execs, queries int }

func (f *fakeConnectionBc) Query(ctx context.Context, q string, args ...interface{}) (dbIface.Rows, error) {
	f.queries++
	return &fakeBcRows{}, nil
}
func (f *fakeConnectionBc) QueryRow(ctx context.Context, q string, args ...interface{}) dbIface.Row {
	f.queries++
	return &fakeBcRowNoRows{}
}
func (f *fakeConnectionBc) Exec(ctx context.Context, q string, args ...interface{}) (dbIface.Result, error) {
	f.execs++
	return &fakeBcResult{}, nil
}
func (f *fakeConnectionBc) Begin(ctx context.Context) (dbIface.Transaction, error) {
	return nil, errors.New("not supported")
}
func (f *fakeConnectionBc) BeginTx(ctx context.Context, opts *sql.TxOptions) (dbIface.Transaction, error) {
	return nil, errors.New("not supported")
}
func (f *fakeConnectionBc) Ping(ctx context.Context) error { return nil }
func (f *fakeConnectionBc) Close() error                   { return nil }
func (f *fakeConnectionBc) Stats() sql.DBStats             { return sql.DBStats{} }

type fakeBcRows struct{}

func (r *fakeBcRows) Next() bool                     { return false }
func (r *fakeBcRows) Scan(dest ...interface{}) error { return nil }
func (r *fakeBcRows) Close() error                   { return nil }
func (r *fakeBcRows) Err() error                     { return nil }

// fakeBcRowNoRows retorna sql.ErrNoRows para simular ausência de registro.
type fakeBcRowNoRows struct{}

func (r *fakeBcRowNoRows) Scan(dest ...interface{}) error { return sql.ErrNoRows }

type fakeBcResult struct{}

func (r *fakeBcResult) LastInsertId() (int64, error) { return 0, nil }
func (r *fakeBcResult) RowsAffected() (int64, error) { return 1, nil }

// fakeConnectionBcRowSuccess permite testar caminho de sucesso de FindByID.
type fakeConnectionBcRowSuccess struct{ execs, queries int }

func (f *fakeConnectionBcRowSuccess) Query(ctx context.Context, q string, args ...interface{}) (dbIface.Rows, error) {
	f.queries++
	return &fakeBcRows{}, nil
}
func (f *fakeConnectionBcRowSuccess) QueryRow(ctx context.Context, q string, args ...interface{}) dbIface.Row {
	f.queries++
	return &fakeBcRowSuccess{}
}
func (f *fakeConnectionBcRowSuccess) Exec(ctx context.Context, q string, args ...interface{}) (dbIface.Result, error) {
	f.execs++
	return &fakeBcResult{}, nil
}
func (f *fakeConnectionBcRowSuccess) Begin(ctx context.Context) (dbIface.Transaction, error) {
	return nil, errors.New("not supported")
}
func (f *fakeConnectionBcRowSuccess) BeginTx(ctx context.Context, opts *sql.TxOptions) (dbIface.Transaction, error) {
	return nil, errors.New("not supported")
}
func (f *fakeConnectionBcRowSuccess) Ping(ctx context.Context) error { return nil }
func (f *fakeConnectionBcRowSuccess) Close() error                   { return nil }
func (f *fakeConnectionBcRowSuccess) Stats() sql.DBStats             { return sql.DBStats{} }

type fakeBcRowSuccess struct{}

func (r *fakeBcRowSuccess) Scan(dest ...interface{}) error {
	// Ordem esperada conforme repositório.
	if len(dest) != 13 {
		return errors.New("unexpected dest len")
	}
	if idPtr, ok := dest[0].(*uuid.UUID); ok {
		*idPtr = uuid.New()
	}
	if netPtr, ok := dest[1].(*entity.BlockchainNetwork); ok {
		*netPtr = entity.NetworkTron
	}
	if hashPtr, ok := dest[2].(*string); ok {
		*hashPtr = "txhash123"
	}
	if fromPtr, ok := dest[3].(*string); ok {
		*fromPtr = "FROMADDR"
	}
	if toPtr, ok := dest[4].(*string); ok {
		*toPtr = "TOADDR"
	}
	if amtPtr, ok := dest[5].(*decimal.Decimal); ok {
		*amtPtr = decimal.NewFromInt(42)
	}
	if confPtr, ok := dest[6].(*int); ok {
		*confPtr = 7
	}
	if statusPtr, ok := dest[7].(*string); ok {
		*statusPtr = "confirmed"
	}
	if blockPtr, ok := dest[8].(*int64); ok {
		*blockPtr = 99999
	}
	if gasPtr, ok := dest[9].(*int64); ok {
		*gasPtr = 21000
	}
	now := time.Now()
	if createdPtr, ok := dest[10].(*time.Time); ok {
		*createdPtr = now.Add(-time.Hour)
	}
	if updatedPtr, ok := dest[11].(*time.Time); ok {
		*updatedPtr = now
	}
	if confirmedPtr, ok := dest[12].(*sql.NullTime); ok {
		confirmedPtr.Time = now
		confirmedPtr.Valid = true
	}
	return nil
}

func TestPostgresBlockchainTransactionRepository_CreateUpdateAndConfirmations(t *testing.T) {
	fc := &fakeConnectionBc{}
	repo := NewPostgresBlockchainTransactionRepository(fc)
	tx := entity.NewBlockchainTransaction(entity.NetworkTron, "ADDR1", "ADDR2", decimal.NewFromInt(10))
	if err := repo.Create(context.Background(), tx); err != nil {
		t.Fatalf("create err: %v", err)
	}
	if fc.execs == 0 {
		t.Fatalf("esperava Exec em create")
	}
	tx.Confirm("hashXYZ", 12345, 3)
	if err := repo.Update(context.Background(), tx); err != nil {
		t.Fatalf("update err: %v", err)
	}
	if fc.execs < 2 {
		t.Fatalf("esperava segundo Exec em update")
	}
	if err := repo.UpdateConfirmations(context.Background(), tx.TransactionHash, 8); err != nil {
		t.Fatalf("update confirmations err: %v", err)
	}
	if fc.execs < 3 {
		t.Fatalf("esperava terceiro Exec em update confirmations")
	}
}

func TestPostgresBlockchainTransactionRepository_FindMethods_NoRows(t *testing.T) {
	fc := &fakeConnectionBc{}
	repo := NewPostgresBlockchainTransactionRepository(fc)
	// FindByID deve retornar (nil,nil) quando sql.ErrNoRows
	out, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("esperava err nil quando no rows: %v", err)
	}
	if out != nil {
		t.Fatalf("esperado nil out em no rows")
	}
	// FindByHash mesmo comportamento
	out2, err := repo.FindByHash(context.Background(), "unknown")
	if err != nil || out2 != nil {
		t.Fatalf("FindByHash no rows deve retornar nil,nil; got out=%v err=%v", out2, err)
	}
	// FindByAddress retorna slice vazio
	list, err := repo.FindByAddress(context.Background(), "ADDR-X")
	if err != nil {
		t.Fatalf("find address err: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("esperado slice vazio quando rows.Next false")
	}
	if fc.queries < 3 { // duas QueryRow + uma Query
		t.Fatalf("esperava queries registradas >=3, got %d", fc.queries)
	}
}

func TestPostgresBlockchainTransactionRepository_FindByID_Success(t *testing.T) {
	fc := &fakeConnectionBcRowSuccess{}
	repo := NewPostgresBlockchainTransactionRepository(fc)
	tx, err := repo.FindByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("esperava sucesso sem erro: %v", err)
	}
	if tx == nil || tx.TransactionHash != "txhash123" || tx.ConfirmedAt == nil {
		t.Fatalf("transação esperada preenchida e confirmada; got %#v", tx)
	}
	if fc.queries == 0 {
		t.Fatalf("esperava incremento queries")
	}
}

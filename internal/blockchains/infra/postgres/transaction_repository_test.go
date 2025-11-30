package postgres

import (
	"context"
	"testing"
	"time"

	"financial-system-pro/internal/blockchains/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestTransactionRepository_CreateAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock err: %v", err)
	}
	defer db.Close()
	sx := sqlx.NewDb(db, "postgres")

	// Expect insert
	mock.ExpectExec(`INSERT INTO transactions`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewTransactionRepository(sx)
	now := time.Now()
	amt, _ := domain.NewAmountFromString("500")
	tx := domain.Transaction{From: "f", To: "t", Amount: amt, Status: domain.TxPending, CreatedAt: now}
	created, err := repo.CreateTransaction(context.Background(), tx)
	if err != nil {
		t.Fatalf("create err: %v", err)
	}
	if created.Hash == "" {
		t.Fatalf("expected generated hash")
	}

	// Expect select
	rows := sqlmock.NewRows([]string{"hash", "from_addr", "to_addr", "amount", "status", "block_num", "created_at"}).AddRow(created.Hash, "f", "t", "500", string(domain.TxPending), 0, now)
	mock.ExpectQuery(`SELECT hash, from_addr, to_addr, amount, status, block_num, created_at FROM transactions WHERE hash=\$1 LIMIT 1`).WithArgs(created.Hash).WillReturnRows(rows)

	got, err := repo.GetTransaction(context.Background(), created.Hash)
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
	if got.Hash != created.Hash {
		t.Fatalf("expected same hash")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

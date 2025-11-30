package postgres

import (
	"context"
	"testing"

	"financial-system-pro/internal/blockchains/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestBalanceRepository_GetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock err: %v", err)
	}
	defer db.Close()
	sx := sqlx.NewDb(db, "postgres")

	// expected query
	rows := sqlmock.NewRows([]string{"amount"}).AddRow("12345")
	mock.ExpectQuery(`SELECT amount FROM balances WHERE address=\$1 LIMIT 1`).WithArgs("addr1").WillReturnRows(rows)

	repo := NewBalanceRepository(sx)
	amt, err := repo.GetBalance(context.Background(), domain.Address("addr1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if amt.Value.String() != "12345" {
		t.Fatalf("expected 12345 got %s", amt.Value.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBalanceRepository_DbNil(t *testing.T) {
	repo := NewBalanceRepository(nil)
	if _, err := repo.GetBalance(context.Background(), domain.Address("a")); err == nil {
		t.Fatalf("expected error when db nil")
	}
}

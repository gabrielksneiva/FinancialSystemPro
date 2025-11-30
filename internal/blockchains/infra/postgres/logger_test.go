package postgres

import (
    "context"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"
)

func TestPGLogger_EnsureSchemaAndLogEvent(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("sqlmock err: %v", err)
    }
    defer db.Close()
    sx := sqlx.NewDb(db, "postgres")
    p := NewPGLogger(sx)

    mock.ExpectExec(`CREATE TABLE IF NOT EXISTS events_log`).WillReturnResult(sqlmock.NewResult(1, 1))
    if err := p.EnsureSchema(context.Background()); err != nil {
        t.Fatalf("ensure schema err: %v", err)
    }

    mock.ExpectExec(`INSERT INTO events_log`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
    if err := p.LogEvent(context.Background(), "t", "p"); err != nil {
        t.Fatalf("log event err: %v", err)
    }

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatalf("unmet expectations: %v", err)
    }
}

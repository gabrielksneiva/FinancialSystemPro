package database

import (
	"context"
	"database/sql"
)

// Connection representa uma conexão com o banco de dados
type Connection interface {
	// Query operations
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)

	// Transaction operations
	Begin(ctx context.Context) (Transaction, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)

	// Connection management
	Ping(ctx context.Context) error
	Close() error
	Stats() sql.DBStats
}

// Transaction representa uma transação de banco de dados
type Transaction interface {
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	Commit() error
	Rollback() error
}

// Rows representa múltiplas linhas de resultado
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

// Row representa uma única linha de resultado
type Row interface {
	Scan(dest ...interface{}) error
}

// Result representa o resultado de uma operação de modificação
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Repository é a interface base para todos os repositórios
type Repository interface {
	GetConnection() Connection
	WithContext(ctx context.Context) Repository
}

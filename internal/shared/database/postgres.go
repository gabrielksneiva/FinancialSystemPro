package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresConnection implementa Connection para PostgreSQL
type PostgresConnection struct {
	db *sql.DB
}

// NewPostgresConnection cria uma nova conexão PostgreSQL
func NewPostgresConnection(connectionString string) (Connection, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configurações de connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Testar conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresConnection{db: db}, nil
}

// Query executa uma query que retorna múltiplas linhas
func (p *PostgresConnection) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PostgresRows{rows: rows}, nil
}

// QueryRow executa uma query que retorna uma única linha
func (p *PostgresConnection) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return &PostgresRow{row: p.db.QueryRowContext(ctx, query, args...)}
}

// Exec executa uma query de modificação (INSERT, UPDATE, DELETE)
func (p *PostgresConnection) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PostgresResult{result: result}, nil
}

// Begin inicia uma transação
func (p *PostgresConnection) Begin(ctx context.Context) (Transaction, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &PostgresTransaction{tx: tx}, nil
}

// BeginTx inicia uma transação com opções customizadas
func (p *PostgresConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
	tx, err := p.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &PostgresTransaction{tx: tx}, nil
}

// Ping verifica se a conexão está ativa
func (p *PostgresConnection) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Close fecha a conexão
func (p *PostgresConnection) Close() error {
	return p.db.Close()
}

// Stats retorna estatísticas da connection pool
func (p *PostgresConnection) Stats() sql.DBStats {
	return p.db.Stats()
}

// PostgresTransaction implementa Transaction
type PostgresTransaction struct {
	tx *sql.Tx
}

func (pt *PostgresTransaction) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := pt.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PostgresRows{rows: rows}, nil
}

func (pt *PostgresTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return &PostgresRow{row: pt.tx.QueryRowContext(ctx, query, args...)}
}

func (pt *PostgresTransaction) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := pt.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PostgresResult{result: result}, nil
}

func (pt *PostgresTransaction) Commit() error {
	return pt.tx.Commit()
}

func (pt *PostgresTransaction) Rollback() error {
	return pt.tx.Rollback()
}

// PostgresRows implementa Rows
type PostgresRows struct {
	rows *sql.Rows
}

func (pr *PostgresRows) Next() bool {
	return pr.rows.Next()
}

func (pr *PostgresRows) Scan(dest ...interface{}) error {
	return pr.rows.Scan(dest...)
}

func (pr *PostgresRows) Close() error {
	return pr.rows.Close()
}

func (pr *PostgresRows) Err() error {
	return pr.rows.Err()
}

// PostgresRow implementa Row
type PostgresRow struct {
	row *sql.Row
}

func (pr *PostgresRow) Scan(dest ...interface{}) error {
	return pr.row.Scan(dest...)
}

// PostgresResult implementa Result
type PostgresResult struct {
	result sql.Result
}

func (pr *PostgresResult) LastInsertId() (int64, error) {
	return pr.result.LastInsertId()
}

func (pr *PostgresResult) RowsAffected() (int64, error) {
	return pr.result.RowsAffected()
}

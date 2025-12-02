package finance

import (
	"context"
	"database/sql"

	"financial-system-pro/internal/domain/finance"
)

// PostgresLedgerRepository implements LedgerRepository using PostgreSQL append-only table.
type PostgresLedgerRepository struct {
	db *sql.DB
}

func NewPostgresLedgerRepository(db *sql.DB) *PostgresLedgerRepository {
	return &PostgresLedgerRepository{db: db}
}

func (r *PostgresLedgerRepository) Append(ctx context.Context, e *finance.LedgerEntry) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO finance_ledger_entries 
		(id, wallet_id, entry_type, amount_value, amount_currency, balance_after, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, e.ID(), e.WalletID(), string(e.Type()), e.Amount().Value(), e.Amount().Currency(),
		e.BalanceAfter(), e.Description(), e.CreatedAt())
	return err
}

func (r *PostgresLedgerRepository) ListByWallet(ctx context.Context, walletID string) ([]*finance.LedgerEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, wallet_id, entry_type, amount_value, amount_currency, balance_after, description, created_at
		FROM finance_ledger_entries
		WHERE wallet_id = $1
		ORDER BY created_at ASC
	`, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

func (r *PostgresLedgerRepository) ListSince(ctx context.Context, walletID string, sinceID string) ([]*finance.LedgerEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, wallet_id, entry_type, amount_value, amount_currency, balance_after, description, created_at
		FROM finance_ledger_entries
		WHERE wallet_id = $1 AND id > $2
		ORDER BY created_at ASC
	`, walletID, sinceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

func (r *PostgresLedgerRepository) scanEntries(rows *sql.Rows) ([]*finance.LedgerEntry, error) {
	var entries []*finance.LedgerEntry
	for rows.Next() {
		var id, walletID, entryType, curr, desc string
		var amtVal, balAfter int64
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &walletID, &entryType, &amtVal, &curr, &balAfter, &desc, &createdAt); err != nil {
			return nil, err
		}

		amt, err := finance.NewAmount(amtVal, curr)
		if err != nil {
			return nil, err
		}

		entry, err := finance.NewLedgerEntry(id, walletID, finance.EntryType(entryType), amt, balAfter, desc)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// PostgresTransactional implements Transactional for wrapping operations in DB transactions.
type PostgresTransactional struct {
	db *sql.DB
}

func NewPostgresTransactional(db *sql.DB) *PostgresTransactional {
	return &PostgresTransactional{db: db}
}

type txKey struct{}

func (t *PostgresTransactional) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func getTx(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}

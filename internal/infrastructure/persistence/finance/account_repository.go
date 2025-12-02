package finance

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"financial-system-pro/internal/domain/finance"
)

// PostgresAccountRepository implements AccountRepository using PostgreSQL.
type PostgresAccountRepository struct {
	db *sql.DB
}

func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) Get(ctx context.Context, id string) (*finance.Account, error) {
	var holderID, status string
	err := r.db.QueryRowContext(ctx, `SELECT holder_id, status FROM finance_accounts WHERE id = $1`, id).
		Scan(&holderID, &status)
	if err == sql.ErrNoRows {
		return nil, errors.New("account not found")
	}
	if err != nil {
		return nil, err
	}

	acc, err := finance.NewAccount(id, holderID)
	if err != nil {
		return nil, err
	}

	if status == "inactive" {
		_ = acc.Deactivate()
	}

	return acc, nil
}

func (r *PostgresAccountRepository) Save(ctx context.Context, a *finance.Account) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO finance_accounts (id, holder_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			holder_id = EXCLUDED.holder_id,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, a.ID(), a.HolderID(), string(a.Status()), time.Now().UTC(), time.Now().UTC())
	return err
}

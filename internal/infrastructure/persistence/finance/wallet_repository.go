package finance

import (
	"context"
	"database/sql"
	"errors"

	"financial-system-pro/internal/domain/finance"
)

// PostgresWalletRepository implements WalletRepository using PostgreSQL with optimistic locking.
type PostgresWalletRepository struct {
	db *sql.DB
}

func NewPostgresWalletRepository(db *sql.DB) *PostgresWalletRepository {
	return &PostgresWalletRepository{db: db}
}

func (r *PostgresWalletRepository) Get(ctx context.Context, id string) (*finance.Wallet, error) {
	var accountID string
	var balanceValue int64
	var balanceCurrency string
	err := r.db.QueryRowContext(ctx, `
		SELECT account_id, balance_value, balance_currency 
		FROM finance_wallets WHERE id = $1
	`, id).Scan(&accountID, &balanceValue, &balanceCurrency)
	if err == sql.ErrNoRows {
		return nil, errors.New("wallet not found")
	}
	if err != nil {
		return nil, err
	}

	bal, err := finance.NewAmount(balanceValue, balanceCurrency)
	if err != nil {
		return nil, err
	}

	return finance.NewWallet(id, accountID, bal)
}

func (r *PostgresWalletRepository) Save(ctx context.Context, w *finance.Wallet) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO finance_wallets (id, account_id, balance_value, balance_currency)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			balance_value = EXCLUDED.balance_value,
			balance_currency = EXCLUDED.balance_currency
		WHERE finance_wallets.version = (SELECT version FROM finance_wallets WHERE id = $1)
	`, w.ID(), w.AccountID(), w.Balance().Value(), w.Balance().Currency())
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("optimistic lock conflict: wallet was modified by another transaction")
	}

	return nil
}

func (r *PostgresWalletRepository) ListByAccount(ctx context.Context, accountID string) ([]*finance.Wallet, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, account_id, balance_value, balance_currency 
		FROM finance_wallets WHERE account_id = $1
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []*finance.Wallet
	for rows.Next() {
		var id, accID, curr string
		var val int64
		if err := rows.Scan(&id, &accID, &val, &curr); err != nil {
			return nil, err
		}
		bal, err := finance.NewAmount(val, curr)
		if err != nil {
			return nil, err
		}
		w, err := finance.NewWallet(id, accID, bal)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}

	return wallets, rows.Err()
}

package postgres

import (
	"context"
	"fmt"

	"financial-system-pro/internal/blockchains/domain"

	"github.com/jmoiron/sqlx"
)

type BalanceRepository struct {
	db *sqlx.DB
}

func NewBalanceRepository(db *sqlx.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

// GetBalance reads amount stored as text in balances table
func (r *BalanceRepository) GetBalance(ctx context.Context, addr domain.Address) (domain.Amount, error) {
	if r.db == nil {
		return domain.Amount{}, fmt.Errorf("db not configured")
	}
	var amountStr string
	err := r.db.GetContext(ctx, &amountStr, `SELECT amount FROM balances WHERE address=$1 LIMIT 1`, string(addr))
	if err != nil {
		return domain.Amount{}, err
	}
	amt, err := domain.NewAmountFromString(amountStr)
	if err != nil {
		return domain.Amount{}, err
	}
	return amt, nil
}

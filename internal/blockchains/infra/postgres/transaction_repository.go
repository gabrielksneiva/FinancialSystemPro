package postgres

import (
	"context"
	"fmt"
	"time"

	"financial-system-pro/internal/blockchains/domain"

	"github.com/jmoiron/sqlx"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	if r.db == nil {
		return domain.Transaction{}, fmt.Errorf("db not configured")
	}
	if tx.Hash == "" {
		// generate a simple hash for storage test; real implementation should use proper tx id
		tx.Hash = "tx_" + time.Now().Format("20060102150405")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO transactions (hash, from_addr, to_addr, amount, status, block_num, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, tx.Hash, string(tx.From), string(tx.To), tx.Amount.Value.String(), string(tx.Status), tx.BlockNum, tx.CreatedAt)
	if err != nil {
		return domain.Transaction{}, err
	}
	return tx, nil
}

func (r *TransactionRepository) GetTransaction(ctx context.Context, hash string) (domain.Transaction, error) {
	if r.db == nil {
		return domain.Transaction{}, fmt.Errorf("db not configured")
	}
	var t struct {
		Hash      string    `db:"hash"`
		FromAddr  string    `db:"from_addr"`
		ToAddr    string    `db:"to_addr"`
		AmountStr string    `db:"amount"`
		Status    string    `db:"status"`
		BlockNum  uint64    `db:"block_num"`
		CreatedAt time.Time `db:"created_at"`
	}
	err := r.db.GetContext(ctx, &t, `SELECT hash, from_addr, to_addr, amount, status, block_num, created_at FROM transactions WHERE hash=$1 LIMIT 1`, hash)
	if err != nil {
		return domain.Transaction{}, err
	}
	amt, err := domain.NewAmountFromString(t.AmountStr)
	if err != nil {
		return domain.Transaction{}, err
	}
	from, _ := domain.NewAddress(t.FromAddr)
	to, _ := domain.NewAddress(t.ToAddr)
	return domain.Transaction{Hash: t.Hash, From: from, To: to, Amount: amt, Status: domain.TxStatus(t.Status), BlockNum: t.BlockNum, CreatedAt: t.CreatedAt}, nil
}

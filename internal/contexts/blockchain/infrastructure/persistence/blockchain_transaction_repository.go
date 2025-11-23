package persistence

import (
	"context"
	"database/sql"
	"errors"
	"financial-system-pro/internal/contexts/blockchain/domain/entity"
	"financial-system-pro/internal/shared/database"
	"time"

	"github.com/google/uuid"
)

// PostgresBlockchainTransactionRepository implementa BlockchainTransactionRepository
type PostgresBlockchainTransactionRepository struct {
	conn   database.Connection
	schema string
}

// NewPostgresBlockchainTransactionRepository cria um novo repositório
func NewPostgresBlockchainTransactionRepository(conn database.Connection) *PostgresBlockchainTransactionRepository {
	return &PostgresBlockchainTransactionRepository{
		conn:   conn,
		schema: "blockchain_context",
	}
}

// Create insere uma nova transação blockchain
func (r *PostgresBlockchainTransactionRepository) Create(ctx context.Context, tx *entity.BlockchainTransaction) error {
	query := `
		INSERT INTO ` + r.schema + `.blockchain_transactions 
		(id, network, transaction_hash, from_address, to_address, amount, confirmations, 
		 status, block_number, gas_used, created_at, updated_at, confirmed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.conn.Exec(ctx, query,
		tx.ID,
		tx.Network,
		tx.TransactionHash,
		tx.FromAddress,
		tx.ToAddress,
		tx.Amount,
		tx.Confirmations,
		tx.Status,
		tx.BlockNumber,
		tx.GasUsed,
		tx.CreatedAt,
		tx.UpdatedAt,
		tx.ConfirmedAt,
	)

	return err
}

// FindByID busca uma transação por ID
func (r *PostgresBlockchainTransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.BlockchainTransaction, error) {
	query := `
		SELECT id, network, transaction_hash, from_address, to_address, amount, confirmations,
		       status, block_number, gas_used, created_at, updated_at, confirmed_at
		FROM ` + r.schema + `.blockchain_transactions
		WHERE id = $1
	`

	tx := &entity.BlockchainTransaction{}
	var confirmedAt sql.NullTime

	err := r.conn.QueryRow(ctx, query, id).Scan(
		&tx.ID,
		&tx.Network,
		&tx.TransactionHash,
		&tx.FromAddress,
		&tx.ToAddress,
		&tx.Amount,
		&tx.Confirmations,
		&tx.Status,
		&tx.BlockNumber,
		&tx.GasUsed,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&confirmedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if confirmedAt.Valid {
		tx.ConfirmedAt = &confirmedAt.Time
	}

	return tx, nil
}

// FindByHash busca uma transação por hash
func (r *PostgresBlockchainTransactionRepository) FindByHash(ctx context.Context, hash string) (*entity.BlockchainTransaction, error) {
	query := `
		SELECT id, network, transaction_hash, from_address, to_address, amount, confirmations,
		       status, block_number, gas_used, created_at, updated_at, confirmed_at
		FROM ` + r.schema + `.blockchain_transactions
		WHERE transaction_hash = $1
	`

	tx := &entity.BlockchainTransaction{}
	var confirmedAt sql.NullTime

	err := r.conn.QueryRow(ctx, query, hash).Scan(
		&tx.ID,
		&tx.Network,
		&tx.TransactionHash,
		&tx.FromAddress,
		&tx.ToAddress,
		&tx.Amount,
		&tx.Confirmations,
		&tx.Status,
		&tx.BlockNumber,
		&tx.GasUsed,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&confirmedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if confirmedAt.Valid {
		tx.ConfirmedAt = &confirmedAt.Time
	}

	return tx, nil
}

// FindByAddress busca todas as transações de um endereço
func (r *PostgresBlockchainTransactionRepository) FindByAddress(ctx context.Context, address string) ([]*entity.BlockchainTransaction, error) {
	query := `
		SELECT id, network, transaction_hash, from_address, to_address, amount, confirmations,
		       status, block_number, gas_used, created_at, updated_at, confirmed_at
		FROM ` + r.schema + `.blockchain_transactions
		WHERE from_address = $1 OR to_address = $1
		ORDER BY created_at DESC
	`

	rows, err := r.conn.Query(ctx, query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*entity.BlockchainTransaction

	for rows.Next() {
		tx := &entity.BlockchainTransaction{}
		var confirmedAt sql.NullTime

		err := rows.Scan(
			&tx.ID,
			&tx.Network,
			&tx.TransactionHash,
			&tx.FromAddress,
			&tx.ToAddress,
			&tx.Amount,
			&tx.Confirmations,
			&tx.Status,
			&tx.BlockNumber,
			&tx.GasUsed,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&confirmedAt,
		)

		if err != nil {
			return nil, err
		}

		if confirmedAt.Valid {
			tx.ConfirmedAt = &confirmedAt.Time
		}

		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

// Update atualiza uma transação existente
func (r *PostgresBlockchainTransactionRepository) Update(ctx context.Context, tx *entity.BlockchainTransaction) error {
	query := `
		UPDATE ` + r.schema + `.blockchain_transactions
		SET confirmations = $2, status = $3, block_number = $4, 
		    gas_used = $5, updated_at = $6, confirmed_at = $7
		WHERE id = $1
	`

	tx.UpdatedAt = time.Now()

	_, err := r.conn.Exec(ctx, query,
		tx.ID,
		tx.Confirmations,
		tx.Status,
		tx.BlockNumber,
		tx.GasUsed,
		tx.UpdatedAt,
		tx.ConfirmedAt,
	)

	return err
}

// UpdateConfirmations atualiza apenas as confirmações de uma transação
func (r *PostgresBlockchainTransactionRepository) UpdateConfirmations(ctx context.Context, hash string, confirmations int) error {
	query := `
		UPDATE ` + r.schema + `.blockchain_transactions
		SET confirmations = $2, updated_at = $3
		WHERE transaction_hash = $1
	`

	_, err := r.conn.Exec(ctx, query, hash, confirmations, time.Now())
	return err
}

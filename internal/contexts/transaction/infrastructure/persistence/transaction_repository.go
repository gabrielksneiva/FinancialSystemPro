package persistence

import (
	"context"
	"database/sql"
	"errors"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	"financial-system-pro/internal/shared/database"
	"time"

	"github.com/google/uuid"
)

// PostgresTransactionRepository implementa TransactionRepository usando PostgreSQL
type PostgresTransactionRepository struct {
	conn   database.Connection
	schema string
}

// NewPostgresTransactionRepository cria um novo repositório de transações
func NewPostgresTransactionRepository(conn database.Connection) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{
		conn:   conn,
		schema: "transaction_context",
	}
}

// Create insere uma nova transação no banco
func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	query := `
		INSERT INTO ` + r.schema + `.transactions 
		(id, user_id, type, amount, status, transaction_hash, from_address, to_address, 
		 callback_url, error_message, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.conn.Exec(ctx, query,
		tx.ID,
		tx.UserID,
		tx.Type,
		tx.Amount,
		tx.Status,
		tx.TransactionHash,
		tx.FromAddress,
		tx.ToAddress,
		tx.CallbackURL,
		tx.ErrorMessage,
		tx.CreatedAt,
		tx.UpdatedAt,
		tx.CompletedAt,
	)

	return err
}

// FindByID busca uma transação por ID
func (r *PostgresTransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, status, transaction_hash, from_address, to_address,
		       callback_url, error_message, created_at, updated_at, completed_at
		FROM ` + r.schema + `.transactions
		WHERE id = $1
	`

	tx := &entity.Transaction{}
	var completedAt sql.NullTime

	err := r.conn.QueryRow(ctx, query, id).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.Type,
		&tx.Amount,
		&tx.Status,
		&tx.TransactionHash,
		&tx.FromAddress,
		&tx.ToAddress,
		&tx.CallbackURL,
		&tx.ErrorMessage,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&completedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if completedAt.Valid {
		tx.CompletedAt = &completedAt.Time
	}

	return tx, nil
}

// FindByUserID busca todas as transações de um usuário
func (r *PostgresTransactionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, status, transaction_hash, from_address, to_address,
		       callback_url, error_message, created_at, updated_at, completed_at
		FROM ` + r.schema + `.transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*entity.Transaction

	for rows.Next() {
		tx := &entity.Transaction{}
		var completedAt sql.NullTime

		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.Type,
			&tx.Amount,
			&tx.Status,
			&tx.TransactionHash,
			&tx.FromAddress,
			&tx.ToAddress,
			&tx.CallbackURL,
			&tx.ErrorMessage,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&completedAt,
		)

		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			tx.CompletedAt = &completedAt.Time
		}

		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// FindByHash busca uma transação por hash
func (r *PostgresTransactionRepository) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, status, transaction_hash, from_address, to_address,
		       callback_url, error_message, created_at, updated_at, completed_at
		FROM ` + r.schema + `.transactions
		WHERE transaction_hash = $1
	`

	tx := &entity.Transaction{}
	var completedAt sql.NullTime

	err := r.conn.QueryRow(ctx, query, hash).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.Type,
		&tx.Amount,
		&tx.Status,
		&tx.TransactionHash,
		&tx.FromAddress,
		&tx.ToAddress,
		&tx.CallbackURL,
		&tx.ErrorMessage,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&completedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if completedAt.Valid {
		tx.CompletedAt = &completedAt.Time
	}

	return tx, nil
}

// Update atualiza uma transação existente
func (r *PostgresTransactionRepository) Update(ctx context.Context, tx *entity.Transaction) error {
	query := `
		UPDATE ` + r.schema + `.transactions
		SET status = $2, transaction_hash = $3, error_message = $4, 
		    updated_at = $5, completed_at = $6
		WHERE id = $1
	`

	tx.UpdatedAt = time.Now()

	_, err := r.conn.Exec(ctx, query,
		tx.ID,
		tx.Status,
		tx.TransactionHash,
		tx.ErrorMessage,
		tx.UpdatedAt,
		tx.CompletedAt,
	)

	return err
}

// UpdateStatus atualiza apenas o status de uma transação
func (r *PostgresTransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TransactionStatus) error {
	query := `
		UPDATE ` + r.schema + `.transactions
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.conn.Exec(ctx, query, id, status, time.Now())
	return err
}

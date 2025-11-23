package persistence

import (
	"context"
	"database/sql"
	"errors"
	"financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/shared/database"
	"time"

	"github.com/google/uuid"
)

// PostgresWalletRepository implementa WalletRepository usando PostgreSQL
type PostgresWalletRepository struct {
	conn   database.Connection
	schema string
}

// NewPostgresWalletRepository cria um novo repositório de wallets
func NewPostgresWalletRepository(conn database.Connection) *PostgresWalletRepository {
	return &PostgresWalletRepository{
		conn:   conn,
		schema: "user_context",
	}
}

// Create insere uma nova wallet no banco
func (r *PostgresWalletRepository) Create(ctx context.Context, wallet *entity.Wallet) error {
	query := `
		INSERT INTO ` + r.schema + `.wallet_info 
		(id, user_id, address, encrypted_private_key, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.conn.Exec(ctx, query,
		wallet.ID,
		wallet.UserID,
		wallet.Address,
		wallet.EncryptedPrivKey,
		wallet.Balance,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)

	return err
}

// FindByUserID busca a wallet de um usuário
func (r *PostgresWalletRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	query := `
		SELECT id, user_id, address, encrypted_private_key, balance, created_at, updated_at
		FROM ` + r.schema + `.wallet_info
		WHERE user_id = $1
	`

	wallet := &entity.Wallet{}
	err := r.conn.QueryRow(ctx, query, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Address,
		&wallet.EncryptedPrivKey,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return wallet, nil
}

// FindByAddress busca uma wallet por endereço
func (r *PostgresWalletRepository) FindByAddress(ctx context.Context, address string) (*entity.Wallet, error) {
	query := `
		SELECT id, user_id, address, encrypted_private_key, balance, created_at, updated_at
		FROM ` + r.schema + `.wallet_info
		WHERE address = $1
	`

	wallet := &entity.Wallet{}
	err := r.conn.QueryRow(ctx, query, address).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Address,
		&wallet.EncryptedPrivKey,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return wallet, nil
}

// UpdateBalance atualiza o saldo de uma wallet
func (r *PostgresWalletRepository) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	query := `
		UPDATE ` + r.schema + `.wallet_info
		SET balance = $2, updated_at = $3
		WHERE user_id = $1
	`

	_, err := r.conn.Exec(ctx, query, userID, balance, time.Now())
	return err
}

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

// PostgresUserRepository implementa UserRepository usando PostgreSQL
type PostgresUserRepository struct {
	conn   database.Connection
	schema string
}

// NewPostgresUserRepository cria um novo repositório de usuários
func NewPostgresUserRepository(conn database.Connection) *PostgresUserRepository {
	return &PostgresUserRepository{
		conn:   conn,
		schema: "user_context",
	}
}

// Create insere um novo usuário no banco
func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO ` + r.schema + `.users (id, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.conn.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// FindByID busca um usuário por ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password, created_at, updated_at
		FROM ` + r.schema + `.users
		WHERE id = $1
	`

	user := &entity.User{}
	err := r.conn.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// FindByEmail busca um usuário por email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password, created_at, updated_at
		FROM ` + r.schema + `.users
		WHERE email = $1
	`

	user := &entity.User{}
	err := r.conn.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// Update atualiza um usuário existente
func (r *PostgresUserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE ` + r.schema + `.users
		SET email = $2, password = $3, updated_at = $4
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.conn.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.UpdatedAt,
	)

	return err
}

// Delete remove um usuário
func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ` + r.schema + `.users WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	return err
}

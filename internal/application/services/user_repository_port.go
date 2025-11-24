package services

import (
	"context"
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
)

// UserRepositoryPort define operações específicas de usuário, segregando do DatabasePort genérico.
// Esta interface será usada por AuthService e UserService para reduzir acoplamento.
type UserRepositoryPort interface {
	FindByEmail(ctx context.Context, email string) (*repositories.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*repositories.User, error)
	Save(ctx context.Context, user *repositories.User) error
}

// userRepositoryAdapter implementa UserRepositoryPort reutilizando DatabasePort existente.
// Permite migração incremental sem tocar em infraestrutura imediata.
type userRepositoryAdapter struct{ db DatabasePort }

func NewUserRepositoryAdapter(db DatabasePort) UserRepositoryPort {
	return &userRepositoryAdapter{db: db}
}

func (a *userRepositoryAdapter) FindByEmail(ctx context.Context, email string) (*repositories.User, error) {
	return a.db.FindUserByField("email", email)
}

func (a *userRepositoryAdapter) FindByID(ctx context.Context, id uuid.UUID) (*repositories.User, error) {
	return a.db.FindUserByField("id", id.String())
}

func (a *userRepositoryAdapter) Save(ctx context.Context, user *repositories.User) error {
	return a.db.Insert(user)
}

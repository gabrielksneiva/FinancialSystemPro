package application

import (
	"context"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"

	"github.com/google/uuid"
)

// UserRepository define operações de persistência do agregado User.
// Mantém foco em invariantes de usuário e evita métodos genéricos (FindUserByField).
type UserRepository interface {
	Save(ctx context.Context, user *userEntity.User) error
	FindByEmail(ctx context.Context, email string) (*userEntity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error)
}

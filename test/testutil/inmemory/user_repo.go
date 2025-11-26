package inmemory

import (
	"context"
	"sync"

	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/domain/errors"

	"github.com/google/uuid"
)

type UserRepository struct {
	mu      sync.RWMutex
	users   map[uuid.UUID]*userEntity.User
	byEmail map[string]*userEntity.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:   make(map[uuid.UUID]*userEntity.User),
		byEmail: make(map[string]*userEntity.User),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *userEntity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byEmail[string(user.Email)]; exists {
		return errors.NewValidationError("email", "email already registered")
	}
	r.users[user.ID] = user
	r.byEmail[string(user.Email)] = user
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, errors.NewNotFoundError("user")
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.byEmail[email]
	if !exists {
		return nil, errors.NewNotFoundError("user")
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *userEntity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return errors.NewNotFoundError("user")
	}
	if old := r.users[user.ID]; old.Email != user.Email {
		delete(r.byEmail, string(old.Email))
		r.byEmail[string(user.Email)] = user
	}
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.users[id]
	if !exists {
		return errors.NewNotFoundError("user")
	}
	delete(r.byEmail, string(user.Email))
	delete(r.users, id)
	return nil
}

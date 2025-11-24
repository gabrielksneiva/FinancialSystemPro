package inmemory

import (
	"context"
	"testing"
	"time"

	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	domainErrors "financial-system-pro/internal/domain/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryCRUD(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()

	u := userEntity.NewUser("test@example.com", "pass")
	require.NoError(t, repo.Create(ctx, u))

	// duplicate email
	dup := userEntity.NewUser("test@example.com", "other")
	err := repo.Create(ctx, dup)
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrValidation, err.(*domainErrors.AppError).Code)

	// find by id
	foundByID, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, u.Email, foundByID.Email)

	// find by email
	foundByEmail, err := repo.FindByEmail(ctx, u.Email)
	require.NoError(t, err)
	require.Equal(t, u.ID, foundByEmail.ID)

	// update email (use copy to trigger reindex logic)
	updated := *u
	updated.Email = "updated@example.com"
	updated.UpdatedAt = time.Now()
	require.NoError(t, repo.Update(ctx, &updated))

	// old email should not be found
	_, err = repo.FindByEmail(ctx, "test@example.com")
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrRecordNotFound, err.(*domainErrors.AppError).Code)

	// new email should be found
	_, err = repo.FindByEmail(ctx, "updated@example.com")
	require.NoError(t, err)

	// delete
	require.NoError(t, repo.Delete(ctx, u.ID))
	_, err = repo.FindByID(ctx, u.ID)
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrRecordNotFound, err.(*domainErrors.AppError).Code)
}

func TestUserRepositoryNotFound(t *testing.T) {
	repo := NewUserRepository()
	ctx := context.Background()
	_, err := repo.FindByID(ctx, uuid.New())
	require.Error(t, err)
	require.Equal(t, domainErrors.ErrRecordNotFound, err.(*domainErrors.AppError).Code)
}

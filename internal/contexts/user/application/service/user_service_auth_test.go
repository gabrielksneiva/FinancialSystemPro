package service

import (
	"context"
	"financial-system-pro/internal/contexts/user/domain/entity"
	userRepo "financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/contexts/user/domain/valueobject"
	"financial-system-pro/internal/shared/events"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// minimal in-memory user & wallet repos for auth test
type authTestUserRepo struct{ users map[uuid.UUID]*entity.User }

func newAuthTestUserRepo() *authTestUserRepo {
	return &authTestUserRepo{users: make(map[uuid.UUID]*entity.User)}
}
func (r *authTestUserRepo) Create(ctx context.Context, u *entity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *authTestUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return r.users[id], nil
}
func (r *authTestUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	for _, u := range r.users {
		if u.Email.String() == email {
			return u, nil
		}
	}
	return nil, nil
}
func (r *authTestUserRepo) Update(ctx context.Context, u *entity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *authTestUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

var _ userRepo.UserRepository = (*authTestUserRepo)(nil)

type authTestWalletRepo struct{}

func (authTestWalletRepo) Create(ctx context.Context, w *entity.Wallet) error { return nil }
func (authTestWalletRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	return nil, nil
}
func (authTestWalletRepo) FindByAddress(ctx context.Context, address string) (*entity.Wallet, error) {
	return nil, nil
}
func (authTestWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	return nil
}

var _ userRepo.WalletRepository = (*authTestWalletRepo)(nil)

func TestAuthenticate_UserNotFoundReturnsInvalidCredentials(t *testing.T) {
	logger := zap.NewNop()
	bus := events.NewInMemoryBus(logger)
	usrRepo := newAuthTestUserRepo() // empty
	walletRepo := authTestWalletRepo{}
	svc := NewUserService(usrRepo, walletRepo, bus, logger)
	_, err := svc.Authenticate(context.Background(), "missing@test.com", "pw")
	if err == nil || err.Error() != ErrInvalidCredentials.Error() {
		t.Fatalf("expected ErrInvalidCredentials got %v", err)
	}
}

func TestAuthenticate_Sucesso(t *testing.T) {
	t.Setenv("SECRET_KEY", "test-secret")
	logger := zap.NewNop()
	bus := events.NewInMemoryBus(logger)
	usrRepo := newAuthTestUserRepo()
	walletRepo := authTestWalletRepo{}
	plain := "senha123"
	emailVO, _ := valueobject.NewEmail("ok@test.com")
	hashedVO, _ := valueobject.HashFromRaw(plain)
	u := entity.NewUser(emailVO, hashedVO)
	_ = usrRepo.Create(context.Background(), u)
	svc := NewUserService(usrRepo, walletRepo, bus, logger)
	user, err := svc.Authenticate(context.Background(), "ok@test.com", plain)
	if err != nil {
		t.Fatalf("esperado autenticação bem sucedida, erro: %v", err)
	}
	if user == nil || user.Email.String() != "ok@test.com" {
		t.Fatalf("usuário incorreto retornado")
	}
}

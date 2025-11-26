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

type memUserRepo2 struct{ users map[uuid.UUID]*entity.User }

func newMemUserRepo2() *memUserRepo2 { return &memUserRepo2{users: make(map[uuid.UUID]*entity.User)} }
func (r *memUserRepo2) Create(ctx context.Context, u *entity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *memUserRepo2) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return r.users[id], nil
}
func (r *memUserRepo2) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	for _, u := range r.users {
		if u.Email.String() == email {
			return u, nil
		}
	}
	return nil, nil
}
func (r *memUserRepo2) Update(ctx context.Context, u *entity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *memUserRepo2) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

var _ userRepo.UserRepository = (*memUserRepo2)(nil)

type memWalletRepo2 struct{ wallets map[uuid.UUID]*entity.Wallet }

func newMemWalletRepo2() *memWalletRepo2 {
	return &memWalletRepo2{wallets: make(map[uuid.UUID]*entity.Wallet)}
}
func (r *memWalletRepo2) Create(ctx context.Context, w *entity.Wallet) error {
	r.wallets[w.UserID] = w
	return nil
}
func (r *memWalletRepo2) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	return r.wallets[userID], nil
}
func (r *memWalletRepo2) FindByAddress(ctx context.Context, address string) (*entity.Wallet, error) {
	for _, w := range r.wallets {
		if w.Address == address {
			return w, nil
		}
	}
	return nil, nil
}
func (r *memWalletRepo2) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	if w := r.wallets[userID]; w != nil {
		w.Balance = balance
	}
	return nil
}

var _ userRepo.WalletRepository = (*memWalletRepo2)(nil)

func TestGetUserWallet_Sucesso(t *testing.T) {
	lg := zap.NewNop()
	bus := events.NewInMemoryBus(lg)
	ur := newMemUserRepo2()
	wr := newMemWalletRepo2()
	uid := uuid.New()
	e, _ := valueobject.NewEmail("w@test.com")
	p, _ := valueobject.HashFromRaw("password123")
	_ = ur.Create(context.Background(), &entity.User{ID: uid, Email: e, Password: p})
	_ = wr.Create(context.Background(), &entity.Wallet{UserID: uid, Address: "A", Balance: 42})
	svc := NewUserService(ur, wr, bus, lg)
	wallet, err := svc.GetUserWallet(context.Background(), uid)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if wallet == nil || wallet.Balance != 42 {
		t.Fatalf("wallet balance incorreto")
	}
}

func TestGetUserWallet_Nil(t *testing.T) {
	lg := zap.NewNop()
	bus := events.NewInMemoryBus(lg)
	ur := newMemUserRepo2()
	wr := newMemWalletRepo2() // nenhuma wallet criada
	uid := uuid.New()
	e2, _ := valueobject.NewEmail("nw@test.com")
	p2, _ := valueobject.HashFromRaw("password123")
	_ = ur.Create(context.Background(), &entity.User{ID: uid, Email: e2, Password: p2})
	svc := NewUserService(ur, wr, bus, lg)
	wallet, err := svc.GetUserWallet(context.Background(), uid)
	if err != nil {
		t.Fatalf("não deveria retornar erro para wallet ausente: %v", err)
	}
	if wallet != nil {
		t.Fatalf("esperado wallet nil quando não existe, obtido %+v", wallet)
	}
}

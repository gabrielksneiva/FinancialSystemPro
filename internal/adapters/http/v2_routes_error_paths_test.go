package http

import (
	"context"
	txnService "financial-system-pro/internal/contexts/transaction/application/service"
	txnEntity "financial-system-pro/internal/contexts/transaction/domain/entity"
	txnRepo "financial-system-pro/internal/contexts/transaction/domain/repository"
	userService "financial-system-pro/internal/contexts/user/application/service"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	userRepo "financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"
	"financial-system-pro/internal/shared/utils"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Repositórios mínimos reutilizados (similar aos do teste principal)
type epUserRepo struct {
	users map[uuid.UUID]*userEntity.User
}
type epWalletRepo struct {
	wallets map[uuid.UUID]*userEntity.Wallet
}
type epTxRepo struct {
	txs map[uuid.UUID]*txnEntity.Transaction
}

func newEpUserRepo() *epUserRepo { return &epUserRepo{users: make(map[uuid.UUID]*userEntity.User)} }
func newEpWalletRepo() *epWalletRepo {
	return &epWalletRepo{wallets: make(map[uuid.UUID]*userEntity.Wallet)}
}
func newEpTxRepo() *epTxRepo { return &epTxRepo{txs: make(map[uuid.UUID]*txnEntity.Transaction)} }

// Implementações interface
func (r *epUserRepo) Create(ctx context.Context, u *userEntity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *epUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	return r.users[id], nil
}
func (r *epUserRepo) FindByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	for _, u := range r.users {
		if strings.EqualFold(string(u.Email), email) {
			return u, nil
		}
	}
	return nil, nil
}
func (r *epUserRepo) Update(ctx context.Context, u *userEntity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *epUserRepo) Delete(ctx context.Context, id uuid.UUID) error { delete(r.users, id); return nil }

var _ userRepo.UserRepository = (*epUserRepo)(nil)

func (r *epWalletRepo) Create(ctx context.Context, w *userEntity.Wallet) error {
	r.wallets[w.UserID] = w
	return nil
}
func (r *epWalletRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*userEntity.Wallet, error) {
	return r.wallets[userID], nil
}
func (r *epWalletRepo) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	for _, w := range r.wallets {
		if w.Address == address {
			return w, nil
		}
	}
	return nil, nil
}
func (r *epWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	if w := r.wallets[userID]; w != nil {
		w.Balance = balance
	}
	return nil
}

var _ userRepo.WalletRepository = (*epWalletRepo)(nil)

func (r *epTxRepo) Create(ctx context.Context, tx *txnEntity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *epTxRepo) FindByID(ctx context.Context, id uuid.UUID) (*txnEntity.Transaction, error) {
	return r.txs[id], nil
}
func (r *epTxRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*txnEntity.Transaction, error) {
	var list []*txnEntity.Transaction
	for _, tx := range r.txs {
		if tx.UserID == userID {
			list = append(list, tx)
		}
	}
	return list, nil
}
func (r *epTxRepo) FindByHash(ctx context.Context, hash string) (*txnEntity.Transaction, error) {
	for _, tx := range r.txs {
		if tx.TransactionHash == hash {
			return tx, nil
		}
	}
	return nil, nil
}
func (r *epTxRepo) Update(ctx context.Context, tx *txnEntity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *epTxRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status txnEntity.TransactionStatus) error {
	if tx := r.txs[id]; tx != nil {
		tx.Status = status
	}
	return nil
}

var _ txnRepo.TransactionRepository = (*epTxRepo)(nil)

// Helper para criar app e retornar token
func setupAppForErrors(t *testing.T) (*fiber.App, string) {
	logger := zap.NewNop()
	bus := events.NewInMemoryBus(logger)
	br := breaker.NewBreakerManager(logger)
	ur := newEpUserRepo()
	wr := newEpWalletRepo()
	tr := newEpTxRepo()
	svcUser := userService.NewUserService(ur, wr, bus, logger)
	svcTxn := txnService.NewTransactionService(tr, ur, wr, bus, br, logger)
	app := fiber.New()
	registerV2DDDRoutes(app, svcUser, svcTxn, logger, br)
	// criar token diretamente para evitar dependências do endpoint de login
	token, _ := utils.CreateJWTToken(map[string]interface{}{"ID": uuid.New().String()})
	return app, token
}

func TestV2Routes_DepositInvalidAmount(t *testing.T) {
	app, token := setupAppForErrors(t)
	req := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader(`{"amount":"-5"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("req err: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("esperado 400 obtido %d", resp.StatusCode)
	}
}

func TestV2Routes_DepositBadBody(t *testing.T) {
	app, token := setupAppForErrors(t)
	req := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader(`not-json`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("req err: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("esperado 400 invalid body obtido %d", resp.StatusCode)
	}
}

func TestV2Routes_WithdrawUnauthorized(t *testing.T) {
	app, _ := setupAppForErrors(t)
	req := httptest.NewRequest("POST", "/v2/transactions/withdraw", strings.NewReader(`{"amount":"5"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("req err: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("esperado 401 obtido %d", resp.StatusCode)
	}
}

func TestV2Routes_GetWalletInvalidID(t *testing.T) {
	app, _ := setupAppForErrors(t)
	req := httptest.NewRequest("GET", "/v2/users/invalid-uuid/wallet", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("req err: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("esperado 400 invalid id obtido %d", resp.StatusCode)
	}
}

func TestV2Routes_WithdrawInvalidAmount(t *testing.T) {
	app, token := setupAppForErrors(t)
	req := httptest.NewRequest("POST", "/v2/transactions/withdraw", strings.NewReader(`{"amount":"-1"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("req err: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("esperado 400 withdraw invalid amount obtido %d", resp.StatusCode)
	}
}

func TestV2Routes_DepositUnauthorized(t *testing.T) {
	app, _ := setupAppForErrors(t)
	req := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader(`{"amount":"10"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("req err: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("esperado 401 deposit unauthorized obtido %d", resp.StatusCode)
	}
}

func TestV2Routes_LoginInvalidCredentials(t *testing.T) {
	// criar app sem usuário correspondente
	logger := zap.NewNop()
	bus := events.NewInMemoryBus(logger)
	br := breaker.NewBreakerManager(logger)
	ur := newEpUserRepo()
	wr := newEpWalletRepo()
	tr := newEpTxRepo()
	// só criar usuário diferente
	svcUser := userService.NewUserService(ur, wr, bus, logger)
	svcTxn := txnService.NewTransactionService(tr, ur, wr, bus, br, logger)
	app := fiber.New()
	registerV2DDDRoutes(app, svcUser, svcTxn, logger, br)
	// tentativa de login com usuário inexistente
	req := httptest.NewRequest("POST", "/v2/auth/login", strings.NewReader(`{"email":"x@y.com","password":"pw"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("login req err: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("esperado 401 login inválido obtido %d", resp.StatusCode)
	}
}

func TestV2Routes_CreateUserConflict(t *testing.T) {
	logger := zap.NewNop()
	bus := events.NewInMemoryBus(logger)
	br := breaker.NewBreakerManager(logger)
	ur := newEpUserRepo()
	wr := newEpWalletRepo()
	tr := newEpTxRepo()
	svcUser := userService.NewUserService(ur, wr, bus, logger)
	svcTxn := txnService.NewTransactionService(tr, ur, wr, bus, br, logger)
	app := fiber.New()
	registerV2DDDRoutes(app, svcUser, svcTxn, logger, br)
	// criar
	req1 := httptest.NewRequest("POST", "/v2/users", strings.NewReader(`{"email":"a@b.com","password":"password"}`))
	req1.Header.Set("Content-Type", "application/json")
	resp1, _ := app.Test(req1)
	if resp1.StatusCode != fiber.StatusCreated {
		t.Fatalf("esperado 201 primeira criação obtido %d", resp1.StatusCode)
	}
	// duplicar
	req2 := httptest.NewRequest("POST", "/v2/users", strings.NewReader(`{"email":"a@b.com","password":"password"}`))
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := app.Test(req2)
	if resp2.StatusCode != fiber.StatusConflict {
		t.Fatalf("esperado 409 conflito obtido %d", resp2.StatusCode)
	}
}

package http

import (
	"context"
	"encoding/json"
	txnService "financial-system-pro/internal/contexts/transaction/application/service"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	txnRepo "financial-system-pro/internal/contexts/transaction/domain/repository"
	userService "financial-system-pro/internal/contexts/user/application/service"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	userRepo "financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// In-memory implementations
type inMemoryUserRepo struct {
	users map[uuid.UUID]*userEntity.User
}
type inMemoryWalletRepo struct {
	wallets map[uuid.UUID]*userEntity.Wallet
}
type inMemoryTxRepo struct {
	txs map[uuid.UUID]*entity.Transaction
}

func newInMemoryUserRepo() *inMemoryUserRepo {
	return &inMemoryUserRepo{users: make(map[uuid.UUID]*userEntity.User)}
}
func newInMemoryWalletRepo() *inMemoryWalletRepo {
	return &inMemoryWalletRepo{wallets: make(map[uuid.UUID]*userEntity.Wallet)}
}
func newInMemoryTxRepo() *inMemoryTxRepo {
	return &inMemoryTxRepo{txs: make(map[uuid.UUID]*entity.Transaction)}
}

// UserRepository
func (r *inMemoryUserRepo) Create(ctx context.Context, user *userEntity.User) error {
	r.users[user.ID] = user
	return nil
}
func (r *inMemoryUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	return r.users[id], nil
}
func (r *inMemoryUserRepo) FindByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, nil
}
func (r *inMemoryUserRepo) Update(ctx context.Context, user *userEntity.User) error {
	r.users[user.ID] = user
	return nil
}
func (r *inMemoryUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

// WalletRepository
func (r *inMemoryWalletRepo) Create(ctx context.Context, w *userEntity.Wallet) error {
	r.wallets[w.UserID] = w
	return nil
}
func (r *inMemoryWalletRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*userEntity.Wallet, error) {
	return r.wallets[userID], nil
}
func (r *inMemoryWalletRepo) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	for _, w := range r.wallets {
		if w.Address == address {
			return w, nil
		}
	}
	return nil, nil
}
func (r *inMemoryWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	if w := r.wallets[userID]; w != nil {
		w.Balance = balance
		return nil
	}
	return nil
}

// TransactionRepository
func (r *inMemoryTxRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *inMemoryTxRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	return r.txs[id], nil
}
func (r *inMemoryTxRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	var list []*entity.Transaction
	for _, tx := range r.txs {
		if tx.UserID == userID {
			list = append(list, tx)
		}
	}
	return list, nil
}
func (r *inMemoryTxRepo) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	for _, tx := range r.txs {
		if tx.TransactionHash == hash {
			return tx, nil
		}
	}
	return nil, nil
}
func (r *inMemoryTxRepo) Update(ctx context.Context, tx *entity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *inMemoryTxRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TransactionStatus) error {
	if tx := r.txs[id]; tx != nil {
		tx.Status = status
		return nil
	}
	return nil
}

// Compile-time checks
var _ userRepo.UserRepository = (*inMemoryUserRepo)(nil)
var _ userRepo.WalletRepository = (*inMemoryWalletRepo)(nil)
var _ txnRepo.TransactionRepository = (*inMemoryTxRepo)(nil)

func TestV2Routes_UserLifecycleAndTransactions(t *testing.T) {
	// Setup infra
	logger := zap.NewNop()
	eventBus := events.NewInMemoryBus(logger)
	breakerManager := breaker.NewBreakerManager(logger)

	// Repos
	ur := newInMemoryUserRepo()
	wr := newInMemoryWalletRepo()
	tr := newInMemoryTxRepo()

	// Services
	dddUserSvc := userService.NewUserService(ur, wr, eventBus, logger)
	dddTxnSvc := txnService.NewTransactionService(tr, ur, wr, eventBus, breakerManager, logger)

	// App
	app := fiber.New()
	registerV2DDDRoutes(app, dddUserSvc, dddTxnSvc, logger, breakerManager)

	// 1. Create user
	req := httptest.NewRequest("POST", "/v2/users", strings.NewReader(`{"email":"test@example.com","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("create user request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 got %d", resp.StatusCode)
	}
	var userResp map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&userResp)
	userIDStr, ok := userResp["id"].(string)
	if !ok {
		t.Fatalf("missing user id in response")
	}
	userID, _ := uuid.Parse(userIDStr)

	// Seed wallet manually (DDD user service doesn't auto-create)
	_ = wr.Create(context.Background(), &userEntity.Wallet{UserID: userID, Address: "WADDR", Balance: 0})

	// 2. Login
	loginReq := httptest.NewRequest("POST", "/v2/auth/login", strings.NewReader(`{"email":"test@example.com","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	if loginResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 login got %d", loginResp.StatusCode)
	}
	var loginData map[string]interface{}
	_ = json.NewDecoder(loginResp.Body).Decode(&loginData)
	token, ok := loginData["token"].(string)
	if !ok || token == "" {
		t.Fatalf("missing token")
	}

	// Helper to authorized request
	authHeader := "Bearer " + token

	// 3. Deposit
	depReq := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader(`{"amount":"10"}`))
	depReq.Header.Set("Authorization", authHeader)
	depReq.Header.Set("Content-Type", "application/json")
	depResp, err := app.Test(depReq)
	if err != nil {
		t.Fatalf("deposit request failed: %v", err)
	}
	if depResp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("expected 202 deposit got %d", depResp.StatusCode)
	}

	// 4. Withdraw (valid - amount <= balance after deposit)
	wReq := httptest.NewRequest("POST", "/v2/transactions/withdraw", strings.NewReader(`{"amount":"5"}`))
	wReq.Header.Set("Authorization", authHeader)
	wReq.Header.Set("Content-Type", "application/json")
	wResp, err := app.Test(wReq)
	if err != nil {
		t.Fatalf("withdraw request failed: %v", err)
	}
	if wResp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("expected 202 withdraw got %d", wResp.StatusCode)
	}

	// 5. History
	hReq := httptest.NewRequest("GET", "/v2/transactions/history", nil)
	hReq.Header.Set("Authorization", authHeader)
	hResp, err := app.Test(hReq)
	if err != nil {
		t.Fatalf("history request failed: %v", err)
	}
	if hResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 history got %d", hResp.StatusCode)
	}
	var hData map[string]interface{}
	_ = json.NewDecoder(hResp.Body).Decode(&hData)
	if _, ok := hData["transactions"].([]interface{}); !ok {
		t.Fatalf("missing transactions array")
	}

	// 6. Wallet info
	walletReq := httptest.NewRequest("GET", "/v2/users/"+userIDStr+"/wallet", nil)
	walletResp, err := app.Test(walletReq)
	if err != nil {
		t.Fatalf("wallet request failed: %v", err)
	}
	if walletResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 wallet got %d", walletResp.StatusCode)
	}
}

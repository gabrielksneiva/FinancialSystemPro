package http

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	txnService "financial-system-pro/internal/contexts/transaction/application/service"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	txnRepo "financial-system-pro/internal/contexts/transaction/domain/repository"
	userService "financial-system-pro/internal/contexts/user/application/service"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	userRepo "financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"
	"financial-system-pro/internal/shared/utils"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Reuse in-memory repos defined in other *_test.go (same package) or define minimal failing repo here.

// failingWalletRepo always returns error for FindByUserID to trigger circuit breaker.
type failingWalletRepo struct{}

func (f *failingWalletRepo) Create(ctx context.Context, w *userEntity.Wallet) error { return nil }
func (f *failingWalletRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*userEntity.Wallet, error) {
	return nil, errors.New("wallet fetch failure")
}
func (f *failingWalletRepo) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	return nil, errors.New("wallet fetch failure")
}
func (f *failingWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	return errors.New("update not supported")
}

// Ensure interface compliance
var _ userRepo.WalletRepository = (*failingWalletRepo)(nil)

// simpleTxRepo minimal in-memory tx repo for circuit breaker test
type simpleTxRepo struct {
	txs map[uuid.UUID]*entity.Transaction
}

func newSimpleTxRepo() *simpleTxRepo {
	return &simpleTxRepo{txs: make(map[uuid.UUID]*entity.Transaction)}
}
func (r *simpleTxRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *simpleTxRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	return r.txs[id], nil
}
func (r *simpleTxRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	var list []*entity.Transaction
	for _, tx := range r.txs {
		if tx.UserID == userID {
			list = append(list, tx)
		}
	}
	return list, nil
}
func (r *simpleTxRepo) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	for _, tx := range r.txs {
		if tx.TransactionHash == hash {
			return tx, nil
		}
	}
	return nil, nil
}
func (r *simpleTxRepo) Update(ctx context.Context, tx *entity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *simpleTxRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TransactionStatus) error {
	if tx := r.txs[id]; tx != nil {
		tx.Status = status
	}
	return nil
}

var _ txnRepo.TransactionRepository = (*simpleTxRepo)(nil)

// helperCreateAuthedApp sets up app with user + wallet + JWT token
func helperCreateAuthedApp(t *testing.T, walletBalance float64) (*fiber.App, string, uuid.UUID, *inMemoryWalletRepo) {
	t.Helper()
	t.Setenv("SECRET_KEY", "test-secret")
	t.Setenv("EXPIRATION_TIME", "3600")

	logger := zap.NewNop()
	eventBus := events.NewInMemoryBus(logger)
	breakerManager := breaker.NewBreakerManager(logger)

	ur := newInMemoryUserRepo()
	wr := newInMemoryWalletRepo()
	tr := newInMemoryTxRepo()

	userSvc := userService.NewUserService(ur, wr, eventBus, logger)
	txnSvc := txnService.NewTransactionService(tr, ur, wr, eventBus, breakerManager, logger)

	app := fiber.New()
	registerV2DDDRoutes(app, userSvc, txnSvc, logger, breakerManager)

	uid := uuid.New()
	_ = ur.Create(context.Background(), &userEntity.User{ID: uid, Email: "jwt@test.com", Password: "hashed"})
	_ = wr.Create(context.Background(), &userEntity.Wallet{UserID: uid, Address: "ADDR", Balance: walletBalance})

	token, err := utils.CreateJWTToken(map[string]interface{}{"ID": uid.String()})
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}
	return app, token, uid, wr
}

func TestV2JWT_DepositInvalidAmounts(t *testing.T) {
	app, token, _, _ := helperCreateAuthedApp(t, 0)

	for _, amt := range []string{"0", "-1"} {
		req := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader("{\"amount\":\""+amt+"\"}"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("amount %s expected 400 got %d", amt, resp.StatusCode)
		}
	}
}

func TestV2JWT_WithdrawInsufficientBalance(t *testing.T) {
	app, token, uid, wr := helperCreateAuthedApp(t, 5) // balance 5
	req := httptest.NewRequest("POST", "/v2/transactions/withdraw", strings.NewReader("{\"amount\":\"10\"}"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("withdraw failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
	// balance should remain 5
	wallet, _ := wr.FindByUserID(context.Background(), uid)
	if wallet.Balance != 5 {
		t.Errorf("wallet balance changed unexpectedly: %v", wallet.Balance)
	}
}

func TestV2JWT_DepositValidUpdatesBalance(t *testing.T) {
	app, token, uid, wr := helperCreateAuthedApp(t, 0)
	req := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader("{\"amount\":\"10\"}"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("deposit failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("expected 202 got %d", resp.StatusCode)
	}
	w, _ := wr.FindByUserID(context.Background(), uid)
	if w.Balance < 10 {
		t.Errorf("expected balance >=10 got %v", w.Balance)
	}
}

func TestV2JWT_InvalidToken(t *testing.T) {
	app, _, _, _ := helperCreateAuthedApp(t, 0)
	req := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader("{\"amount\":\"10\"}"))
	req.Header.Set("Authorization", "Bearer invalidtoken")
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", resp.StatusCode)
	}
}

func TestV2JWT_MissingToken(t *testing.T) {
	app, _, _, _ := helperCreateAuthedApp(t, 0)
	req := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader("{\"amount\":\"10\"}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", resp.StatusCode)
	}
}

func TestV2CircuitBreaker_OpenAfterFailures(t *testing.T) {
	t.Setenv("SECRET_KEY", "test-secret")
	logger := zap.NewNop()
	eventBus := events.NewInMemoryBus(logger)
	breakerManager := breaker.NewBreakerManager(logger)

	// user repo & failing wallet repo
	ur := newInMemoryUserRepo()
	failingWR := &failingWalletRepo{}
	tr := newSimpleTxRepo()
	userSvc := userService.NewUserService(ur, failingWR, eventBus, logger)
	txnSvc := txnService.NewTransactionService(tr, ur, failingWR, eventBus, breakerManager, logger)
	app := fiber.New()
	registerV2DDDRoutes(app, userSvc, txnSvc, logger, breakerManager)

	uid := uuid.New()
	_ = ur.Create(context.Background(), &userEntity.User{ID: uid, Email: "breaker@test.com", Password: "hashed"})
	token, _ := utils.CreateJWTToken(map[string]interface{}{"ID": uid.String()})

	// Perform 6 failing deposit attempts to trip breaker (>5 consecutive failures)
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("POST", "/v2/transactions/deposit", strings.NewReader("{\"amount\":\"1\"}"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		if resp.StatusCode != fiber.StatusInternalServerError {
			t.Errorf("attempt %d expected 500 got %d", i, resp.StatusCode)
		}
	}

	breakerState := breakerManager.GetBreaker(breaker.BreakerTransactionToUser).State().String()
	if breakerState != "open" {
		t.Fatalf("expected breaker open got %s", breakerState)
	}
}

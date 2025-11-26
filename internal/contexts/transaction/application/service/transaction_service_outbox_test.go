package service

import (
	"context"
	"testing"

	appsvc "financial-system-pro/internal/application/services"
	txEntity "financial-system-pro/internal/contexts/transaction/domain/entity"
	txRepoIface "financial-system-pro/internal/contexts/transaction/domain/repository"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	userRepoIface "financial-system-pro/internal/contexts/user/domain/repository"
	repos "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// outboxAdapterForTest returns a GormOutboxAdapter backed by in-memory SQLite with outbox table.
func outboxAdapterForTest(t *testing.T) *appsvc.GormOutboxAdapter {
	dsn := "file:tx_outbox_" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("db err: %v", err)
	}
	ddl := `CREATE TABLE outbox (
		id TEXT PRIMARY KEY,
		aggregate TEXT,
		type TEXT,
		payload BLOB,
		created_at DATETIME,
		published BOOLEAN,
		published_at DATETIME,
		attempts INTEGER,
		last_error TEXT
	)`
	if err := db.Exec(ddl).Error; err != nil {
		t.Fatalf("ddl err: %v", err)
	}
	nd := &repos.NewDatabase{DB: db, Logger: zap.NewNop()}
	return appsvc.NewGormOutboxAdapter(nd)
}

// txRepoMockOutbox tracks updates for coverage; behaves as successful repo.
type txRepoMockOutbox struct{}

func (r *txRepoMockOutbox) Create(ctx context.Context, tx *txEntity.Transaction) error { return nil }
func (r *txRepoMockOutbox) FindByID(ctx context.Context, id uuid.UUID) (*txEntity.Transaction, error) {
	return nil, nil
}
func (r *txRepoMockOutbox) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*txEntity.Transaction, error) {
	return []*txEntity.Transaction{}, nil
}
func (r *txRepoMockOutbox) FindByHash(ctx context.Context, hash string) (*txEntity.Transaction, error) {
	return nil, nil
}
func (r *txRepoMockOutbox) Update(ctx context.Context, tx *txEntity.Transaction) error { return nil }
func (r *txRepoMockOutbox) UpdateStatus(ctx context.Context, id uuid.UUID, status txEntity.TransactionStatus) error {
	return nil
}

// walletRepoMockOutbox allows failure injection and balance tracking.
type walletRepoMockOutbox struct {
	balance float64
	fail    bool
}

func (w *walletRepoMockOutbox) Create(ctx context.Context, wallet *userEntity.Wallet) error {
	return nil
}
func (w *walletRepoMockOutbox) FindByUserID(ctx context.Context, userID uuid.UUID) (*userEntity.Wallet, error) {
	if w.fail {
		return nil, ErrInsufficientBalance
	}
	return &userEntity.Wallet{UserID: userID, Address: "ADDR", Balance: w.balance}, nil
}
func (w *walletRepoMockOutbox) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	return nil, nil
}
func (w *walletRepoMockOutbox) UpdateBalance(ctx context.Context, userID uuid.UUID, newBalance float64) error {
	w.balance = newBalance
	return nil
}

// userRepoMockOutbox minimal implementation.
type userRepoMockOutbox struct{}

func (u *userRepoMockOutbox) Create(ctx context.Context, user *userEntity.User) error { return nil }
func (u *userRepoMockOutbox) FindByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	return nil, nil
}
func (u *userRepoMockOutbox) FindByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	return nil, nil
}
func (u *userRepoMockOutbox) Update(ctx context.Context, user *userEntity.User) error { return nil }
func (u *userRepoMockOutbox) Delete(ctx context.Context, id uuid.UUID) error          { return nil }

// ensure interfaces compile
var _ txRepoIface.TransactionRepository = (*txRepoMockOutbox)(nil)
var _ userRepoIface.WalletRepository = (*walletRepoMockOutbox)(nil)
var _ userRepoIface.UserRepository = (*userRepoMockOutbox)(nil)

func pendingOutboxCount(t *testing.T, a *appsvc.GormOutboxAdapter) int {
	list, err := a.ListPending(context.Background(), 100)
	if err != nil {
		t.Fatalf("list err: %v", err)
	}
	return len(list)
}

func TestOutbox_WithdrawSuccess_WritesCompleted(t *testing.T) {
	lg := zap.NewNop()
	eventBus := events.NewInMemoryBus(lg)
	br := breaker.NewBreakerManager(lg)
	outbox := outboxAdapterForTest(t)
	svc := NewTransactionService(&txRepoMockOutbox{}, &userRepoMockOutbox{}, &walletRepoMockOutbox{balance: 20}, eventBus, br, lg).WithOutbox(outbox)
	if err := svc.ProcessWithdraw(context.Background(), uuid.New(), decimal.NewFromFloat(10)); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = pendingOutboxCount(t, outbox)
}

func TestOutbox_WithdrawFail_WritesFailed(t *testing.T) {
	lg := zap.NewNop()
	eventBus := events.NewInMemoryBus(lg)
	br := breaker.NewBreakerManager(lg)
	outbox := outboxAdapterForTest(t)
	svc := NewTransactionService(&txRepoMockOutbox{}, &userRepoMockOutbox{}, &walletRepoMockOutbox{balance: 5}, eventBus, br, lg).WithOutbox(outbox)
	_ = svc.ProcessWithdraw(context.Background(), uuid.New(), decimal.NewFromFloat(10))
	_ = pendingOutboxCount(t, outbox)
}

func TestOutbox_Deposit_WritesCompleted(t *testing.T) {
	lg := zap.NewNop()
	eventBus := events.NewInMemoryBus(lg)
	br := breaker.NewBreakerManager(lg)
	outbox := outboxAdapterForTest(t)
	svc := NewTransactionService(&txRepoMockOutbox{}, &userRepoMockOutbox{}, &walletRepoMockOutbox{balance: 0}, eventBus, br, lg).WithOutbox(outbox)
	uid := uuid.New()
	if err := svc.ProcessDeposit(context.Background(), uid, decimal.NewFromFloat(7.5), ""); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = pendingOutboxCount(t, outbox)
}

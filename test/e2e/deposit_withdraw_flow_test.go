package e2e_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"financial-system-pro/internal/contexts/transaction/application/service"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// In-memory transaction repo
type memTxnRepo struct {
	txs map[uuid.UUID]*entity.Transaction
}

func (r *memTxnRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *memTxnRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	return r.txs[id], nil
}
func (r *memTxnRepo) FindByUserID(ctx context.Context, uid uuid.UUID) ([]*entity.Transaction, error) {
	var list []*entity.Transaction
	for _, tx := range r.txs {
		if tx.UserID == uid {
			list = append(list, tx)
		}
	}
	return list, nil
}
func (r *memTxnRepo) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	for _, tx := range r.txs {
		if tx.TransactionHash == hash {
			return tx, nil
		}
	}
	return nil, nil
}
func (r *memTxnRepo) Update(ctx context.Context, tx *entity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *memTxnRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TransactionStatus) error {
	if tx := r.txs[id]; tx != nil {
		tx.Status = status
	}
	return nil
}

// In-memory user repo
type memUserRepo struct {
	users map[uuid.UUID]*userEntity.User
}

func (r *memUserRepo) Create(ctx context.Context, u *userEntity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *memUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	return r.users[id], nil
}
func (r *memUserRepo) FindByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}
func (r *memUserRepo) Update(ctx context.Context, u *userEntity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *memUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

// In-memory wallet repo
type memWalletRepo struct {
	wallets map[uuid.UUID]*userEntity.Wallet
}

func (r *memWalletRepo) Create(ctx context.Context, w *userEntity.Wallet) error {
	r.wallets[w.UserID] = w
	return nil
}
func (r *memWalletRepo) FindByUserID(ctx context.Context, uid uuid.UUID) (*userEntity.Wallet, error) {
	return r.wallets[uid], nil
}
func (r *memWalletRepo) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	for _, w := range r.wallets {
		if w.Address == address {
			return w, nil
		}
	}
	return nil, nil
}
func (r *memWalletRepo) UpdateBalance(ctx context.Context, uid uuid.UUID, balance float64) error {
	if w := r.wallets[uid]; w != nil {
		w.Balance = balance
	}
	return nil
}

func TestAcceptance_DepositAndWithdrawFlow(t *testing.T) {
	logger := zap.NewNop()
	eventBus := events.NewInMemoryBus(logger)
	br := breaker.NewBreakerManager(logger)

	txr := &memTxnRepo{txs: make(map[uuid.UUID]*entity.Transaction)}
	ur := &memUserRepo{users: make(map[uuid.UUID]*userEntity.User)}
	wr := &memWalletRepo{wallets: make(map[uuid.UUID]*userEntity.Wallet)}

	uid := uuid.New()
	_ = ur.Create(context.Background(), &userEntity.User{ID: uid, Email: "acc@test.com", Password: "hash"})
	_ = wr.Create(context.Background(), &userEntity.Wallet{UserID: uid, Address: "ADDR", Balance: 0})

	svc := service.NewTransactionService(txr, ur, wr, eventBus, br, logger)

	app := fiber.New()
	app.Post("/deposit", func(c *fiber.Ctx) error {
		amount := decimal.NewFromInt(10)
		_ = svc.ProcessDeposit(context.Background(), uid, amount, "")
		return c.SendStatus(202)
	})
	app.Post("/withdraw", func(c *fiber.Ctx) error {
		amount := decimal.NewFromInt(5)
		_ = svc.ProcessWithdraw(context.Background(), uid, amount)
		return c.SendStatus(200)
	})

	depReq := httptest.NewRequest("POST", "/deposit", nil)
	depResp, err := app.Test(depReq)
	require.NoError(t, err)
	assert.Equal(t, 202, depResp.StatusCode)
	_ = depResp.Body.Close()
	wReq := httptest.NewRequest("POST", "/withdraw", nil)
	wResp, err := app.Test(wReq)
	require.NoError(t, err)
	assert.Equal(t, 200, wResp.StatusCode)
	_ = wResp.Body.Close()

	wallet, _ := wr.FindByUserID(context.Background(), uid)
	assert.Equal(t, 5.0, wallet.Balance)

	time.Sleep(30 * time.Millisecond)
}

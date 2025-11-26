package service

import (
	"context"
	"testing"

	"financial-system-pro/internal/contexts/transaction/domain/entity"
	txnRepo "financial-system-pro/internal/contexts/transaction/domain/repository"
	userEntity "financial-system-pro/internal/contexts/user/domain/entity"
	userRepo "financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/shared/breaker"
	"financial-system-pro/internal/shared/events"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Repositórios in-memory para testes
type memTxnRepo struct {
	txs map[uuid.UUID]*entity.Transaction
}

func newMemTxnRepo() *memTxnRepo { return &memTxnRepo{txs: make(map[uuid.UUID]*entity.Transaction)} }
func (r *memTxnRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	r.txs[tx.ID] = tx
	return nil
}
func (r *memTxnRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	return r.txs[id], nil
}
func (r *memTxnRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	var l []*entity.Transaction
	for _, tx := range r.txs {
		if tx.UserID == userID {
			l = append(l, tx)
		}
	}
	return l, nil
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

var _ txnRepo.TransactionRepository = (*memTxnRepo)(nil)

type memUserRepo struct {
	users map[uuid.UUID]*userEntity.User
}

func newMemUserRepo() *memUserRepo { return &memUserRepo{users: make(map[uuid.UUID]*userEntity.User)} }
func (r *memUserRepo) Create(ctx context.Context, u *userEntity.User) error {
	r.users[u.ID] = u
	return nil
}
func (r *memUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	return r.users[id], nil
}
func (r *memUserRepo) FindByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	for _, u := range r.users {
		if u.Email.String() == email {
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

var _ userRepo.UserRepository = (*memUserRepo)(nil)

type memWalletRepo struct {
	wallets map[uuid.UUID]*userEntity.Wallet
}

func newMemWalletRepo() *memWalletRepo {
	return &memWalletRepo{wallets: make(map[uuid.UUID]*userEntity.Wallet)}
}
func (r *memWalletRepo) Create(ctx context.Context, w *userEntity.Wallet) error {
	r.wallets[w.UserID] = w
	return nil
}
func (r *memWalletRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*userEntity.Wallet, error) {
	return r.wallets[userID], nil
}
func (r *memWalletRepo) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	for _, w := range r.wallets {
		if w.Address == address {
			return w, nil
		}
	}
	return nil, nil
}
func (r *memWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	if w := r.wallets[userID]; w != nil {
		w.Balance = balance
	}
	return nil
}

var _ userRepo.WalletRepository = (*memWalletRepo)(nil)

// Helper de setup
func setupService(t *testing.T, balance float64) (*TransactionService, *memTxnRepo, *memWalletRepo, uuid.UUID) {
	t.Helper()
	lg := zap.NewNop()
	eventBus := events.NewInMemoryBus(lg)
	br := breaker.NewBreakerManager(lg)
	txr := newMemTxnRepo()
	ur := newMemUserRepo()
	wr := newMemWalletRepo()
	uid := uuid.New()
	_ = ur.Create(context.Background(), &userEntity.User{ID: uid, Email: "t@t.com", Password: "hash"})
	_ = wr.Create(context.Background(), &userEntity.Wallet{UserID: uid, Address: "ADDR", Balance: balance})
	svc := NewTransactionService(txr, ur, wr, eventBus, br, lg)
	return svc, txr, wr, uid
}

func TestProcessDeposit_Sucesso(t *testing.T) {
	svc, txr, wr, uid := setupService(t, 0)
	amt, _ := decimal.NewFromString("10")
	if err := svc.ProcessDeposit(context.Background(), uid, amt, ""); err != nil {
		t.Fatalf("erro deposito: %v", err)
	}
	w, _ := wr.FindByUserID(context.Background(), uid)
	if w.Balance < 10 {
		t.Fatalf("esperado saldo >=10 obtido %v", w.Balance)
	}
	// verifica transação criada
	var found *entity.Transaction
	for _, tx := range txr.txs {
		if tx.UserID == uid && tx.Type == entity.TransactionTypeDeposit {
			found = tx
		}
	}
	if found == nil || found.Status != entity.TransactionStatusCompleted {
		t.Fatalf("transação não concluída corretamente")
	}
}

func TestProcessWithdraw_InsufficientBalance(t *testing.T) {
	svc, _, _, uid := setupService(t, 5)
	amt, _ := decimal.NewFromString("10")
	err := svc.ProcessWithdraw(context.Background(), uid, amt)
	if err == nil || err != ErrInsufficientBalance {
		t.Fatalf("esperado ErrInsufficientBalance, obtido %v", err)
	}
}

func TestProcessWithdraw_Sucesso(t *testing.T) {
	svc, txr, wr, uid := setupService(t, 20)
	amt, _ := decimal.NewFromString("5")
	if err := svc.ProcessWithdraw(context.Background(), uid, amt); err != nil {
		t.Fatalf("erro withdraw: %v", err)
	}
	w, _ := wr.FindByUserID(context.Background(), uid)
	if w.Balance != 15 {
		t.Fatalf("saldo esperado 15 obtido %v", w.Balance)
	}
	var found *entity.Transaction
	for _, tx := range txr.txs {
		if tx.UserID == uid && tx.Type == entity.TransactionTypeWithdraw {
			found = tx
		}
	}
	if found == nil || found.Status != entity.TransactionStatusCompleted {
		t.Fatalf("transação withdraw não concluída")
	}
}

func TestGetTransactionHistory(t *testing.T) {
	svc, txr, _, uid := setupService(t, 0)
	a1, _ := decimal.NewFromString("3")
	a2, _ := decimal.NewFromString("4")
	_ = svc.ProcessDeposit(context.Background(), uid, a1, "")
	_ = svc.ProcessDeposit(context.Background(), uid, a2, "")
	list, err := svc.GetTransactionHistory(context.Background(), uid)
	if err != nil {
		t.Fatalf("erro history: %v", err)
	}
	if len(list) < 2 {
		t.Fatalf("esperado >=2 transações, obtido %d", len(list))
	}
	if len(txr.txs) < 2 {
		t.Fatalf("repo deveria conter >=2 transações")
	}
}

// Repositório de wallet que força erro para testar falha / breaker
type failingWalletRepo struct{}

func (failingWalletRepo) Create(ctx context.Context, w *userEntity.Wallet) error { return nil }
func (failingWalletRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*userEntity.Wallet, error) {
	return nil, ErrWalletNotFound
}
func (failingWalletRepo) FindByAddress(ctx context.Context, address string) (*userEntity.Wallet, error) {
	return nil, ErrWalletNotFound
}
func (failingWalletRepo) UpdateBalance(ctx context.Context, userID uuid.UUID, balance float64) error {
	return ErrWalletNotFound
}

var _ userRepo.WalletRepository = (*failingWalletRepo)(nil)

func TestProcessDeposit_FalhaWallet(t *testing.T) {
	lg := zap.NewNop()
	eventBus := events.NewInMemoryBus(lg)
	br := breaker.NewBreakerManager(lg)
	txr := newMemTxnRepo()
	ur := newMemUserRepo()
	uid := uuid.New()
	_ = ur.Create(context.Background(), &userEntity.User{ID: uid, Email: "f@t.com", Password: "hash"})
	wr := failingWalletRepo{}
	svc := NewTransactionService(txr, ur, wr, eventBus, br, lg)
	amt, _ := decimal.NewFromString("2")
	err := svc.ProcessDeposit(context.Background(), uid, amt, "")
	if err == nil {
		t.Fatalf("esperado erro de wallet")
	}
	// transação deve estar failed
	var failed *entity.Transaction
	for _, tx := range txr.txs {
		failed = tx
	}
	if failed == nil || failed.Status != entity.TransactionStatusFailed {
		t.Fatalf("esperado status failed")
	}
}

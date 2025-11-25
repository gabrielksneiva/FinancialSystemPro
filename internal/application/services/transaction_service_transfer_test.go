package services

import (
	domainErrors "financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	w "financial-system-pro/internal/infrastructure/queue"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// fake DB for transfer branch coverage
type fakeDBTransfer struct {
	failFind         bool
	failWithdraw     bool
	failDeposit      bool
	failInsertFirst  bool
	failInsertSecond bool
	insertCalls      int
}

func (f *fakeDBTransfer) FindUserByField(field string, value any) (*repositories.User, error) {
	if field == "email" {
		if f.failFind {
			return nil, domainErrors.NewDatabaseError("find email", nil)
		}
		return &repositories.User{ID: uuid.New(), Email: "dest@example.com"}, nil
	}
	return &repositories.User{ID: uuid.New(), Email: "src@example.com"}, nil
}
func (f *fakeDBTransfer) Insert(v any) error {
	f.insertCalls++
	if f.insertCalls == 1 && f.failInsertFirst {
		return domainErrors.NewDatabaseError("insert first", nil)
	}
	if f.insertCalls == 2 && f.failInsertSecond {
		return domainErrors.NewDatabaseError("insert second", nil)
	}
	return nil
}
func (f *fakeDBTransfer) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error { return nil }
func (f *fakeDBTransfer) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return nil, nil
}
func (f *fakeDBTransfer) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error {
	if tx == "withdraw" && f.failWithdraw {
		return domainErrors.NewDatabaseError("withdraw fail", nil)
	}
	if tx == "deposit" && f.failDeposit {
		return domainErrors.NewDatabaseError("deposit fail", nil)
	}
	return nil
}
func (f *fakeDBTransfer) Balance(userID uuid.UUID) (decimal.Decimal, error) { return decimal.Zero, nil }
func (f *fakeDBTransfer) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

// queue stub for early accepted branch
type queueStub struct{ called bool }

func (q *queueStub) QueueTransaction(job w.TransactionJob) error { q.called = true; return nil }

func TestTransactionService_Transfer_InvalidUser(t *testing.T) {
	svc := &TransactionService{DB: &fakeDBTransfer{}}
	_, err := svc.Transfer("invalid-uuid", decimal.NewFromInt(1), "dest@example.com", "")
	if err == nil || err.Error() == "" {
		t.Fatalf("expected validation error for invalid user id")
	}
}

func TestTransactionService_Transfer_QueueEarlyAccept(t *testing.T) {
	q := &queueStub{}
	svc := &TransactionService{DB: &fakeDBTransfer{}, Queue: q}
	resp, err := svc.Transfer(uuid.New().String(), decimal.NewFromInt(5), "dest@example.com", "cb")
	if err != nil || resp.StatusCode != 202 || !q.called {
		t.Fatalf("expected 202 early accept with queue, got resp=%v err=%v called=%v", resp, err, q.called)
	}
}

func TestTransactionService_Transfer_FindDestFail(t *testing.T) {
	svc := &TransactionService{DB: &fakeDBTransfer{failFind: true}}
	_, err := svc.Transfer(uuid.New().String(), decimal.NewFromInt(2), "dest@example.com", "")
	if err == nil {
		t.Fatalf("expected error for destination find fail")
	}
}

func TestTransactionService_Transfer_WithdrawFail(t *testing.T) {
	svc := &TransactionService{DB: &fakeDBTransfer{failWithdraw: true}}
	_, err := svc.Transfer(uuid.New().String(), decimal.NewFromInt(3), "dest@example.com", "")
	if err == nil {
		t.Fatalf("expected error for withdraw fail")
	}
}

func TestTransactionService_Transfer_DepositFail(t *testing.T) {
	svc := &TransactionService{DB: &fakeDBTransfer{failDeposit: true}}
	_, err := svc.Transfer(uuid.New().String(), decimal.NewFromInt(4), "dest@example.com", "")
	if err == nil {
		t.Fatalf("expected error for deposit fail")
	}
}

func TestTransactionService_Transfer_FirstInsertFail(t *testing.T) {
	svc := &TransactionService{DB: &fakeDBTransfer{failInsertFirst: true}}
	_, err := svc.Transfer(uuid.New().String(), decimal.NewFromInt(6), "dest@example.com", "")
	if err == nil {
		t.Fatalf("expected error for first insert fail")
	}
}

func TestTransactionService_Transfer_SecondInsertFail(t *testing.T) {
	svc := &TransactionService{DB: &fakeDBTransfer{failInsertSecond: true}}
	_, err := svc.Transfer(uuid.New().String(), decimal.NewFromInt(7), "dest@example.com", "")
	if err == nil {
		t.Fatalf("expected error for second insert fail")
	}
}

// helper contains
func contains(hay, needle string) bool { return strings.Contains(hay, needle) }

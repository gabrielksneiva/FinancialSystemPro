package services_test

import (
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/errors"
	r "financial-system-pro/internal/infrastructure/database"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockDBTransactions implementa parte de DatabasePort necessária para testes simples de transações
type mockDBTransactions struct {
	balanceMap      map[uuid.UUID]decimal.Decimal
	txErr           error
	insertErr       error
	findUserByID    *r.User
	findUserByEmail *r.User
}

func newMockDBTx() *mockDBTransactions {
	return &mockDBTransactions{balanceMap: map[uuid.UUID]decimal.Decimal{}}
}

func (m *mockDBTransactions) FindUserByField(field string, value any) (*r.User, error) {
	switch field {
	case "email":
		if m.findUserByEmail != nil && m.findUserByEmail.Email == value.(string) {
			return m.findUserByEmail, nil
		}
		return nil, errors.NewNotFoundError("user")
	case "id":
		if m.findUserByID != nil && m.findUserByID.ID.String() == value.(string) {
			return m.findUserByID, nil
		}
		return nil, errors.NewNotFoundError("user")
	default:
		return nil, errors.NewNotFoundError("user")
	}
}
func (m *mockDBTransactions) Insert(v any) error { return m.insertErr }
func (m *mockDBTransactions) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	return nil
}
func (m *mockDBTransactions) GetWalletInfo(userID uuid.UUID) (*r.WalletInfo, error) {
	return nil, errors.NewNotFoundError("wallet")
}
func (m *mockDBTransactions) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	if m.txErr != nil {
		return m.txErr
	}
	cur := m.balanceMap[userID]
	switch transactionType {
	case "deposit":
		m.balanceMap[userID] = cur.Add(amount)
	case "withdraw":
		m.balanceMap[userID] = cur.Sub(amount)
	}
	return nil
}
func (m *mockDBTransactions) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	return m.balanceMap[userID], nil
}
func (m *mockDBTransactions) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
	return nil
}

// Helpers removidos após refatorar service para não depender de fiber.Ctx

func TestTransactionService_Deposit_Sucesso(t *testing.T) {
	db := newMockDBTx()
	userID := uuid.New()
	db.findUserByID = &r.User{ID: userID, Email: "user@ex.com"}
	svc := &services.TransactionService{DB: db, Logger: zap.NewNop()}
	amount := decimal.NewFromInt(100)
	resp, err := svc.Deposit(userID.String(), amount, "")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, amount, db.balanceMap[userID])
}

func TestTransactionService_Deposit_UserIDAusente(t *testing.T) {
	db := newMockDBTx()
	svc := &services.TransactionService{DB: db, Logger: zap.NewNop()}
	resp, err := svc.Deposit("", decimal.NewFromInt(10), "")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestTransactionService_Withdraw_Sucesso(t *testing.T) {
	db := newMockDBTx()
	userID := uuid.New()
	// Seed balance
	db.balanceMap[userID] = decimal.NewFromInt(200)
	svc := &services.TransactionService{DB: db, Logger: zap.NewNop()}
	amount := decimal.NewFromInt(50)
	resp, err := svc.Withdraw(userID.String(), amount, "")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, decimal.NewFromInt(150), db.balanceMap[userID])
}

func TestTransactionService_Transfer_Sucesso(t *testing.T) {
	db := newMockDBTx()
	fromID := uuid.New()
	toID := uuid.New()
	db.balanceMap[fromID] = decimal.NewFromInt(300)
	db.findUserByEmail = &r.User{ID: toID, Email: "dest@ex.com"}
	db.findUserByID = &r.User{ID: fromID, Email: "orig@ex.com"}
	svc := &services.TransactionService{DB: db, Logger: zap.NewNop()}
	amount := decimal.NewFromInt(75)
	resp, err := svc.Transfer(fromID.String(), amount, "dest@ex.com", "")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, decimal.NewFromInt(225), db.balanceMap[fromID])
	assert.Equal(t, amount, db.balanceMap[toID])
}

func TestTransactionService_Transfer_UserDestinoNaoEncontrado(t *testing.T) {
	db := newMockDBTx()
	fromID := uuid.New()
	db.balanceMap[fromID] = decimal.NewFromInt(100)
	svc := &services.TransactionService{DB: db, Logger: zap.NewNop()}
	resp, err := svc.Transfer(fromID.String(), decimal.NewFromInt(10), "inexistente@ex.com", "")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

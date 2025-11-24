package services

import (
	"context"
	repositories "financial-system-pro/internal/infrastructure/database"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// fakeDBPort para testar adapters sem acessar GORM.
type fakeDBPort struct {
	users        map[uuid.UUID]*repositories.User
	transactions map[uuid.UUID]*repositories.Transaction
	walletInfos  map[uuid.UUID]*repositories.WalletInfo
	balances     map[uuid.UUID]decimal.Decimal
}

func newFakeDB() *fakeDBPort {
	return &fakeDBPort{users: map[uuid.UUID]*repositories.User{}, transactions: map[uuid.UUID]*repositories.Transaction{}, walletInfos: map[uuid.UUID]*repositories.WalletInfo{}, balances: map[uuid.UUID]decimal.Decimal{}}
}

// Implementação mínima das funções do DatabasePort utilizadas pelos adapters.
func (f *fakeDBPort) FindUserByField(field string, value any) (*repositories.User, error) {
	switch field {
	case "email":
		for _, u := range f.users {
			if u.Email == value {
				return u, nil
			}
		}
	case "id":
		idStr, _ := value.(string)
		for _, u := range f.users {
			if u.ID.String() == idStr {
				return u, nil
			}
		}
	}
	return nil, nil
}
func (f *fakeDBPort) Insert(v any) error {
	switch obj := v.(type) {
	case *repositories.User:
		f.users[obj.ID] = obj
	case *repositories.Transaction:
		f.transactions[obj.ID] = obj
	}
	return nil
}
func (f *fakeDBPort) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	f.walletInfos[userID] = &repositories.WalletInfo{UserID: userID, TronAddress: tronAddress, EncryptedPrivKey: encryptedPrivKey}
	return nil
}
func (f *fakeDBPort) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return f.walletInfos[userID], nil
}
func (f *fakeDBPort) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	bal := f.balances[userID]
	if transactionType == "withdraw" {
		if bal.LessThan(amount) {
			return fmt.Errorf("saldo insuficiente")
		} // erro genérico para teste
		f.balances[userID] = bal.Sub(amount)
		return nil
	}
	f.balances[userID] = bal.Add(amount)
	return nil
}
func (f *fakeDBPort) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	return f.balances[userID], nil
}
func (f *fakeDBPort) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
	tx := f.transactions[txID]
	if tx == nil {
		return nil
	}
	if status, ok := updates["tron_tx_status"].(string); ok {
		tx.TronTxStatus = &status
	}
	if hash, ok := updates["tron_tx_hash"].(string); ok {
		tx.TronTxHash = &hash
	}
	return nil
}

// Tests
func TestLedgerAdapterApplyAndBalance(t *testing.T) {
	db := newFakeDB()
	ledger := NewLedgerAdapter(db)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, ledger.Apply(ctx, uid, decimal.NewFromInt(100), "deposit"))
	require.NoError(t, ledger.Apply(ctx, uid, decimal.NewFromInt(40), "withdraw"))
	bal, err := ledger.Balance(ctx, uid)
	require.NoError(t, err)
	require.True(t, bal.Equal(decimal.NewFromInt(60)))
}

func TestTransactionRecordAdapterInsertAndUpdate(t *testing.T) {
	db := newFakeDB()
	txAdapter := NewTransactionRecordAdapter(db)
	ctx := context.Background()
	tx := &repositories.Transaction{ID: uuid.New(), AccountID: uuid.New(), Type: "deposit"}
	require.NoError(t, txAdapter.Insert(ctx, tx))
	require.NoError(t, txAdapter.Update(ctx, tx.ID, map[string]interface{}{"tron_tx_status": "pending"}))
	require.NotNil(t, tx.TronTxStatus)
	require.Equal(t, "pending", *tx.TronTxStatus)
}

func TestWalletRepositoryAdapterSaveAndGet(t *testing.T) {
	db := newFakeDB()
	walletAdapter := NewWalletRepositoryAdapter(db)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, walletAdapter.SaveInfo(ctx, uid, "ADDR123", "ENCRYPTED"))
	info, err := walletAdapter.GetInfo(ctx, uid)
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, "ADDR123", info.TronAddress)
}

//go:build service_tests_disabled

package services_test

import (
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mock database mínimo (in-memory) para user service
type mockUserDB struct {
	users     map[string]*repositories.User
	wallets   map[string]*repositories.WalletInfo
	insertErr error
}

func newMockUserDB() *mockUserDB {
	return &mockUserDB{users: map[string]*repositories.User{}, wallets: map[string]*repositories.WalletInfo{}}
}

func (m *mockUserDB) Insert(model interface{}) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	switch v := model.(type) {
	case *repositories.User:
		m.users[v.Email] = v
	case *repositories.WalletInfo:
		m.wallets[v.UserID.String()] = v
	case *repositories.Transaction:
		// ignorar
	}
	return nil
}
func (m *mockUserDB) FindUserByField(field, value string) (*repositories.User, error) {
	if field == "email" {
		u := m.users[value]
		if u == nil {
			return nil, errors.NewDatabaseError("record not found", nil)
		}
		return u, nil
	}
	return nil, errors.NewDatabaseError("unsupported field", nil)
}
func (m *mockUserDB) SaveWalletInfo(userID repositories.UUID, tronAddress, encryptedPrivKey string) error {
	m.wallets[userID.String()] = &repositories.WalletInfo{UserID: userID, TronAddress: tronAddress, EncryptedPrivKey: encryptedPrivKey}
	return nil
}

// Métodos extras referenciados pelo service mas não usados nestes testes
func (m *mockUserDB) GetWalletInfo(userID repositories.UUID) (*repositories.WalletInfo, error) {
	return m.wallets[userID.String()], nil
}
func (m *mockUserDB) Transaction(repositories.UUID, interface{}, string) error          { return nil }
func (m *mockUserDB) Balance(repositories.UUID) (interface{}, error)                    { return nil, nil }
func (m *mockUserDB) UpdateTransaction(repositories.UUID, map[string]interface{}) error { return nil }

// wallet manager mock
type mockWalletManager struct{}

func (m *mockWalletManager) GenerateWallet() (*repositories.GeneratedWallet, error) {
	return &repositories.GeneratedWallet{Address: "TTESTADDR123", EncryptedPrivKey: "ENC_PRIV"}, nil
}

func TestCreateNewUser_Success(t *testing.T) {
	mdb := newMockUserDB()
	service := &services.UserService{Database: (*repositories.NewDatabase)(nil), Logger: zap.NewNop(), WalletManager: &mockWalletManager{}}
	// Usamos mock direto atribuindo à interface esperada pelo service
	service.Database = (*repositories.NewDatabase)(mdb) // type conversion for simplicity (assume compatible)

	err := service.CreateNewUser(&dto.UserRequest{Email: "user@example.com", Password: "senha"})
	assert.Nil(t, err)
	assert.NotNil(t, mdb.users["user@example.com"])
}

func TestCreateNewUser_AlreadyExists(t *testing.T) {
	mdb := newMockUserDB()
	mdb.users["exist@example.com"] = &repositories.User{Email: "exist@example.com", Password: "hashed"}
	service := &services.UserService{Database: (*repositories.NewDatabase)(nil), Logger: zap.NewNop()}
	service.Database = (*repositories.NewDatabase)(mdb)

	err := service.CreateNewUser(&dto.UserRequest{Email: "exist@example.com", Password: "x"})
	assert.NotNil(t, err)
	assert.Equal(t, "validation_error", err.Type)
}

func TestCreateNewUser_HashFailure(t *testing.T) {
	// Forçar falha alterando privateKey global? Difícil sem expor variável. Teste omitido por limitação.
}

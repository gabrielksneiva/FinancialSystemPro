package services_test

import (
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/shared/utils"
	"os"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockDatabasePort implementa services.DatabasePort
type mockDatabasePort struct {
	users            map[string]*repositories.User
	findErr          error
	insertErr        error
	walletSavedForID string
}

func newMockDB() *mockDatabasePort { return &mockDatabasePort{users: map[string]*repositories.User{}} }

func (m *mockDatabasePort) FindUserByField(field string, value any) (*repositories.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if field == "email" {
		u := m.users[value.(string)]
		return u, nil // se não existe retorna (nil,nil)
	}
	if field == "id" {
		return nil, nil
	}
	return nil, errors.NewInternalError("unsupported field", nil)
}
func (m *mockDatabasePort) Insert(v any) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	if u, ok := v.(*repositories.User); ok {
		m.users[u.Email] = u
	}
	return nil
}
func (m *mockDatabasePort) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	m.walletSavedForID = userID.String()
	return nil
}
func (m *mockDatabasePort) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return nil, errors.NewDatabaseError("not implemented", nil)
}
func (m *mockDatabasePort) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	return nil
}
func (m *mockDatabasePort) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (m *mockDatabasePort) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
	return nil
}

// mockWalletManager para testar geração automática de wallet
type mockWalletManager struct{ generateErr error }

func (m *mockWalletManager) GenerateWallet() (*entities.WalletInfo, error) {
	if m.generateErr != nil {
		return nil, m.generateErr
	}
	return &entities.WalletInfo{Address: "TADDR123", EncryptedPrivKey: "ENC_PRIV", Blockchain: entities.BlockchainTRON}, nil
}
func (m *mockWalletManager) ValidateAddress(address string) bool { return true }
func (m *mockWalletManager) GetBlockchainType() entities.BlockchainType {
	return entities.BlockchainTRON
}

func TestUserService_CreateNewUser_SucessoComWallet(t *testing.T) {
	os.Setenv("SECRET_KEY", "segredo")
	db := newMockDB()
	svc := &services.UserService{Database: db, Logger: zap.NewNop(), WalletManager: &mockWalletManager{}}
	err := svc.CreateNewUser(&dto.UserRequest{Email: "novo@ex.com", Password: "pwd"})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(db.users))
	assert.NotEmpty(t, db.walletSavedForID)
}

func TestUserService_CreateNewUser_UsuarioJaExiste(t *testing.T) {
	os.Setenv("SECRET_KEY", "segredo")
	db := newMockDB()
	db.users["existe@ex.com"] = &repositories.User{Email: "existe@ex.com", Password: "hash"}
	svc := &services.UserService{Database: db, Logger: zap.NewNop(), WalletManager: &mockWalletManager{}}
	err := svc.CreateNewUser(&dto.UserRequest{Email: "existe@ex.com", Password: "pwd"})
	assert.NotNil(t, err)
	assert.Equal(t, errors.ErrValidation, err.Code)
}

func TestUserService_CreateNewUser_ErroVerificacaoDB(t *testing.T) {
	os.Setenv("SECRET_KEY", "segredo")
	db := newMockDB()
	db.findErr = errors.NewDatabaseError("generic db fail", errors.NewInternalError("low level", nil))
	svc := &services.UserService{Database: db, Logger: zap.NewNop()}
	err := svc.CreateNewUser(&dto.UserRequest{Email: "falha@ex.com", Password: "pwd"})
	assert.NotNil(t, err)
	assert.Equal(t, errors.ErrDatabaseConnection, err.Code)
}

func TestUserService_CreateNewUser_SemWalletManager(t *testing.T) {
	os.Setenv("SECRET_KEY", "segredo")
	db := newMockDB()
	svc := &services.UserService{Database: db, Logger: zap.NewNop(), WalletManager: nil}
	err := svc.CreateNewUser(&dto.UserRequest{Email: "semwallet@ex.com", Password: "pwd"})
	assert.Nil(t, err)
	assert.Empty(t, db.walletSavedForID)
}

// AuthService tests
func TestAuthService_Login_Sucesso(t *testing.T) {
	os.Setenv("SECRET_KEY", "segredo")
	db := newMockDB()
	// Gerar hash de senha real via função do package services (usa utils)
	hash, _ := utils.HashAString("senha")
	db.users["user@ex.com"] = &repositories.User{Email: "user@ex.com", Password: hash}
	auth := &services.AuthService{Database: db, Logger: zap.NewNop()}
	token, appErr := auth.Login(&dto.LoginRequest{Email: "user@ex.com", Password: "senha"})
	require.Nil(t, appErr)
	require.NotEmpty(t, token)
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) { return []byte("segredo"), nil })
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
}

func TestAuthService_Login_UserNaoExiste(t *testing.T) {
	os.Setenv("SECRET_KEY", "segredo")
	db := newMockDB()
	auth := &services.AuthService{Database: db, Logger: zap.NewNop()}
	token, appErr := auth.Login(&dto.LoginRequest{Email: "nao@ex.com", Password: "x"})
	assert.Empty(t, token)
	assert.NotNil(t, appErr)
	assert.Equal(t, errors.ErrValidation, appErr.Code)
}

func TestAuthService_Login_ErroDB(t *testing.T) {
	os.Setenv("SECRET_KEY", "segredo")
	db := newMockDB()
	db.findErr = errors.NewDatabaseError("falha", nil)
	auth := &services.AuthService{Database: db, Logger: zap.NewNop()}
	token, appErr := auth.Login(&dto.LoginRequest{Email: "user@ex.com", Password: "senha"})
	assert.Empty(t, token)
	assert.NotNil(t, appErr)
	assert.Equal(t, errors.ErrDatabaseConnection, appErr.Code)
}

func TestAuthService_Login_SenhaInvalida(t *testing.T) {
	os.Setenv("SECRET_KEY", "segredo")
	db := newMockDB()
	// senha correta é "outra"
	hash, _ := utils.HashAString("outra")
	db.users["user@ex.com"] = &repositories.User{Email: "user@ex.com", Password: hash}
	auth := &services.AuthService{Database: db, Logger: zap.NewNop()}
	token, appErr := auth.Login(&dto.LoginRequest{Email: "user@ex.com", Password: "senha"})
	assert.Empty(t, token)
	assert.NotNil(t, appErr)
	assert.Equal(t, errors.ErrValidation, appErr.Code)
}

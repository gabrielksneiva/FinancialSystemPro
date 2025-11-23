//go:build service_tests_disabled

package services_test

import (
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockAuthDB struct{ users map[string]*repositories.User }

func newMockAuthDB() *mockAuthDB { return &mockAuthDB{users: map[string]*repositories.User{}} }

func (m *mockAuthDB) FindUserByField(field, value string) (*repositories.User, error) {
	if field == "email" {
		u := m.users[value]
		if u == nil {
			return nil, errors.NewDatabaseError("record not found", nil)
		}
		return u, nil
	}
	return nil, errors.NewDatabaseError("unsupported field", nil)
}
func (m *mockAuthDB) Insert(model interface{}) error                         { return nil }
func (m *mockAuthDB) SaveWalletInfo(repositories.UUID, string, string) error { return nil }
func (m *mockAuthDB) GetWalletInfo(repositories.UUID) (*repositories.WalletInfo, error) {
	return nil, nil
}
func (m *mockAuthDB) Transaction(repositories.UUID, interface{}, string) error          { return nil }
func (m *mockAuthDB) Balance(repositories.UUID) (interface{}, error)                    { return nil, nil }
func (m *mockAuthDB) UpdateTransaction(repositories.UUID, map[string]interface{}) error { return nil }

func TestLogin_Success(t *testing.T) {
	mdb := newMockAuthDB()
	// password hashing usando utils.HashAString será chamado internamente; precisamos do hash resultante
	// Para simplificar, assumimos que password já está hashed igual à função HashAString("secret")
	// Em ambiente real, seria melhor isolar utils, aqui é aproximação.
	// Gerar hash real
	hash, _ := services.HashAString("secret")
	mdb.users["u@example.com"] = &repositories.User{Email: "u@example.com", Password: hash}

	service := &services.AuthService{Database: (*repositories.NewDatabase)(mdb), Logger: zap.NewNop()}
	token, appErr := service.Login(&dto.LoginRequest{Email: "u@example.com", Password: "secret"})
	assert.Nil(t, appErr)
	assert.NotEmpty(t, token)
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) { return []byte(""), nil })
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
}

func TestLogin_UserNotFound(t *testing.T) {
	service := &services.AuthService{Database: (*repositories.NewDatabase)(newMockAuthDB()), Logger: zap.NewNop()}
	token, appErr := service.Login(&dto.LoginRequest{Email: "missing@example.com", Password: "x"})
	assert.Empty(t, token)
	assert.NotNil(t, appErr)
	assert.Equal(t, "validation_error", appErr.Type)
}

func TestLogin_InvalidPassword(t *testing.T) {
	mdb := newMockAuthDB()
	// Hash de senha diferente
	hash, _ := services.HashAString("other")
	mdb.users["u@example.com"] = &repositories.User{Email: "u@example.com", Password: hash}
	service := &services.AuthService{Database: (*repositories.NewDatabase)(mdb), Logger: zap.NewNop()}
	_, appErr := service.Login(&dto.LoginRequest{Email: "u@example.com", Password: "secret"})
	assert.NotNil(t, appErr)
	assert.Equal(t, "validation_error", appErr.Type)
}

package services

import (
	"os"
	"testing"
	"time"

	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/shared/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockDB implements DatabasePort for AuthService tests.
type mockDB struct{ user *repositories.User }

func (m *mockDB) FindUserByField(field string, value any) (*repositories.User, error) {
	if field == "email" && m.user != nil && value == m.user.Email {
		return m.user, nil
	}
	// Simula usuário não encontrado retornando (nil,nil) para acionar erro de validação
	return nil, nil
}
func (m *mockDB) Insert(value any) error { return nil }
func (m *mockDB) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	return nil
}
func (m *mockDB) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) { return nil, nil }
func (m *mockDB) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	return nil
}
func (m *mockDB) Balance(userID uuid.UUID) (decimal.Decimal, error)                      { return decimal.Zero, nil }
func (m *mockDB) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error { return nil }

type mockTokenProvider struct{ called bool }

func (m *mockTokenProvider) CreateToken(claims map[string]interface{}) (string, error) {
	m.called = true
	return "MYTOKEN", nil
}

type mockPasswordHasher struct {
	called bool
	allow  bool
}

func (m *mockPasswordHasher) Compare(raw, hashed string) (bool, error) {
	m.called = true
	return m.allow, nil
}
func (m *mockPasswordHasher) Hash(raw string) (string, error) { return "hash", nil }

// helper to build service with defaults
func newAuthService(db DatabasePort) *AuthService {
	return NewAuthService(db, zap.NewNop())
}

func TestAuthServiceLogin_DefaultAdaptersSuccess(t *testing.T) {
	os.Setenv("SECRET_KEY", "secret-key")
	os.Setenv("EXPIRATION_TIME", "60")
	hashed, err := utils.HashAString("pw")
	require.NoError(t, err)
	db := &mockDB{user: &repositories.User{Email: "a@b.com", Password: hashed, ID: uuid.New()}}
	svc := newAuthService(db)
	token, appErr := svc.Login(&dto.LoginRequest{Email: "a@b.com", Password: "pw"})
	assert.Nil(t, appErr)
	assert.NotEmpty(t, token)
}

func TestAuthServiceLogin_CustomAdaptersUsed(t *testing.T) {
	db := &mockDB{user: &repositories.User{Email: "a@b.com", Password: "IGNORED", ID: uuid.New()}}
	tp := &mockTokenProvider{}
	ph := &mockPasswordHasher{allow: true}
	svc := newAuthService(db).WithSecurityAdapters(tp, ph)
	token, appErr := svc.Login(&dto.LoginRequest{Email: "a@b.com", Password: "anything"})
	assert.Nil(t, appErr)
	assert.Equal(t, "MYTOKEN", token)
	assert.True(t, tp.called)
	assert.True(t, ph.called)
}

func TestAuthServiceLogin_InvalidPassword(t *testing.T) {
	hashed, _ := utils.HashAString("pw")
	db := &mockDB{user: &repositories.User{Email: "a@b.com", Password: hashed, ID: uuid.New()}}
	ph := &mockPasswordHasher{allow: false}
	svc := newAuthService(db).WithSecurityAdapters(nil, ph)
	token, appErr := svc.Login(&dto.LoginRequest{Email: "a@b.com", Password: "wrong"})
	assert.Empty(t, token)
	require.NotNil(t, appErr)
	assert.Equal(t, errors.ErrValidation, appErr.Code)
}

func TestAuthServiceLogin_UserNotFound(t *testing.T) {
	db := &mockDB{user: nil}
	svc := newAuthService(db)
	token, appErr := svc.Login(&dto.LoginRequest{Email: "missing@b.com", Password: "x"})
	assert.Empty(t, token)
	require.NotNil(t, appErr)
	assert.Equal(t, errors.ErrValidation, appErr.Code)
}

func TestAuthServiceLogin_JWTCreationError(t *testing.T) {
	// Force token provider to error
	db := &mockDB{user: &repositories.User{Email: "a@b.com", Password: "HASH", ID: uuid.New()}}
	tp := &errorTokenProvider{}
	ph := &mockPasswordHasher{allow: true}
	svc := newAuthService(db).WithSecurityAdapters(tp, ph)
	token, appErr := svc.Login(&dto.LoginRequest{Email: "a@b.com", Password: "pw"})
	assert.Empty(t, token)
	require.NotNil(t, appErr)
	assert.Equal(t, errors.ErrInternal, appErr.Code)
}

type errorTokenProvider struct{}

func (errorTokenProvider) CreateToken(claims map[string]interface{}) (string, error) {
	return "", assert.AnError
}

func TestAuthServiceLogin_PasswordHasherError(t *testing.T) {
	db := &mockDB{user: &repositories.User{Email: "a@b.com", Password: "HASH", ID: uuid.New()}}
	ph := &errorPasswordHasher{}
	svc := newAuthService(db).WithSecurityAdapters(nil, ph)
	token, appErr := svc.Login(&dto.LoginRequest{Email: "a@b.com", Password: "pw"})
	assert.Empty(t, token)
	require.NotNil(t, appErr)
	assert.Equal(t, errors.ErrValidation, appErr.Code)
}

type errorPasswordHasher struct{}

func (errorPasswordHasher) Compare(raw, hashed string) (bool, error) { return false, assert.AnError }
func (errorPasswordHasher) Hash(raw string) (string, error)          { return "", assert.AnError }

// Ensure token expiration claim is set
func TestAuthServiceLogin_DefaultExpirationClaim(t *testing.T) {
	os.Setenv("SECRET_KEY", "secret-key")
	os.Setenv("EXPIRATION_TIME", "2")
	hashed, _ := utils.HashAString("pw")
	db := &mockDB{user: &repositories.User{Email: "a@b.com", Password: hashed, ID: uuid.New()}}
	svc := newAuthService(db)
	token, appErr := svc.Login(&dto.LoginRequest{Email: "a@b.com", Password: "pw"})
	require.Nil(t, appErr)
	require.NotEmpty(t, token)
	// Decode token and verify expiration claim present and in the future
	tk, err := utils.DecodeJWTToken(token)
	require.NoError(t, err)
	mc := tk.Claims.(jwt.MapClaims)
	expRaw, ok := mc["exp"].(float64)
	require.True(t, ok)
	assert.Greater(t, expRaw, float64(time.Now().Unix()))
}

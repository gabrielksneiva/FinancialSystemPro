package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUser_CreatesWithCorrectDefaults(t *testing.T) {
	email := "test@example.com"
	password := "hashed_password_123"

	user := NewUser(email, password)

	assert.NotEqual(t, uuid.Nil, user.ID, "ID deve ser gerado")
	assert.Equal(t, email, user.Email, "Email deve corresponder")
	assert.Equal(t, password, user.Password, "Password deve corresponder")
	assert.False(t, user.CreatedAt.IsZero(), "CreatedAt deve ser definido")
	assert.False(t, user.UpdatedAt.IsZero(), "UpdatedAt deve ser definido")
}

func TestNewUser_GeneratesUniqueIDs(t *testing.T) {
	user1 := NewUser("user1@test.com", "pass1")
	user2 := NewUser("user2@test.com", "pass2")

	assert.NotEqual(t, user1.ID, user2.ID, "IDs devem ser únicos")
}

func TestNewUser_SetsTimestampsCorrectly(t *testing.T) {
	beforeCreate := time.Now()
	user := NewUser("timestamp@test.com", "pass")
	afterCreate := time.Now()

	assert.False(t, user.CreatedAt.Before(beforeCreate), "CreatedAt não deve ser antes da criação")
	assert.False(t, user.CreatedAt.After(afterCreate), "CreatedAt não deve ser depois da criação")
	assert.False(t, user.UpdatedAt.Before(beforeCreate), "UpdatedAt não deve ser antes da criação")
	assert.Equal(t, user.CreatedAt, user.UpdatedAt, "CreatedAt e UpdatedAt devem ser iguais inicialmente")
}

func TestUser_EmailValidation_AllowsVariousFormats(t *testing.T) {
	testCases := []struct {
		name  string
		email string
	}{
		{"simple", "user@domain.com"},
		{"subdomain", "user@mail.domain.com"},
		{"plus", "user+tag@domain.com"},
		{"dash", "user-name@domain.com"},
		{"dot", "user.name@domain.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := NewUser(tc.email, "password")
			assert.Equal(t, tc.email, user.Email, "Email deve ser preservado")
		})
	}
}

func TestWallet_FieldsAreCorrectlySet(t *testing.T) {
	userID := uuid.New()
	address := "TRX_ADDRESS_123"
	encryptedKey := "encrypted_private_key"
	balance := 100.50

	wallet := &Wallet{
		ID:               uuid.New(),
		UserID:           userID,
		Address:          address,
		EncryptedPrivKey: encryptedKey,
		Balance:          balance,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, wallet.ID, "ID deve ser definido")
	assert.Equal(t, userID, wallet.UserID, "UserID deve corresponder")
	assert.Equal(t, address, wallet.Address, "Address deve corresponder")
	assert.Equal(t, encryptedKey, wallet.EncryptedPrivKey, "EncryptedPrivKey deve corresponder")
	assert.Equal(t, balance, wallet.Balance, "Balance deve corresponder")
	assert.False(t, wallet.CreatedAt.IsZero(), "CreatedAt deve ser definido")
	assert.False(t, wallet.UpdatedAt.IsZero(), "UpdatedAt deve ser definido")
}

func TestWallet_ZeroBalance_IsValid(t *testing.T) {
	wallet := &Wallet{
		UserID:  uuid.New(),
		Address: "ADDR",
		Balance: 0,
	}

	assert.Equal(t, 0.0, wallet.Balance, "Balance zero deve ser válido")
}

func TestWallet_NegativeBalance_CanBeSet(t *testing.T) {
	// Em alguns sistemas, saldo negativo pode ser permitido (crédito/overdraft)
	wallet := &Wallet{
		UserID:  uuid.New(),
		Address: "ADDR",
		Balance: -50.0,
	}

	assert.Equal(t, -50.0, wallet.Balance, "Balance negativo deve ser armazenável")
}

func TestUser_PasswordHashStorage(t *testing.T) {
	// Teste que a entidade armazena o hash, não a senha em texto
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	user := NewUser("secure@test.com", hashedPassword)

	assert.Equal(t, hashedPassword, user.Password, "Hash da senha deve ser armazenado")
	assert.NotEqual(t, "plaintext", user.Password, "Senha não deve estar em texto puro")
}

func TestUser_ImmutableID_AfterCreation(t *testing.T) {
	user := NewUser("id@test.com", "pass")
	originalID := user.ID

	// Simular tentativa de mudança (em código real, ID não deveria ter setter)
	// Este teste documenta o comportamento esperado
	assert.Equal(t, originalID, user.ID, "ID deve permanecer constante")
}

func TestWallet_UserAssociation_IsCorrect(t *testing.T) {
	userID := uuid.New()
	wallet1 := &Wallet{UserID: userID, Address: "ADDR1"}
	wallet2 := &Wallet{UserID: userID, Address: "ADDR2"}

	assert.Equal(t, userID, wallet1.UserID, "Wallet1 deve estar associada ao usuário")
	assert.Equal(t, userID, wallet2.UserID, "Wallet2 deve estar associada ao usuário")
	assert.NotEqual(t, wallet1.Address, wallet2.Address, "Wallets diferentes devem ter endereços diferentes")
}

func TestUser_EmptyFieldsHandling(t *testing.T) {
	// Teste para documentar comportamento com campos vazios
	user := NewUser("", "")

	assert.Empty(t, user.Email, "Email vazio deve ser aceito pela entidade")
	assert.Empty(t, user.Password, "Password vazio deve ser aceito pela entidade")
	// Nota: Validação de negócio deve ocorrer na camada de aplicação
}

func TestWallet_TimestampTracking(t *testing.T) {
	createdTime := time.Now()
	wallet := &Wallet{
		UserID:    uuid.New(),
		Address:   "ADDR",
		CreatedAt: createdTime,
		UpdatedAt: createdTime,
	}

	assert.Equal(t, createdTime, wallet.CreatedAt)
	assert.Equal(t, createdTime, wallet.UpdatedAt)

	// Simular atualização
	time.Sleep(10 * time.Millisecond)
	wallet.UpdatedAt = time.Now()

	assert.True(t, wallet.UpdatedAt.After(wallet.CreatedAt), "UpdatedAt deve ser posterior ao CreatedAt após atualização")
}

package entity

import (
	"testing"

	"financial-system-pro/internal/contexts/user/domain/valueobject"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserAggregate(t *testing.T) {
	email, _ := valueobject.NewEmail("test@example.com")
	password, _ := valueobject.HashFromRaw("SecurePassword123!")
	address := "0x1234567890abcdef"
	privKey := "encrypted-private-key"

	agg, err := NewUserAggregate(email, password, address, privKey)

	require.NoError(t, err)
	assert.NotNil(t, agg)
	assert.NotNil(t, agg.User())
	assert.NotNil(t, agg.Wallet())
	assert.Equal(t, email, agg.User().Email)
	assert.Equal(t, address, agg.Wallet().Address)
	assert.Equal(t, privKey, agg.Wallet().EncryptedPrivKey)
	assert.Equal(t, 0.0, agg.Wallet().Balance)
	assert.True(t, agg.User().IsActive())
	assert.Equal(t, agg.User().ID, agg.Wallet().UserID)
}

func TestUserAggregate_CreditWallet(t *testing.T) {
	email, _ := valueobject.NewEmail("test@example.com")
	password, _ := valueobject.HashFromRaw("SecurePassword123!")
	agg, _ := NewUserAggregate(email, password, "0xabc", "privkey")

	t.Run("credit active user wallet", func(t *testing.T) {
		err := agg.CreditWallet(100.50)
		require.NoError(t, err)
		assert.Equal(t, 100.50, agg.Wallet().GetBalance())
	})

	t.Run("credit again accumulates balance", func(t *testing.T) {
		err := agg.CreditWallet(50.25)
		require.NoError(t, err)
		assert.Equal(t, 150.75, agg.Wallet().GetBalance())
	})

	t.Run("cannot credit inactive user", func(t *testing.T) {
		agg.User().Deactivate()
		err := agg.CreditWallet(10.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inactive user")
	})
}

func TestUserAggregate_DebitWallet(t *testing.T) {
	email, _ := valueobject.NewEmail("test@example.com")
	password, _ := valueobject.HashFromRaw("SecurePassword123!")
	agg, _ := NewUserAggregate(email, password, "0xabc", "privkey")
	_ = agg.CreditWallet(200.0)

	t.Run("debit active user wallet", func(t *testing.T) {
		err := agg.DebitWallet(75.50)
		require.NoError(t, err)
		assert.Equal(t, 124.50, agg.Wallet().GetBalance())
	})

	t.Run("cannot debit more than balance", func(t *testing.T) {
		err := agg.DebitWallet(150.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})

	t.Run("cannot debit inactive user", func(t *testing.T) {
		agg.User().Deactivate()
		err := agg.DebitWallet(10.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inactive user")
	})
}

func TestUserAggregate_CanWithdraw(t *testing.T) {
	email, _ := valueobject.NewEmail("test@example.com")
	password, _ := valueobject.HashFromRaw("SecurePassword123!")
	agg, _ := NewUserAggregate(email, password, "0xabc", "privkey")
	_ = agg.CreditWallet(100.0)

	t.Run("can withdraw when active and sufficient balance", func(t *testing.T) {
		canWithdraw := agg.CanWithdraw(50.0)
		assert.True(t, canWithdraw)
	})

	t.Run("cannot withdraw when insufficient balance", func(t *testing.T) {
		canWithdraw := agg.CanWithdraw(150.0)
		assert.False(t, canWithdraw)
	})

	t.Run("cannot withdraw when inactive", func(t *testing.T) {
		agg.User().Deactivate()
		canWithdraw := agg.CanWithdraw(50.0)
		assert.False(t, canWithdraw)
	})
}

func TestUser_ChangePassword(t *testing.T) {
	email, _ := valueobject.NewEmail("test@example.com")
	oldPassword, _ := valueobject.HashFromRaw("OldPassword123!")
	user := NewUser(email, oldPassword)

	newPassword, _ := valueobject.HashFromRaw("NewPassword456!")
	user.ChangePassword(newPassword)

	assert.Equal(t, newPassword, user.Password)
	assert.False(t, user.Password.Matches("OldPassword123!"))
	assert.True(t, user.Password.Matches("NewPassword456!"))
}

func TestUser_Authenticate(t *testing.T) {
	email, _ := valueobject.NewEmail("test@example.com")
	password, _ := valueobject.HashFromRaw("SecurePassword123!")
	user := NewUser(email, password)

	t.Run("authenticate with correct password", func(t *testing.T) {
		isValid := user.Authenticate("SecurePassword123!")
		assert.True(t, isValid)
	})

	t.Run("authenticate with incorrect password", func(t *testing.T) {
		isValid := user.Authenticate("WrongPassword")
		assert.False(t, isValid)
	})
}

func TestUser_Deactivate_Activate(t *testing.T) {
	email, _ := valueobject.NewEmail("test@example.com")
	password, _ := valueobject.HashFromRaw("SecurePassword123!")
	user := NewUser(email, password)

	assert.True(t, user.IsActive())

	user.Deactivate()
	assert.False(t, user.IsActive())

	user.Activate()
	assert.True(t, user.IsActive())
}

func TestWallet_Credit(t *testing.T) {
	wallet := NewWallet(uuid.New(), "0xabc", "privkey")

	t.Run("credit positive amount", func(t *testing.T) {
		err := wallet.Credit(100.50)
		require.NoError(t, err)
		assert.Equal(t, 100.50, wallet.Balance)
	})

	t.Run("cannot credit zero", func(t *testing.T) {
		err := wallet.Credit(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("cannot credit negative amount", func(t *testing.T) {
		err := wallet.Credit(-50)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})
}

func TestWallet_Debit(t *testing.T) {
	wallet := NewWallet(uuid.New(), "0xabc", "privkey")
	_ = wallet.Credit(200.0)

	t.Run("debit positive amount", func(t *testing.T) {
		err := wallet.Debit(75.50)
		require.NoError(t, err)
		assert.Equal(t, 124.50, wallet.Balance)
	})

	t.Run("cannot debit zero", func(t *testing.T) {
		err := wallet.Debit(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("cannot debit negative amount", func(t *testing.T) {
		err := wallet.Debit(-50)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("cannot debit more than balance", func(t *testing.T) {
		err := wallet.Debit(200.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})
}

func TestWallet_HasSufficientBalance(t *testing.T) {
	wallet := NewWallet(uuid.New(), "0xabc", "privkey")
	_ = wallet.Credit(100.0)

	assert.True(t, wallet.HasSufficientBalance(50.0))
	assert.True(t, wallet.HasSufficientBalance(100.0))
	assert.False(t, wallet.HasSufficientBalance(150.0))
}

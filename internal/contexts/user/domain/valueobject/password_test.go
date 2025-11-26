package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashFromRaw_ValidPassword(t *testing.T) {
	hashed, err := HashFromRaw("mypassword123")

	assert.NoError(t, err)
	assert.NotEmpty(t, hashed.String())
	assert.Equal(t, 64, len(hashed.String())) // SHA-256 hex = 64 chars
}

func TestHashFromRaw_TooShort(t *testing.T) {
	_, err := HashFromRaw("12345")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password too short")
}

func TestHashFromRaw_MinimumLength(t *testing.T) {
	hashed, err := HashFromRaw("123456")

	assert.NoError(t, err)
	assert.NotEmpty(t, hashed.String())
}

func TestHashFromRaw_EmptyString(t *testing.T) {
	_, err := HashFromRaw("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password too short")
}

func TestHashedPassword_Matches_Correct(t *testing.T) {
	raw := "mySecurePassword"
	hashed, _ := HashFromRaw(raw)

	assert.True(t, hashed.Matches(raw))
}

func TestHashedPassword_Matches_Incorrect(t *testing.T) {
	hashed, _ := HashFromRaw("correctPassword")

	assert.False(t, hashed.Matches("wrongPassword"))
}

func TestHashedPassword_Matches_EmptyInput(t *testing.T) {
	hashed, _ := HashFromRaw("password123")

	assert.False(t, hashed.Matches(""))
}

func TestHashedPassword_String(t *testing.T) {
	hashed, _ := HashFromRaw("testpassword")

	str := hashed.String()
	assert.NotEmpty(t, str)
	assert.Equal(t, 64, len(str))
}

func TestHashFromRaw_Deterministic(t *testing.T) {
	raw := "samepassword"
	hash1, _ := HashFromRaw(raw)
	hash2, _ := HashFromRaw(raw)

	assert.Equal(t, hash1.String(), hash2.String())
}

func TestHashFromRaw_DifferentInputs(t *testing.T) {
	hash1, _ := HashFromRaw("password1")
	hash2, _ := HashFromRaw("password2")

	assert.NotEqual(t, hash1.String(), hash2.String())
}

func TestHashedPassword_CaseSensitive(t *testing.T) {
	hashed, _ := HashFromRaw("Password")

	assert.True(t, hashed.Matches("Password"))
	assert.False(t, hashed.Matches("password"))
	assert.False(t, hashed.Matches("PASSWORD"))
}

func TestHashedPassword_SpecialCharacters(t *testing.T) {
	raw := "p@ssw0rd!#$"
	hashed, _ := HashFromRaw(raw)

	assert.True(t, hashed.Matches(raw))
	assert.False(t, hashed.Matches("p@ssw0rd!#"))
}

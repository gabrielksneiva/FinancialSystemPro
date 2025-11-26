package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail_ValidFormat(t *testing.T) {
	email, err := NewEmail("test@example.com")

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email.String())
}

func TestNewEmail_ValidWithSubdomain(t *testing.T) {
	email, err := NewEmail("user@mail.example.co.uk")

	assert.NoError(t, err)
	assert.Equal(t, "user@mail.example.co.uk", email.String())
}

func TestNewEmail_InvalidMissingAt(t *testing.T) {
	_, err := NewEmail("testexample.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNewEmail_InvalidMissingDomain(t *testing.T) {
	_, err := NewEmail("test@")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNewEmail_InvalidMissingLocalPart(t *testing.T) {
	_, err := NewEmail("@example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNewEmail_InvalidNoExtension(t *testing.T) {
	_, err := NewEmail("test@example")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNewEmail_EmptyString(t *testing.T) {
	_, err := NewEmail("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNewEmail_WithSpaces(t *testing.T) {
	_, err := NewEmail("test @example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNewEmail_MultipleAt(t *testing.T) {
	_, err := NewEmail("test@@example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestEmail_String(t *testing.T) {
	email, _ := NewEmail("admin@company.org")

	assert.Equal(t, "admin@company.org", email.String())
}

func TestEmail_AsString(t *testing.T) {
	email, _ := NewEmail("user@test.io")
	str := string(email)

	assert.Equal(t, "user@test.io", str)
}

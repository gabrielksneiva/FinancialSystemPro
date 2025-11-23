package utils_test

import (
	"encoding/base64"
	"financial-system-pro/internal/shared/utils"
	"os" // retained for Unsetenv
	"testing"

	"github.com/stretchr/testify/assert"
)

func encodeKey(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func TestEncryptDecryptPrivateKey_Sucesso(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	t.Setenv("ENCRYPTION_KEY", encodeKey(key))

	plaintext := "MINHA_CHAVE_PRIVADA"
	enc, err := utils.EncryptPrivateKey(plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, enc)

	dec, err := utils.DecryptPrivateKey(enc)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, dec)
}

func TestEncryptPrivateKey_ChaveAusente(t *testing.T) {
	os.Unsetenv("ENCRYPTION_KEY")
	_, err := utils.EncryptPrivateKey("x")
	assert.Error(t, err)
}

func TestDecryptPrivateKey_CipherInvalido(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(255 - i)
	}
	t.Setenv("ENCRYPTION_KEY", encodeKey(key))
	_, err := utils.DecryptPrivateKey("invalido!!")
	assert.Error(t, err)
}

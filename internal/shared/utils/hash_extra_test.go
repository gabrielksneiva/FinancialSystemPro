package utils_test

import (
	"financial-system-pro/internal/shared/utils"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Garantir SECRET_KEY e EXPIRATION_TIME para testes
	os.Setenv("SECRET_KEY", "segredo_para_tests")
	os.Setenv("EXPIRATION_TIME", "1")
}

func TestHashAndCompareTwoStrings_Success(t *testing.T) {
	ok, err := utils.HashAndCompareTwoStrings("abc", mustHash("abc"))
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestHashAndCompareTwoStrings_Fail(t *testing.T) {
	ok, err := utils.HashAndCompareTwoStrings("abc", mustHash("xyz"))
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCreateAndDecodeJWTToken(t *testing.T) {
	tokenStr, err := utils.CreateJWTToken(jwt.MapClaims{"ID": "123"})
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	tok, err := utils.DecodeJWTToken(tokenStr)
	assert.NoError(t, err)
	claims, _ := tok.Claims.(jwt.MapClaims)
	assert.Equal(t, "123", claims["ID"])
}

// helper
func mustHash(s string) string {
	h, _ := utils.HashAString(s)
	return h
}

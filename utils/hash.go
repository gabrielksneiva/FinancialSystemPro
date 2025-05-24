package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
)

var privateKey = []byte(os.Getenv("PRIVATE_KEY")) // Troque por uma chave segura

func hashPassword(password string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

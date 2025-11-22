package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// GetEncryptionKey retorna a chave de criptografia de ambiente
// Idealmente, deve ser uma chave de 32 bytes (AES-256)
func GetEncryptionKey() ([]byte, error) {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY não configurada")
	}

	// Decodificar de base64
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar ENCRYPTION_KEY: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY deve ter 32 bytes para AES-256, tem %d", len(key))
	}

	return key, nil
}

// EncryptPrivateKey criptografa a private key usando AES-256-GCM
func EncryptPrivateKey(privateKey string) (string, error) {
	key, err := GetEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("erro ao criar cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("erro ao criar GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("erro ao gerar nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(privateKey), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey descriptografa a private key criptografada
func DecryptPrivateKey(encryptedKey string) (string, error) {
	key, err := GetEncryptionKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar chave criptografada: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("erro ao criar cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("erro ao criar GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext muito pequeno")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("erro ao descriptografar: %w", err)
	}

	return string(plaintext), nil
}

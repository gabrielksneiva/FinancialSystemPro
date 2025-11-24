package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// EncryptionProviderPort define operações de criptografia simétrica para chaves privadas.
// Implementações devem ser determinísticas apenas na parte de chave mestre; nonce deve ser randômico.
type EncryptionProviderPort interface {
	Encrypt(plain string) (string, error)
	Decrypt(encrypted string) (string, error)
}

// NoopEncryptionProvider não altera o conteúdo (fallback quando chave não configurada).
type NoopEncryptionProvider struct{}

func (n NoopEncryptionProvider) Encrypt(plain string) (string, error)     { return plain, nil }
func (n NoopEncryptionProvider) Decrypt(encrypted string) (string, error) { return encrypted, nil }

// AESEncryptionProvider usa AES-GCM. Formato do payload: base64(nonce|ciphertext|tag)
type AESEncryptionProvider struct {
	gcm cipher.AEAD
}

// NewAESEncryptionProviderFromEnv lê ENCRYPTION_MASTER_KEY (32 bytes base64 ou hex) e cria provider.
// Se não configurada ou inválida, retorna erro.
func NewAESEncryptionProviderFromEnv() (EncryptionProviderPort, error) {
	keyRaw := os.Getenv("ENCRYPTION_MASTER_KEY")
	if keyRaw == "" {
		return NoopEncryptionProvider{}, nil
	}
	key, err := decodeKey(keyRaw)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AESEncryptionProvider{gcm: gcm}, nil
}

func decodeKey(s string) ([]byte, error) {
	// primeiro tenta base64
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		if len(b) == 32 { // AES-256
			return b, nil
		}
	}
	// tenta raw hex (sem validação robusta)
	if len(s) == 64 { // 32 bytes em hex
		out := make([]byte, 32)
		for i := 0; i < 32; i++ {
			var v byte
			for j := 0; j < 2; j++ {
				c := s[i*2+j]
				var nibble byte
				switch {
				case c >= '0' && c <= '9':
					nibble = c - '0'
				case c >= 'a' && c <= 'f':
					nibble = c - 'a' + 10
				case c >= 'A' && c <= 'F':
					nibble = c - 'A' + 10
				default:
					return nil, errors.New("invalid hex master key")
				}
				v = (v << 4) | nibble
			}
			out[i] = v
		}
		return out, nil
	}
	return nil, errors.New("unsupported master key format or length")
}

func (p *AESEncryptionProvider) Encrypt(plain string) (string, error) {
	nonce := make([]byte, p.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := p.gcm.Seal(nil, nonce, []byte(plain), nil)
	payload := append(nonce, ct...) // nonce|ciphertext|tag
	return base64.StdEncoding.EncodeToString(payload), nil
}

func (p *AESEncryptionProvider) Decrypt(encrypted string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	nSize := p.gcm.NonceSize()
	if len(raw) < nSize {
		return "", errors.New("ciphertext too short")
	}
	nonce := raw[:nSize]
	ct := raw[nSize:]
	pt, err := p.gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

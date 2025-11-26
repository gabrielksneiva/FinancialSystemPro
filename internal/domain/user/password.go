package user

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// HashedPassword encapsulates a hashed password value.
type HashedPassword struct {
	sha256Hex string
}

// HashPassword creates a hashed password using SHA-256 (placeholder; replace with bcrypt/argon2 in infra).
func HashPassword(raw string) (HashedPassword, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 8 {
		return HashedPassword{}, errors.New("password too short (<8)")
	}
	h := sha256.Sum256([]byte(raw))
	return HashedPassword{sha256Hex: hex.EncodeToString(h[:])}, nil
}

// Matches verifies a raw password against the stored hash.
func (p HashedPassword) Matches(raw string) bool {
	h := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return p.sha256Hex == hex.EncodeToString(h[:])
}

// String NEVER returns raw password, only hash (for logging cautiously).
func (p HashedPassword) String() string { return p.sha256Hex }

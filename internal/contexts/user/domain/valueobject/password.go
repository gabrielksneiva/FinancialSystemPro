package valueobject

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// HashedPassword VO contendo valor já hash e função de verificação.
type HashedPassword string

func HashFromRaw(raw string) (HashedPassword, error) {
	if len(raw) < 6 {
		return "", errors.New("password too short")
	}
	h := sha256.Sum256([]byte(raw))
	return HashedPassword(hex.EncodeToString(h[:])), nil
}

func (hp HashedPassword) Matches(raw string) bool {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:]) == string(hp)
}

func (hp HashedPassword) String() string { return string(hp) }

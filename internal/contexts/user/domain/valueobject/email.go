package valueobject

import (
	"errors"
	"regexp"
)

// Email VO com validação simples de formato.
type Email string

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func NewEmail(v string) (Email, error) {
	if !emailRegex.MatchString(v) {
		return "", errors.New("invalid email format")
	}
	return Email(v), nil
}

func (e Email) String() string { return string(e) }

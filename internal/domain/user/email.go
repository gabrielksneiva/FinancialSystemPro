package user

import (
	"errors"
	"regexp"
	"strings"
)

// Email represents a validated email value object.
type Email struct {
	value string
}

var emailRx = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)

// NewEmail validates and constructs a new Email value object.
func NewEmail(v string) (Email, error) {
	v = strings.TrimSpace(v)
	if len(v) > 254 {
		return Email{}, errors.New("email length exceeds 254 characters")
	}
	if !emailRx.MatchString(v) {
		return Email{}, errors.New("invalid email format")
	}
	return Email{value: strings.ToLower(v)}, nil
}

// String returns the canonical lowercase email.
func (e Email) String() string { return e.value }

// Equals compares two Email value objects.
func (e Email) Equals(other Email) bool { return e.value == other.value }

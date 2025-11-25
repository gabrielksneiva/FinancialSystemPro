package validator

import (
	"testing"
)

type sample struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=18"`
}

func TestValidator_ValidationErrors(t *testing.T) {
	v := New()
	s := sample{Email: "bad-email", Age: 10}
	appErr := v.Validate(&s)
	if appErr == nil {
		t.Fatalf("esperava erro de validação")
	}
	field, _ := appErr.Details["field"].(string)
	if field != "Email" && field != "Age" { // depende da ordem interna
		t.Fatalf("campo inesperado: %s", field)
	}
}

func TestValidator_NoError(t *testing.T) {
	v := New()
	s := sample{Email: "user@example.com", Age: 25}
	if err := v.Validate(&s); err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
}

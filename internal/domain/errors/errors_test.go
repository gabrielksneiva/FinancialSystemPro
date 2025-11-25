package errors

import (
	"net/http"
	"testing"
)

func TestErrorConstructors(t *testing.T) {
	v := NewValidationError("email", "invalid")
	if v.Code != ErrValidation || v.StatusCode != http.StatusBadRequest {
		t.Fatalf("validation mismatch")
	}

	u := NewUnauthorizedError("no auth")
	if u.Code != ErrUnauthorized || u.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized mismatch")
	}

	i := NewInternalError("boom", nil)
	if i.Code != ErrInternal || i.StatusCode != http.StatusInternalServerError {
		t.Fatalf("internal mismatch")
	}

	d := NewDatabaseError("query", nil)
	if d.Code != ErrDatabaseConnection {
		t.Fatalf("db code mismatch")
	}

	n := NewNotFoundError("user")
	if n.Code != ErrRecordNotFound || n.StatusCode != http.StatusNotFound {
		t.Fatalf("not found mismatch")
	}
}

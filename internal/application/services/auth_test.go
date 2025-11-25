package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"financial-system-pro/internal/application/dto"
	domainErrors "financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
)

// Additional edge-case tests for AuthService covering branches not exercised by auth_service_ports_test.go.

// internalErrorPasswordHasher returns (true, error) to trigger internal error branch after password validation.
type internalErrorPasswordHasher struct{}

func (internalErrorPasswordHasher) Compare(raw, hashed string) (bool, error) {
	return true, errors.New("compare failure")
}
func (internalErrorPasswordHasher) Hash(raw string) (string, error) { return "", errors.New("unused") }

// dbErrorUserRepo forces a database error (not 'record not found').
type dbErrorUserRepo struct{}

func (dbErrorUserRepo) FindByEmail(_ context.Context, email string) (*repositories.User, error) {
	return nil, fmt.Errorf("connection lost")
}
func (dbErrorUserRepo) FindByID(_ context.Context, id uuid.UUID) (*repositories.User, error) {
	return nil, fmt.Errorf("no impl")
}
func (dbErrorUserRepo) Save(_ context.Context, user *repositories.User) error { return nil }

// successUserRepo returns a deterministic user.
type successUserRepo struct{ u *repositories.User }

func (s successUserRepo) FindByEmail(_ context.Context, email string) (*repositories.User, error) {
	return s.u, nil
}
func (s successUserRepo) FindByID(_ context.Context, id uuid.UUID) (*repositories.User, error) {
	return s.u, nil
}
func (s successUserRepo) Save(_ context.Context, user *repositories.User) error { return nil }

// noopTokenProvider returns static token.
type noopTokenProvider struct{}

func (noopTokenProvider) CreateToken(claims map[string]interface{}) (string, error) {
	return "TOK", nil
}

func TestAuthService_Login_InternalPasswordErrorBranch(t *testing.T) {
	// Reuse user from existing ports tests; password value irrelevant (Compare returns true)
	user := &repositories.User{Email: "u@example.com", Password: "hash"}
	svc := &AuthService{UserRepo: successUserRepo{u: user}, tokenProvider: noopTokenProvider{}, passwordHasher: internalErrorPasswordHasher{}}
	token, appErr := svc.Login(&dto.LoginRequest{Email: "u@example.com", Password: "pw"})
	if appErr == nil || appErr.Code != domainErrors.ErrInternal {
		t.Fatalf("expected internal error, got %+v", appErr)
	}
	if token != "" {
		t.Fatalf("expected empty token on internal error")
	}
}

func TestAuthService_Login_DatabaseErrorBranch(t *testing.T) {
	svc := &AuthService{UserRepo: dbErrorUserRepo{}, tokenProvider: noopTokenProvider{}, passwordHasher: internalErrorPasswordHasher{}}
	token, appErr := svc.Login(&dto.LoginRequest{Email: "x@example.com", Password: "pw"})
	if appErr == nil || appErr.Code != domainErrors.ErrDatabaseConnection {
		t.Fatalf("expected database connection error, got %+v", appErr)
	}
	if token != "" {
		t.Fatalf("expected empty token on db error")
	}
}

package services

import (
	"context"
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/errors"
	"financial-system-pro/internal/shared/utils"
	"strings"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type AuthService struct {
	Database       DatabasePort // manter por compatibilidade transitória
	UserRepo       UserRepositoryPort
	Logger         *zap.Logger
	tokenProvider  TokenProvider
	passwordHasher PasswordHasher
}

func NewAuthService(db DatabasePort, logger *zap.Logger) *AuthService {
	// Backward compatible constructor: sets default adapters using existing utils
	return &AuthService{Database: db, UserRepo: NewUserRepositoryAdapter(db), Logger: logger,
		tokenProvider:  defaultTokenProvider{},
		passwordHasher: defaultPasswordHasher{},
	}
}

// WithSecurityAdapters permite injetar implementações customizadas sem quebrar chamadas existentes.
func (a *AuthService) WithSecurityAdapters(tp TokenProvider, ph PasswordHasher) *AuthService {
	if tp != nil {
		a.tokenProvider = tp
	}
	if ph != nil {
		a.passwordHasher = ph
	}
	return a
}

func (a *AuthService) Login(loginData *dto.LoginRequest) (string, *errors.AppError) {
	// Defensive: permitir construção manual do struct em testes legados
	if a.UserRepo == nil && a.Database != nil { // fallback para testes antigos
		a.UserRepo = NewUserRepositoryAdapter(a.Database)
	}
	if a.tokenProvider == nil {
		a.tokenProvider = defaultTokenProvider{}
	}
	if a.passwordHasher == nil {
		a.passwordHasher = defaultPasswordHasher{}
	}
	findUserInfo, err := a.UserRepo.FindByEmail(context.Background(), loginData.Email)
	if err != nil && !strings.EqualFold(err.Error(), "record not found") {
		return "", errors.NewDatabaseError("error finding user by email", nil)
	}
	if findUserInfo == nil {
		return "", errors.NewValidationError("email", "Email not registered")
	}

	verifyPassword, err := a.passwordHasher.Compare(loginData.Password, findUserInfo.Password)
	if !verifyPassword {
		return "", errors.NewValidationError("password", "Invalid password")
	}
	if err != nil {
		return "", errors.NewInternalError("error while comparing passwords", nil)
	}

	jwtToken, err := a.tokenProvider.CreateToken(jwt.MapClaims{"ID": findUserInfo.ID})
	if err != nil {
		if a.Logger != nil {
			a.Logger.Error("error creating JWT token", zap.Error(err))
		}
		return "", errors.NewInternalError("error creating JWT token", nil)
	}

	return jwtToken, nil
}

// Adapters padrão (wrappers) - usam utils existentes para não exigir refatoração ampla imediata.
type defaultTokenProvider struct{}

func (defaultTokenProvider) CreateToken(claims map[string]interface{}) (string, error) {
	// utils.CreateJWTToken espera jwt.MapClaims; convert map
	mc := jwt.MapClaims{}
	for k, v := range claims {
		mc[k] = v
	}
	return utils.CreateJWTToken(mc)
}

type defaultPasswordHasher struct{}

func (defaultPasswordHasher) Compare(raw, hashed string) (bool, error) {
	return utils.HashAndCompareTwoStrings(raw, hashed)
}
func (defaultPasswordHasher) Hash(raw string) (string, error) {
	return utils.HashAString(raw)
}

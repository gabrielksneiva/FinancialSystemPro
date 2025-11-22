package services

import (
	"financial-system-pro/domain"
	"financial-system-pro/repositories"
	"financial-system-pro/utils"
	"strings"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type NewAuthService struct {
	Database *repositories.NewDatabase
	Logger   *zap.Logger
}

func (a *NewAuthService) Login(loginData *domain.LoginRequest) (string, *domain.AppError) {
	findUserInfo, err := a.Database.FindUserByField("email", loginData.Email)
	if err != nil && !strings.EqualFold(err.Error(), "record not found") {
		return "", domain.NewDatabaseError("error finding user by email", nil)
	}
	if findUserInfo == nil {
		return "", domain.NewValidationError("email", "Email not registered")
	}

	verifyPassword, err := utils.HashAndCompareTwoStrings(loginData.Password, findUserInfo.Password)
	if !verifyPassword {
		return "", domain.NewValidationError("password", "Invalid password")
	}
	if err != nil {
		return "", domain.NewInternalError("error while comparing passwords", nil)
	}

	jwtToken, err := utils.CreateJWTToken(jwt.MapClaims{"ID": findUserInfo.ID})
	if err != nil {
		a.Logger.Error("error creating JWT token", zap.Error(err))
		return "", domain.NewInternalError("error creating JWT token", nil)
	}

	return jwtToken, nil
}

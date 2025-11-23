package services

import (
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/shared/utils"
	"strings"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type NewAuthService struct {
	Database *repositories.NewDatabase
	Logger   *zap.Logger
}

func (a *NewAuthService) Login(loginData *dto.LoginRequest) (string, *errors.AppError) {
	findUserInfo, err := a.Database.FindUserByField("email", loginData.Email)
	if err != nil && !strings.EqualFold(err.Error(), "record not found") {
		return "", errors.NewDatabaseError("error finding user by email", nil)
	}
	if findUserInfo == nil {
		return "", errors.NewValidationError("email", "Email not registered")
	}

	verifyPassword, err := utils.HashAndCompareTwoStrings(loginData.Password, findUserInfo.Password)
	if !verifyPassword {
		return "", errors.NewValidationError("password", "Invalid password")
	}
	if err != nil {
		return "", errors.NewInternalError("error while comparing passwords", nil)
	}

	jwtToken, err := utils.CreateJWTToken(jwt.MapClaims{"ID": findUserInfo.ID})
	if err != nil {
		a.Logger.Error("error creating JWT token", zap.Error(err))
		return "", errors.NewInternalError("error creating JWT token", nil)
	}

	return jwtToken, nil
}

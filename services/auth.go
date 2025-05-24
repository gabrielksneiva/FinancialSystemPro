package services

import (
	"financial-system-pro/domain"
	"financial-system-pro/repositories"
	"financial-system-pro/utils"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
)

type NewAuthService struct {
	Database *repositories.NewDatabase
}

func (a *NewAuthService) Login(loginData *domain.LoginRequest) (string, error) {
	findUserInfo, err := a.Database.FindUserByField("email", loginData.Email)
	if err != nil && !strings.EqualFold(err.Error(), "record not found") {
		return "", fmt.Errorf("error while finding user on db: %s", err)
	}
	if findUserInfo == nil {
		return "", fmt.Errorf("user no exists")
	}

	verifyPassword, err := utils.HashAndCompareTwoStrings(loginData.Password, findUserInfo.Password)
	if !verifyPassword {
		return "", fmt.Errorf("invalid password")
	}

	jwtToken, err := utils.CreateJWTToken(jwt.MapClaims{"ID": findUserInfo.ID})
	if err != nil {
		return "", fmt.Errorf("error while create JWT token: %s", err)
	}

	return jwtToken, nil

}

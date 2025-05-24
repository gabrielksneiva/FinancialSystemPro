package services

import (
	"financial-system-pro/domain"
	"financial-system-pro/repositories"
	"financial-system-pro/utils"
	"fmt"
	"strings"
)

type NewUserService struct {
	Database *repositories.NewDatabase
}

func (s *NewUserService) CreateNewUser(userRequest *domain.UserRequest) error {
	exists, err := s.verifyIfUserAlreadyExists(userRequest.Email)
	if err != nil {
		return fmt.Errorf("error on verifying user: %s", err)
	}
	if exists {
		return fmt.Errorf("user already exists")
	}

	hashedPassword, err := utils.HashAString(userRequest.Password)
	if err != nil {
		return fmt.Errorf("error on hash password: %s", err)
	}

	parsedUser := repositories.User{
		Email:    userRequest.Email,
		Password: hashedPassword,
	}

	err = s.Database.Insert(&parsedUser)
	if err != nil {
		return fmt.Errorf("error on insert user: %s", err)
	}

	return nil
}

func (s *NewUserService) verifyIfUserAlreadyExists(email string) (bool, error) {
	result, err := s.Database.FindUserByField("email", email)
	if err != nil && !strings.EqualFold(err.Error(), "record not found") {
		return false, err
	}

	if result != nil {
		return true, nil
	}

	return false, nil
}

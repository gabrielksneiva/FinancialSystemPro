package services

import (
	"financial-system-pro/domain"
	"financial-system-pro/repositories"
	"financial-system-pro/utils"
	"strings"

	"go.uber.org/zap"
)

type NewUserService struct {
	Database      *repositories.NewDatabase
	Logger        *zap.Logger
	WalletManager domain.WalletManager
}

func (s *NewUserService) CreateNewUser(userRequest *domain.UserRequest) *domain.AppError {
	// Verificar se usuário já existe
	exists, err := s.verifyIfUserAlreadyExists(userRequest.Email)
	if err != nil {
		s.Logger.Error("error verifying user existence",
			zap.String("email", userRequest.Email),
			zap.Error(err),
		)
		return domain.NewDatabaseError("user verification failed", nil)
	}

	if exists {
		s.Logger.Info("user already exists",
			zap.String("email", userRequest.Email),
		)
		return domain.NewValidationError("email", "Email already registered")
	}

	// Hash da senha
	hashedPassword, err := utils.HashAString(userRequest.Password)
	if err != nil {
		s.Logger.Error("error hashing password",
			zap.String("email", userRequest.Email),
			zap.Error(err),
		)
		return domain.NewInternalError("password hashing failed", nil)
	}

	// Criar usuário
	parsedUser := repositories.User{
		Email:    userRequest.Email,
		Password: hashedPassword,
	}

	err = s.Database.Insert(&parsedUser)
	if err != nil {
		s.Logger.Error("error inserting user",
			zap.String("email", userRequest.Email),
			zap.Error(err),
		)
		return domain.NewDatabaseError("user creation failed", nil)
	}

	// Gerar carteira TRON automaticamente
	if s.WalletManager != nil {
		wallet, err := s.WalletManager.GenerateWallet()
		if err != nil {
			s.Logger.Warn("error generating wallet, user created but without wallet",
				zap.String("email", userRequest.Email),
				zap.Error(err),
			)
		} else {
			// Criptografar private key (por enquanto apenas armazenar o address)
			errWallet := s.Database.SaveWalletInfo(parsedUser.ID, wallet.Address, "")
			if errWallet != nil {
				s.Logger.Warn("error saving wallet info",
					zap.String("email", userRequest.Email),
					zap.String("user_id", parsedUser.ID.String()),
					zap.Error(errWallet),
				)
			} else {
				s.Logger.Info("wallet generated and saved",
					zap.String("email", userRequest.Email),
					zap.String("user_id", parsedUser.ID.String()),
					zap.String("wallet_address", wallet.Address),
				)
			}
		}
	}

	s.Logger.Info("user created successfully",
		zap.String("email", userRequest.Email),
	)
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

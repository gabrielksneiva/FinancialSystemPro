package services

import (
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"strings"

	"go.uber.org/zap"
)

type UserService struct {
	Database       DatabasePort
	Logger         *zap.Logger
	WalletManager  entities.WalletManager
	passwordHasher PasswordHasher
}

func NewUserService(db DatabasePort, logger *zap.Logger, walletManager entities.WalletManager) *UserService {
	return &UserService{Database: db, Logger: logger, WalletManager: walletManager, passwordHasher: defaultPasswordHasher{}}
}

// WithPasswordHasher permite definir hasher customizado mantendo construtor antigo.
func (s *UserService) WithPasswordHasher(ph PasswordHasher) *UserService {
	if ph != nil {
		s.passwordHasher = ph
	}
	return s
}

func (s *UserService) CreateNewUser(userRequest *dto.UserRequest) *errors.AppError {
	exists, err := s.verifyIfUserAlreadyExists(userRequest.Email)
	if err != nil {
		// logging
		return errors.NewDatabaseError("user verification failed", nil)
	}
	if exists {
		return errors.NewValidationError("email", "Email already registered")
	}

	// Garantir hasher default se struct instanciado manualmente em testes antigos
	if s.passwordHasher == nil {
		s.passwordHasher = defaultPasswordHasher{}
	}
	hashedPassword, err := s.passwordHasher.Hash(userRequest.Password)
	if err != nil {
		return errors.NewInternalError("password hashing failed", nil)
	}

	parsedUser := repositories.User{Email: userRequest.Email, Password: hashedPassword}
	if err := s.Database.Insert(&parsedUser); err != nil {
		return errors.NewDatabaseError("user creation failed", nil)
	}

	if s.WalletManager != nil {
		wallet, err := s.WalletManager.GenerateWallet()
		if err == nil {
			_ = s.Database.SaveWalletInfo(parsedUser.ID, wallet.Address, wallet.EncryptedPrivKey)
		}
	}

	return nil
}

func (s *UserService) verifyIfUserAlreadyExists(email string) (bool, error) {
	result, err := s.Database.FindUserByField("email", email)
	if err != nil && !strings.EqualFold(err.Error(), "record not found") {
		return false, err
	}
	if result != nil {
		return true, nil
	}
	return false, nil
}

func (s *UserService) GetDatabase() DatabasePort { return s.Database }

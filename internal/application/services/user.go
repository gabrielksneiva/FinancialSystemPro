package services

import (
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/shared/utils"
	"strings"

	"go.uber.org/zap"
)

type NewUserService struct {
	Database      *repositories.NewDatabase
	Logger        *zap.Logger
	WalletManager entities.WalletManager
}

func (s *NewUserService) CreateNewUser(userRequest *dto.UserRequest) *errors.AppError {
	// Verificar se usuário já existe
	exists, err := s.verifyIfUserAlreadyExists(userRequest.Email)
	if err != nil {
		s.Logger.Error("error verifying user existence",
			zap.String("email", userRequest.Email),
			zap.Error(err),
		)
		return errors.NewDatabaseError("user verification failed", nil)
	}

	if exists {
		s.Logger.Info("user already exists",
			zap.String("email", userRequest.Email),
		)
		return errors.NewValidationError("email", "Email already registered")
	}

	// Hash da senha
	hashedPassword, err := utils.HashAString(userRequest.Password)
	if err != nil {
		s.Logger.Error("error hashing password",
			zap.String("email", userRequest.Email),
			zap.Error(err),
		)
		return errors.NewInternalError("password hashing failed", nil)
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
		return errors.NewDatabaseError("user creation failed", nil)
	}

	// Log ANTES de tentar gerar wallet
	s.Logger.Info("user inserted, now attempting wallet generation",
		zap.String("email", userRequest.Email),
		zap.Bool("wallet_manager_is_nil", s.WalletManager == nil),
	)

	// Gerar carteira TRON automaticamente
	if s.WalletManager != nil {
		s.Logger.Info("wallet_manager exists, calling GenerateWallet()",
			zap.String("email", userRequest.Email),
		)

		wallet, err := s.WalletManager.GenerateWallet()
		if err != nil {
			s.Logger.Warn("error generating wallet, user created but without wallet",
				zap.String("email", userRequest.Email),
				zap.Error(err),
			)
		} else {
			// Log detalhado ANTES de salvar
			s.Logger.Info("wallet generated - BEFORE SAVING",
				zap.String("email", userRequest.Email),
				zap.String("user_id", parsedUser.ID.String()),
				zap.String("wallet_address", wallet.Address),
				zap.Int("address_length", len(wallet.Address)),
				zap.String("address_first_char", string(wallet.Address[0])),
			)

			// Salvar a private key criptografada junto com o endereço
			errWallet := s.Database.SaveWalletInfo(parsedUser.ID, wallet.Address, wallet.EncryptedPrivKey)
			if errWallet != nil {
				s.Logger.Warn("error saving wallet info",
					zap.String("email", userRequest.Email),
					zap.String("user_id", parsedUser.ID.String()),
					zap.Error(errWallet),
				)
			} else {
				s.Logger.Info("wallet generated and saved - AFTER SAVING",
					zap.String("email", userRequest.Email),
					zap.String("user_id", parsedUser.ID.String()),
					zap.String("wallet_address", wallet.Address),
				)
			}
		}
	} else {
		s.Logger.Warn("wallet_manager is NIL, skipping wallet generation",
			zap.String("email", userRequest.Email),
		)
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

// GetDatabase retorna a instância do database (para compatibilidade com interface)
func (s *NewUserService) GetDatabase() *repositories.NewDatabase {
	return s.Database
}

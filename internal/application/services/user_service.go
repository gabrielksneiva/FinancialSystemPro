package services

import (
	"context"
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"strings"

	"go.uber.org/zap"
)

type UserService struct {
	Database       DatabasePort // temporário enquanto migramos para repos dedicados
	UserRepo       UserRepositoryPort
	WalletRepo     WalletRepositoryPort
	Logger         *zap.Logger
	WalletManager  entities.WalletManager
	passwordHasher PasswordHasher
	MultiChainSvc  *MultiChainWalletService // novo serviço multi-chain opcional
}

func NewUserService(db DatabasePort, logger *zap.Logger, walletManager entities.WalletManager) *UserService {
	return &UserService{Database: db, UserRepo: NewUserRepositoryAdapter(db), WalletRepo: NewWalletRepositoryAdapter(db), Logger: logger, WalletManager: walletManager, passwordHasher: defaultPasswordHasher{}}
}

// WithPasswordHasher permite definir hasher customizado mantendo construtor antigo.
func (s *UserService) WithPasswordHasher(ph PasswordHasher) *UserService {
	if ph != nil {
		s.passwordHasher = ph
	}
	return s
}

// WithMultiChainWalletService injeta serviço multi-chain (builder pattern)
func (s *UserService) WithMultiChainWalletService(ms *MultiChainWalletService) *UserService {
	s.MultiChainSvc = ms
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

	// Legacy TRON wallet generation (mantém compatibilidade)
	if s.WalletManager != nil {
		wallet, err := s.WalletManager.GenerateWallet()
		if err == nil {
			if s.WalletRepo == nil && s.Database != nil {
				s.WalletRepo = NewWalletRepositoryAdapter(s.Database)
			}
			_ = s.WalletRepo.SaveInfo(context.Background(), parsedUser.ID, wallet.Address, wallet.EncryptedPrivKey)
		}
	}

	// Nova geração multi-chain (TRON) usando registro, se disponível
	if s.MultiChainSvc != nil && s.MultiChainSvc.Registry != nil {
		if _, err := s.MultiChainSvc.GenerateAndPersist(context.Background(), parsedUser.ID, entities.BlockchainTRON); err == nil {
			// Sucesso gera wallet multi-chain TRON; futuras chains podem ser condicionais
		} else if s.Logger != nil {
			s.Logger.Warn("multi-chain wallet generation failed", zap.String("user_id", parsedUser.ID.String()), zap.Error(err))
		}
	}

	return nil
}

func (s *UserService) verifyIfUserAlreadyExists(email string) (bool, error) {
	if s.UserRepo == nil && s.Database != nil { // fallback para instâncias antigas em testes
		s.UserRepo = NewUserRepositoryAdapter(s.Database)
	}
	result, err := s.UserRepo.FindByEmail(context.Background(), email)
	if err != nil && !strings.EqualFold(err.Error(), "record not found") {
		return false, err
	}
	if result != nil {
		return true, nil
	}
	return false, nil
}

func (s *UserService) GetDatabase() DatabasePort { return s.Database }

package service

import (
	"context"
	"financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/contexts/user/domain/repository"
	"financial-system-pro/internal/shared/events"
	"financial-system-pro/internal/shared/utils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// UserService gerencia a lógica de negócio de usuários
type UserService struct {
	userRepo   repository.UserRepository
	walletRepo repository.WalletRepository
	eventBus   events.Bus
	logger     *zap.Logger
}

// NewUserService cria uma nova instância do serviço
func NewUserService(
	userRepo repository.UserRepository,
	walletRepo repository.WalletRepository,
	eventBus events.Bus,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepo:   userRepo,
		walletRepo: walletRepo,
		eventBus:   eventBus,
		logger:     logger,
	}
}

// CreateUser cria um novo usuário com wallet
func (s *UserService) CreateUser(ctx context.Context, email, password string) (*entity.User, error) {
	// Verificar se usuário já existe
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		s.logger.Warn("user already exists", zap.String("email", email))
		return nil, ErrUserAlreadyExists
	}

	// Hash da senha
	hashedPassword, err := utils.HashAString(password)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, ErrPasswordHashFailed
	}

	// Criar entidade
	user := entity.NewUser(email, hashedPassword)

	// Persistir
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	// Publicar evento
	s.eventBus.PublishAsync(ctx, events.UserCreatedEvent{
		BaseEvent: events.BaseEvent{Type: "user.created"},
		UserID:    user.ID,
		Email:     user.Email,
		Name:      "",
		CreatedAt: user.CreatedAt,
	})

	s.logger.Info("user created successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", email),
	)

	return user, nil
}

// Authenticate valida credenciais e retorna o usuário
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil { // adiciona verificação nil para evitar panic
		s.logger.Warn("user not found", zap.String("email", email))
		return nil, ErrInvalidCredentials
	}

	// Verificar senha
	valid, err := utils.HashAndCompareTwoStrings(password, user.Password)
	if err != nil || !valid {
		s.logger.Warn("invalid password", zap.String("email", email))
		return nil, ErrInvalidCredentials
	}

	// Publicar evento
	s.eventBus.PublishAsync(ctx, events.UserAuthenticatedEvent{
		BaseEvent: events.BaseEvent{Type: "user.authenticated"},
		UserID:    user.ID,
		Email:     user.Email,
		IPAddress: "",
		UserAgent: "",
	})

	return user, nil
}

// GetUserWallet retorna a wallet do usuário
func (s *UserService) GetUserWallet(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user wallet",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return wallet, nil
}

// UpdateBalance atualiza o saldo da wallet
func (s *UserService) UpdateBalance(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error {
	balance, _ := amount.Float64()
	return s.walletRepo.UpdateBalance(ctx, userID, balance)
}

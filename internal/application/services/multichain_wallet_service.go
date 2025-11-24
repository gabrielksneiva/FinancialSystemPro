package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	"fmt"

	"github.com/google/uuid"
)

// MultiChainWalletService orquestra geração e persistência de carteiras multi-chain.
type MultiChainWalletService struct {
	Registry   *BlockchainRegistry
	Repo       OnChainWalletRepositoryPort
	Encryption EncryptionProviderPort
}

func NewMultiChainWalletService(reg *BlockchainRegistry, repo OnChainWalletRepositoryPort) *MultiChainWalletService {
	return &MultiChainWalletService{Registry: reg, Repo: repo, Encryption: NoopEncryptionProvider{}}
}

// WithEncryption injeta provider de criptografia (opcional).
func (s *MultiChainWalletService) WithEncryption(p EncryptionProviderPort) *MultiChainWalletService {
	if p != nil {
		s.Encryption = p
	}
	return s
}

// GenerateAndPersist gera uma carteira para usuário e blockchain específica se ainda não existir.
func (s *MultiChainWalletService) GenerateAndPersist(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*entities.GeneratedWallet, error) {
	if s.Registry == nil {
		return nil, fmt.Errorf("registry não configurado")
	}
	if s.Repo == nil {
		return nil, fmt.Errorf("repositório não configurado")
	}
	exists, err := s.Repo.Exists(ctx, userID, chain)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("wallet já existe para chain: %s", chain)
	}
	gw, err := s.Registry.Get(chain)
	if err != nil {
		return nil, err
	}
	gen, err := gw.GenerateWallet(ctx)
	if err != nil {
		return nil, err
	}
	// Placeholder de private key (não exposta pelos gateways atuais)
	plainPriv := "PRIVATE_KEY_PLACEHOLDER"
	encryptedPriv := plainPriv
	if s.Encryption != nil {
		if enc, err := s.Encryption.Encrypt(plainPriv); err == nil {
			encryptedPriv = enc
		} else {
			return nil, fmt.Errorf("falha ao criptografar chave: %w", err)
		}
	}
	if err := s.Repo.Save(ctx, userID, gen, encryptedPriv); err != nil {
		return nil, err
	}
	gen.UserID = userID
	return gen, nil
}

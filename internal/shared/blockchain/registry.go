package blockchain

import (
	"fmt"
	"sync"
)

// DefaultRegistry implementa Registry com thread-safety
type DefaultRegistry struct {
	mu        sync.RWMutex
	providers map[ChainType]Provider
}

// NewRegistry cria um novo registry de providers
func NewRegistry() Registry {
	return &DefaultRegistry{
		providers: make(map[ChainType]Provider),
	}
}

// Register adiciona um provider ao registry
func (r *DefaultRegistry) Register(provider Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	chainType := provider.ChainType()
	if chainType == "" {
		return fmt.Errorf("provider must have a valid chain type")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[chainType]; exists {
		return fmt.Errorf("provider for chain type %s already registered", chainType)
	}

	r.providers[chainType] = provider
	return nil
}

// Get retorna um provider para o chain type especificado
func (r *DefaultRegistry) Get(chainType ChainType) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[chainType]
	if !exists {
		return nil, fmt.Errorf("no provider registered for chain type %s", chainType)
	}

	return provider, nil
}

// List retorna todos os providers registrados
func (r *DefaultRegistry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers
}

// Exists verifica se um provider está registrado
func (r *DefaultRegistry) Exists(chainType ChainType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.providers[chainType]
	return exists
}

// Unregister remove um provider do registry
func (r *DefaultRegistry) Unregister(chainType ChainType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[chainType]; !exists {
		return fmt.Errorf("no provider registered for chain type %s", chainType)
	}

	delete(r.providers, chainType)
	return nil
}

// MustGet retorna um provider ou entra em panic se não existir
func (r *DefaultRegistry) MustGet(chainType ChainType) Provider {
	provider, err := r.Get(chainType)
	if err != nil {
		panic(err)
	}
	return provider
}

// GetAll retorna um mapa de todos os providers por chain type
func (r *DefaultRegistry) GetAll() map[ChainType]Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Retorna uma cópia para evitar modificações externas
	providers := make(map[ChainType]Provider, len(r.providers))
	for chainType, provider := range r.providers {
		providers[chainType] = provider
	}

	return providers
}

// Count retorna o número de providers registrados
func (r *DefaultRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.providers)
}

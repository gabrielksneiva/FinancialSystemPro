package application

import (
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	"fmt"
)

// BlockchainRegistry provê lookup de gateways multi-chain.
// Mantido simples (in-memory) para evolução posterior com carregamento dinâmico de config.
type BlockchainRegistry struct {
	gateways map[entities.BlockchainType]services.BlockchainGatewayPort
}

// NewBlockchainRegistry constrói registro com gateways opcionais.
func NewBlockchainRegistry(gws ...services.BlockchainGatewayPort) *BlockchainRegistry {
	reg := &BlockchainRegistry{gateways: make(map[entities.BlockchainType]services.BlockchainGatewayPort)}
	for _, gw := range gws {
		if gw == nil {
			continue
		}
		reg.gateways[gw.ChainType()] = gw
	}
	return reg
}

// Register adiciona ou substitui gateway para chain.
func (r *BlockchainRegistry) Register(gw services.BlockchainGatewayPort) {
	if r.gateways == nil {
		r.gateways = make(map[entities.BlockchainType]services.BlockchainGatewayPort)
	}
	if gw == nil {
		return
	}
	r.gateways[gw.ChainType()] = gw
}

// Get retorna gateway para uma chain ou erro se ausente.
func (r *BlockchainRegistry) Get(chain entities.BlockchainType) (services.BlockchainGatewayPort, error) {
	gw, ok := r.gateways[chain]
	if !ok || gw == nil {
		return nil, fmt.Errorf("gateway não encontrado para chain: %s", chain)
	}
	return gw, nil
}

// Has verifica se há gateway registrado.
func (r *BlockchainRegistry) Has(chain entities.BlockchainType) bool {
	_, ok := r.gateways[chain]
	return ok
}

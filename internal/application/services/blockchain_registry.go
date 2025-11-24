package services

import (
	"financial-system-pro/internal/domain/entities"
	"fmt"
)

// BlockchainRegistry provê lookup de gateways multi-chain (in-memory).
// Localizado em services para evitar ciclo de import.
type BlockchainRegistry struct {
	gateways map[entities.BlockchainType]BlockchainGatewayPort
}

func NewBlockchainRegistry(gws ...BlockchainGatewayPort) *BlockchainRegistry {
	reg := &BlockchainRegistry{gateways: make(map[entities.BlockchainType]BlockchainGatewayPort)}
	for _, gw := range gws {
		if gw == nil {
			continue
		}
		reg.gateways[gw.ChainType()] = gw
	}
	return reg
}

func (r *BlockchainRegistry) Register(gw BlockchainGatewayPort) {
	if r.gateways == nil {
		r.gateways = make(map[entities.BlockchainType]BlockchainGatewayPort)
	}
	if gw == nil {
		return
	}
	r.gateways[gw.ChainType()] = gw
}

func (r *BlockchainRegistry) Get(chain entities.BlockchainType) (BlockchainGatewayPort, error) {
	gw, ok := r.gateways[chain]
	if !ok || gw == nil {
		return nil, fmt.Errorf("gateway não encontrado para chain: %s", chain)
	}
	return gw, nil
}

func (r *BlockchainRegistry) Has(chain entities.BlockchainType) bool {
	_, ok := r.gateways[chain]
	return ok
}

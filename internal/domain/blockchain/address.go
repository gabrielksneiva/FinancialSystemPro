package blockchain

import (
	"errors"
	"strings"
)

// ChainType enumerates supported blockchains.
type ChainType string

const (
	ChainTRON     ChainType = "tron"
	ChainEthereum ChainType = "ethereum"
	ChainBitcoin  ChainType = "bitcoin"
)

// Address value object with basic validation per chain.
type Address struct {
	value string
	chain ChainType
}

// NewAddress validates format heuristically (real validation should be in adapters).
func NewAddress(chain ChainType, value string) (Address, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Address{}, errors.New("address empty")
	}
	// Minimal heuristics; deeper validation stays in infrastructure adapters.
	minLen := 26
	if chain == ChainBitcoin {
		minLen = 26
	}
	if len(value) < minLen {
		return Address{}, errors.New("address too short")
	}
	return Address{value: value, chain: chain}, nil
}

func (a Address) String() string   { return a.value }
func (a Address) Chain() ChainType { return a.chain }

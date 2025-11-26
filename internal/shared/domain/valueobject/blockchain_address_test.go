package valueobject

import (
	"testing"
)

func TestNewBlockchainAddress(t *testing.T) {
	tests := []struct {
		name           string
		address        string
		blockchainType string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid TRON address",
			address:        "TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC",
			blockchainType: "tron",
			wantErr:        false,
		},
		{
			name:           "valid Ethereum address",
			address:        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			blockchainType: "ethereum",
			wantErr:        false,
		},
		{
			name:           "valid Bitcoin legacy address",
			address:        "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			blockchainType: "bitcoin",
			wantErr:        false,
		},
		{
			name:           "valid Bitcoin bech32 address",
			address:        "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			blockchainType: "bitcoin",
			wantErr:        false,
		},
		{
			name:           "empty address",
			address:        "",
			blockchainType: "tron",
			wantErr:        true,
			errContains:    "address cannot be empty",
		},
		{
			name:           "empty blockchain type",
			address:        "TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC",
			blockchainType: "",
			wantErr:        true,
			errContains:    "blockchain type is required",
		},
		{
			name:           "invalid TRON address format",
			address:        "InvalidTronAddress",
			blockchainType: "tron",
			wantErr:        true,
			errContains:    "invalid TRON address format",
		},
		{
			name:           "invalid Ethereum address format",
			address:        "0xinvalidaddress",
			blockchainType: "ethereum",
			wantErr:        true,
			errContains:    "invalid Ethereum address format",
		},
		{
			name:           "unsupported blockchain",
			address:        "some_address",
			blockchainType: "solana",
			wantErr:        true,
			errContains:    "unsupported blockchain type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewBlockchainAddress(tt.address, tt.blockchainType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewBlockchainAddress() expected error containing '%s', got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewBlockchainAddress() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("NewBlockchainAddress() unexpected error = %v", err)
					return
				}
				if addr.Address() != tt.address {
					t.Errorf("NewBlockchainAddress() address = %v, want %v", addr.Address(), tt.address)
				}
				if addr.BlockchainType() != tt.blockchainType {
					t.Errorf("NewBlockchainAddress() blockchainType = %v, want %v", addr.BlockchainType(), tt.blockchainType)
				}
			}
		})
	}
}

func TestBlockchainAddress_Equals(t *testing.T) {
	tests := []struct {
		name  string
		addr1 BlockchainAddress
		addr2 BlockchainAddress
		want  bool
	}{
		{
			name:  "equal addresses",
			addr1: mustNewBlockchainAddress("TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC", "tron"),
			addr2: mustNewBlockchainAddress("TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC", "tron"),
			want:  true,
		},
		{
			name:  "different addresses same blockchain",
			addr1: mustNewBlockchainAddress("TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC", "tron"),
			addr2: mustNewBlockchainAddress("TTestAddrABCDEFGHJKLMNPQRSTUVWXYZa", "tron"),
			want:  false,
		},
		{
			name:  "different blockchains",
			addr1: mustNewBlockchainAddress("TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC", "tron"),
			addr2: mustNewBlockchainAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "ethereum"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.addr1.Equals(tt.addr2); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainAddress_String(t *testing.T) {
	address := "TJRyWwFs9wTFGZg3JbrVriFbNfCug5tDeC"
	addr := mustNewBlockchainAddress(address, "tron")

	if addr.String() != address {
		t.Errorf("String() = %v, want %v", addr.String(), address)
	}
}

// Helper functions
func mustNewBlockchainAddress(address, blockchainType string) BlockchainAddress {
	addr, err := NewBlockchainAddress(address, blockchainType)
	if err != nil {
		panic(err)
	}
	return addr
}

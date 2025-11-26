package valueobject

import (
	"testing"
)

func TestNewTransactionHash(t *testing.T) {
	tests := []struct {
		name           string
		hash           string
		blockchainType string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid TRON transaction hash",
			hash:           "a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd",
			blockchainType: "tron",
			wantErr:        false,
		},
		{
			name:           "valid Ethereum transaction hash",
			hash:           "0xa1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd",
			blockchainType: "ethereum",
			wantErr:        false,
		},
		{
			name:           "valid Bitcoin transaction hash",
			hash:           "a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd",
			blockchainType: "bitcoin",
			wantErr:        false,
		},
		{
			name:           "empty hash",
			hash:           "",
			blockchainType: "tron",
			wantErr:        true,
			errContains:    "transaction hash cannot be empty",
		},
		{
			name:           "empty blockchain type",
			hash:           "a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd",
			blockchainType: "",
			wantErr:        true,
			errContains:    "blockchain type is required",
		},
		{
			name:           "invalid TRON hash format - too short",
			hash:           "invalidhash",
			blockchainType: "tron",
			wantErr:        true,
			errContains:    "invalid TRON transaction hash format",
		},
		{
			name:           "invalid Ethereum hash format - missing 0x",
			hash:           "a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd",
			blockchainType: "ethereum",
			wantErr:        true,
			errContains:    "invalid Ethereum transaction hash format",
		},
		{
			name:           "invalid hash - contains invalid characters",
			hash:           "g1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd",
			blockchainType: "tron",
			wantErr:        true,
			errContains:    "invalid TRON transaction hash format",
		},
		{
			name:           "unsupported blockchain",
			hash:           "a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd",
			blockchainType: "solana",
			wantErr:        true,
			errContains:    "unsupported blockchain type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txHash, err := NewTransactionHash(tt.hash, tt.blockchainType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTransactionHash() expected error containing '%s', got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewTransactionHash() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("NewTransactionHash() unexpected error = %v", err)
					return
				}
				if txHash.Hash() != tt.hash {
					t.Errorf("NewTransactionHash() hash = %v, want %v", txHash.Hash(), tt.hash)
				}
				if txHash.BlockchainType() != tt.blockchainType {
					t.Errorf("NewTransactionHash() blockchainType = %v, want %v", txHash.BlockchainType(), tt.blockchainType)
				}
			}
		})
	}
}

func TestTransactionHash_Equals(t *testing.T) {
	tests := []struct {
		name  string
		hash1 TransactionHash
		hash2 TransactionHash
		want  bool
	}{
		{
			name:  "equal hashes",
			hash1: mustNewTransactionHash("a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", "tron"),
			hash2: mustNewTransactionHash("a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", "tron"),
			want:  true,
		},
		{
			name:  "different hashes same blockchain",
			hash1: mustNewTransactionHash("a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", "tron"),
			hash2: mustNewTransactionHash("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "tron"),
			want:  false,
		},
		{
			name:  "different blockchains",
			hash1: mustNewTransactionHash("a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", "tron"),
			hash2: mustNewTransactionHash("0xa1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", "ethereum"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hash1.Equals(tt.hash2); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionHash_ShortHash(t *testing.T) {
	tests := []struct {
		name      string
		hash      TransactionHash
		wantShort string
	}{
		{
			name:      "64 character hash",
			hash:      mustNewTransactionHash("a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", "tron"),
			wantShort: "a1b2c3d4...1234abcd",
		},
		{
			name:      "ethereum hash with 0x",
			hash:      mustNewTransactionHash("0xa1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", "ethereum"),
			wantShort: "0xa1b2c3...1234abcd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hash.ShortHash(); got != tt.wantShort {
				t.Errorf("ShortHash() = %v, want %v", got, tt.wantShort)
			}
		})
	}
}

func TestTransactionHash_String(t *testing.T) {
	hash := "a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd"
	txHash := mustNewTransactionHash(hash, "tron")

	if txHash.String() != hash {
		t.Errorf("String() = %v, want %v", txHash.String(), hash)
	}
}

// Helper functions
func mustNewTransactionHash(hash, blockchainType string) TransactionHash {
	txHash, err := NewTransactionHash(hash, blockchainType)
	if err != nil {
		panic(err)
	}
	return txHash
}

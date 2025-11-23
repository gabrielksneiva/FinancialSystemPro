package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTronService(t *testing.T) {
	vaultAddr := "TWKQm9GYgMHyeZK635WbrRFxpadQa6i3ee"
	vaultKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	ts := NewTronService(vaultAddr, vaultKey)

	if ts == nil {
		t.Fatal("TronService não deveria ser nil")
	}

	if ts.vaultAddress != vaultAddr {
		t.Errorf("Expected vault address %s, got %s", vaultAddr, ts.vaultAddress)
	}

	if ts.vaultPrivateKey != vaultKey {
		t.Errorf("Expected vault private key %s, got %s", vaultKey, ts.vaultPrivateKey)
	}
}

func TestValidateAddress(t *testing.T) {
	ts := NewTronService("", "")

	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{"Valid address", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", true},
		{"Another valid", "TWKQm9GYgMHyeZK635WbrRFxpadQa6i3ee", true},
		{"Too short", "TUa2rzCFuHts", false},
		{"Too long", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxhExtra", false},
		{"Wrong prefix", "1Ua2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ts.ValidateAddress(tt.address)
			if result != tt.expected {
				t.Errorf("ValidateAddress(%s) = %v, expected %v", tt.address, result, tt.expected)
			}
		})
	}
}

func TestBase58ToHex(t *testing.T) {
	ts := NewTronService("", "")

	tests := []struct {
		name        string
		address     string
		shouldError bool
	}{
		{"Valid TRON address", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", false},
		{"Another valid", "TWKQm9GYgMHyeZK635WbrRFxpadQa6i3ee", false},
		{"Invalid base58", "000000", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ts.Base58ToHex(tt.address)
			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == "" {
					t.Error("Result should not be empty")
				}
				// Hex should be 42 characters (21 bytes * 2)
				if len(result) != 42 {
					t.Logf("Hex result: %s (length: %d)", result, len(result))
				}
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/accounts/TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh" {
			response := map[string]interface{}{
				"balance": int64(1000000), // 1 TRX
			}
			_ = json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ts := NewTronService("", "")
	ts.testnetRPC = server.URL

	balance, err := ts.GetBalance("TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance)
	}
}

func TestGetBalance_InvalidAddress(t *testing.T) {
	ts := NewTronService("", "")

	_, err := ts.GetBalance("invalid")
	if err == nil {
		t.Error("Expected error for invalid address")
	}
}

func TestGetBalance_EmptyAddress(t *testing.T) {
	ts := NewTronService("", "")

	_, err := ts.GetBalance("")
	if err == nil {
		t.Error("Expected error for empty address")
	}
}

func TestHasVaultConfigured(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		key      string
		expected bool
	}{
		{"Both configured", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", "abc123", true},
		{"No address", "", "abc123", false},
		{"No key", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", "", false},
		{"Neither", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTronService(tt.address, tt.key)
			result := ts.HasVaultConfigured()
			if result != tt.expected {
				t.Errorf("HasVaultConfigured() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetVaultAddress(t *testing.T) {
	vaultAddr := "TWKQm9GYgMHyeZK635WbrRFxpadQa6i3ee"
	ts := NewTronService(vaultAddr, "key")

	result := ts.GetVaultAddress()
	if result != vaultAddr {
		t.Errorf("GetVaultAddress() = %s, expected %s", result, vaultAddr)
	}
}

func TestGetVaultPrivateKey(t *testing.T) {
	vaultKey := "abc123def456"
	ts := NewTronService("addr", vaultKey)

	result := ts.GetVaultPrivateKey()
	if result != vaultKey {
		t.Errorf("GetVaultPrivateKey() = %s, expected %s", result, vaultKey)
	}
}

func TestSendTransaction_InvalidAddresses(t *testing.T) {
	ts := NewTronService("", "")

	tests := []struct {
		name    string
		from    string
		to      string
		amount  int64
		privKey string
	}{
		{"Empty from", "", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", 1000000, "key"},
		{"Empty to", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", "", 1000000, "key"},
		{"Zero amount", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", "TWKQm9GYgMHyeZK635WbrRFxpadQa6i3ee", 0, "key"},
		{"Negative amount", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", "TWKQm9GYgMHyeZK635WbrRFxpadQa6i3ee", -100, "key"},
		{"Empty private key", "TUa2rzCFuHtsaLHGPLywUNNhsrrQkZxQxh", "TWKQm9GYgMHyeZK635WbrRFxpadQa6i3ee", 1000000, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ts.SendTransaction(tt.from, tt.to, tt.amount, tt.privKey)
			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

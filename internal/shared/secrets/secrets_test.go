package secrets

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalSecretManagerStore(t *testing.T) {
	manager := NewLocalSecretManager()

	err := manager.Store("test_key", "test_value")
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}
}

func TestLocalSecretManagerRetrieve(t *testing.T) {
	manager := NewLocalSecretManager()

	// Store a value
	err := manager.Store("retrieve_key", "retrieve_value")
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}

	// Retrieve it
	value, err := manager.Retrieve("retrieve_key")
	if err != nil {
		t.Errorf("Retrieve failed: %v", err)
	}

	if value != "retrieve_value" {
		t.Errorf("Expected 'retrieve_value', got '%s'", value)
	}
}

func TestLocalSecretManagerRetrieveNotFound(t *testing.T) {
	manager := NewLocalSecretManager()

	_, err := manager.Retrieve("nonexistent_key")
	if err != ErrSecretNotFound {
		t.Errorf("Expected ErrSecretNotFound, got %v", err)
	}
}

func TestLocalSecretManagerDelete(t *testing.T) {
	manager := NewLocalSecretManager()

	// Store a value
	_ = manager.Store("delete_key", "delete_value")

	// Delete it
	err := manager.Delete("delete_key")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify it's deleted
	_, err = manager.Retrieve("delete_key")
	if err != ErrSecretNotFound {
		t.Errorf("Expected ErrSecretNotFound after deletion, got %v", err)
	}
}

func TestLocalSecretManagerExists(t *testing.T) {
	manager := NewLocalSecretManager()

	// Store a value
	_ = manager.Store("exists_key", "exists_value")

	// Check it exists
	if !manager.Exists("exists_key") {
		t.Error("Expected key to exist")
	}

	// Check nonexistent key
	if manager.Exists("nonexistent_key") {
		t.Error("Expected key to not exist")
	}
}

func TestSecretErrorCode(t *testing.T) {
	err := ErrSecretNotFound
	if err.Code() != "SECRET_NOT_FOUND" {
		t.Errorf("Expected code 'SECRET_NOT_FOUND', got '%s'", err.Code())
	}
}

func TestSecretErrorMessage(t *testing.T) {
	// Note: Can't create SecretError with struct literal since fields are private
	// Testing through ErrSecretNotFound instead
	err := ErrSecretNotFound
	if err.Error() != "secret not found" {
		t.Errorf("Expected 'secret not found', got '%s'", err.Error())
	}
}

// VaultSecretManager Tests with Mock HTTP Server

func TestVaultSecretManagerStore(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("X-Vault-Token") != "test-token" {
			t.Errorf("Expected X-Vault-Token header")
		}

		// Decode request body
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)

		data, ok := payload["data"].(map[string]interface{})
		if !ok || data["value"] != "test_value" {
			t.Errorf("Invalid payload format")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := NewVaultSecretManager(server.URL, "test-token", "secret/data")
	err := manager.Store("test_key", "test_value")
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}
}

func TestVaultSecretManagerRetrieve(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("X-Vault-Token") != "test-token" {
			t.Errorf("Expected X-Vault-Token header")
		}

		// Return Vault KV v2 response format
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]string{
					"value": "retrieved_value",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	manager := NewVaultSecretManager(server.URL, "test-token", "secret/data")
	value, err := manager.Retrieve("test_key")
	if err != nil {
		t.Errorf("Retrieve failed: %v", err)
	}
	if value != "retrieved_value" {
		t.Errorf("Expected 'retrieved_value', got '%s'", value)
	}
}

func TestVaultSecretManagerRetrieveNotFound(t *testing.T) {
	// Create mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	manager := NewVaultSecretManager(server.URL, "test-token", "secret/data")
	_, err := manager.Retrieve("nonexistent_key")
	if err != ErrSecretNotFound {
		t.Errorf("Expected ErrSecretNotFound, got %v", err)
	}
}

func TestVaultSecretManagerDelete(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.Header.Get("X-Vault-Token") != "test-token" {
			t.Errorf("Expected X-Vault-Token header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	manager := NewVaultSecretManager(server.URL, "test-token", "secret/data")
	err := manager.Delete("test_key")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

func TestVaultSecretManagerExists(t *testing.T) {
	// Create mock server that simulates key existence check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/secret/data/existing_key" {
			// Return success for existing key
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]string{
						"value": "some_value",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		} else {
			// Return 404 for non-existing key
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	manager := NewVaultSecretManager(server.URL, "test-token", "secret/data")

	// Test existing key
	if !manager.Exists("existing_key") {
		t.Error("Expected key to exist")
	}

	// Test non-existing key
	if manager.Exists("nonexistent_key") {
		t.Error("Expected key to not exist")
	}
}

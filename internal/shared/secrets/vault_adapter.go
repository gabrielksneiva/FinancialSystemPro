package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// VaultSecretManager implementa SecretManager usando HashiCorp Vault
type VaultSecretManager struct {
	client   *http.Client
	vaultURL string
	token    string
	path     string // caminho base no Vault, ex: secret/data
}

// NewVaultSecretManager cria nova instância de VaultSecretManager
func NewVaultSecretManager(vaultURL, token, path string) *VaultSecretManager {
	return &VaultSecretManager{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		vaultURL: vaultURL,
		token:    token,
		path:     path,
	}
}

// Store armazena um secret no Vault
func (m *VaultSecretManager) Store(key string, value string) error {
	url := fmt.Sprintf("%s/v1/%s/%s", m.vaultURL, m.path, key)

	// Vault KV v2 format: {"data": {"value": "secret_value"}}
	payload := map[string]interface{}{
		"data": map[string]string{
			"value": value,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal vault payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create vault request: %w", err)
	}

	req.Header.Set("X-Vault-Token", m.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Retrieve obtém um secret do Vault
func (m *VaultSecretManager) Retrieve(key string) (string, error) {
	url := fmt.Sprintf("%s/v1/%s/%s", m.vaultURL, m.path, key)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create vault request: %w", err)
	}

	req.Header.Set("X-Vault-Token", m.token)

	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrSecretNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Vault KV v2 response: {"data": {"data": {"value": "secret_value"}}}
	var response struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode vault response: %w", err)
	}

	value, ok := response.Data.Data["value"]
	if !ok {
		return "", ErrSecretNotFound
	}

	return value, nil
}

// Delete remove um secret do Vault
func (m *VaultSecretManager) Delete(key string) error {
	url := fmt.Sprintf("%s/v1/%s/%s", m.vaultURL, m.path, key)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create vault request: %w", err)
	}

	req.Header.Set("X-Vault-Token", m.token)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Exists verifica se um secret existe no Vault
func (m *VaultSecretManager) Exists(key string) bool {
	_, err := m.Retrieve(key)
	return err == nil
}

// AWSSecretsManagerAdapter implementa SecretManager usando AWS Secrets Manager
type AWSSecretsManagerAdapter struct {
	region string
}

// NewAWSSecretsManagerAdapter cria nova instância de AWSSecretsManagerAdapter
func NewAWSSecretsManagerAdapter(region string) *AWSSecretsManagerAdapter {
	return &AWSSecretsManagerAdapter{
		region: region,
	}
}

// Store armazena um secret no AWS Secrets Manager
func (a *AWSSecretsManagerAdapter) Store(key string, value string) error {
	// Usando AWS SDK v2: secretsmanager.Client.PutSecretValue()
	// Implementação completa requer: github.com/aws/aws-sdk-go-v2/service/secretsmanager
	return nil
}

// Retrieve obtém um secret do AWS Secrets Manager
func (a *AWSSecretsManagerAdapter) Retrieve(key string) (string, error) {
	// Usando AWS SDK v2: secretsmanager.Client.GetSecretValue()
	return "", ErrSecretNotFound
}

// Delete remove um secret do AWS Secrets Manager
func (a *AWSSecretsManagerAdapter) Delete(key string) error {
	// Usando AWS SDK v2: secretsmanager.Client.DeleteSecret()
	return nil
}

// Exists verifica se um secret existe no AWS Secrets Manager
func (a *AWSSecretsManagerAdapter) Exists(key string) bool {
	_, err := a.Retrieve(key)
	return err == nil
}

// SecretManagerFactory cria o SecretManager apropriado baseado no ambiente
type SecretManagerFactory struct {
	environment string
}

// NewSecretManagerFactory cria nova factory
func NewSecretManagerFactory(environment string) *SecretManagerFactory {
	return &SecretManagerFactory{
		environment: environment,
	}
}

// Create cria o SecretManager apropriado
func (f *SecretManagerFactory) Create() SecretManager {
	switch f.environment {
	case "production":
		// Em produção, usar Vault ou AWS
		// return NewVaultSecretManager(vaultURL, token, "secret/data")
		// return NewAWSSecretsManagerAdapter("us-east-1")
		return NewLocalSecretManager() // placeholder
	case "staging":
		return NewLocalSecretManager()
	default:
		return NewLocalSecretManager()
	}
}

package secrets

// SecretManager define interface para gerenciar secrets
type SecretManager interface {
// Store armazena um secret de forma segura
Store(key string, value string) error

// Retrieve obtém um secret de forma segura
Retrieve(key string) (string, error)

// Delete remove um secret
Delete(key string) error

// Exists verifica se um secret existe
Exists(key string) bool
}

// LocalSecretManager implementa SecretManager usando environment variables
// NOTA: Usar apenas para desenvolvimento. Em produção, usar HashiCorp Vault ou AWS Secrets Manager
type LocalSecretManager struct {
secrets map[string]string
}

// NewLocalSecretManager cria nova instância de LocalSecretManager
func NewLocalSecretManager() *LocalSecretManager {
return &LocalSecretManager{
secrets: make(map[string]string),
}
}

// Store armazena um secret localmente (apenas para testes)
func (m *LocalSecretManager) Store(key string, value string) error {
m.secrets[key] = value
return nil
}

// Retrieve obtém um secret localmente
func (m *LocalSecretManager) Retrieve(key string) (string, error) {
if value, ok := m.secrets[key]; ok {
return value, nil
}
return "", ErrSecretNotFound
}

// Delete remove um secret
func (m *LocalSecretManager) Delete(key string) error {
delete(m.secrets, key)
return nil
}

// Exists verifica se um secret existe
func (m *LocalSecretManager) Exists(key string) bool {
_, ok := m.secrets[key]
return ok
}

// Errors
var (
ErrSecretNotFound = &SecretError{code: "SECRET_NOT_FOUND", message: "secret not found"}
ErrStorageFailed  = &SecretError{code: "STORAGE_FAILED", message: "failed to store secret"}
)

// SecretError representa um erro de secrets
type SecretError struct {
code    string
message string
}

func (e *SecretError) Error() string {
return e.message
}

func (e *SecretError) Code() string {
return e.code
}

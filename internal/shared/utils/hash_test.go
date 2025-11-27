package utils

import (
	"os"
	"testing"

	"github.com/golang-jwt/jwt"
)

// TestMain configures environment before tests
func TestMain(m *testing.M) {
	// Set secret key for testing
	os.Setenv("SECRET_KEY", "test-secret-key-for-jwt-and-hmac")
	os.Setenv("EXPIRATION_TIME", "3600") // 1 hour
	code := m.Run()
	os.Exit(code)
}

// TestHashAString tests HMAC string hashing
func TestHashAString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple text", "hello"},
		{"numbers", "123456"},
		{"special chars", "!@#$%"},
		{"unicode", "你好"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashAString(tt.input)
			if err != nil {
				t.Fatalf("HashAString failed: %v", err)
			}

			if hash == "" && tt.input != "" {
				t.Error("Hash should not be empty for non-empty input")
			}

			if hash == tt.input {
				t.Error("Hash should be different from input")
			}
		})
	}
}

// TestHashAStringConsistency tests that same input produces same hash
func TestHashAStringConsistency(t *testing.T) {
	input := "test data"

	hash1, err := HashAString(input)
	if err != nil {
		t.Fatalf("First hash failed: %v", err)
	}

	hash2, err := HashAString(input)
	if err != nil {
		t.Fatalf("Second hash failed: %v", err)
	}

	if hash1 != hash2 {
		t.Error("HMAC should produce same hash for same input")
	}
}

// TestHashAndCompareTwoStrings tests string comparison with hashing
func TestHashAndCompareTwoStrings(t *testing.T) {
	original := "password123"

	// Hash the original
	hashed, err := HashAString(original)
	if err != nil {
		t.Fatalf("HashAString failed: %v", err)
	}

	// Compare with correct hash
	match, err := HashAndCompareTwoStrings(original, hashed)
	if err != nil {
		t.Fatalf("HashAndCompareTwoStrings failed: %v", err)
	}
	if !match {
		t.Error("Should match for correct string")
	}

	// Compare with wrong hash
	wrongMatch, err := HashAndCompareTwoStrings("wrongpassword", hashed)
	if err != nil {
		t.Fatalf("HashAndCompareTwoStrings failed: %v", err)
	}
	if wrongMatch {
		t.Error("Should not match for different string")
	}
}

// TestCreateJWTToken tests JWT token creation
func TestCreateJWTToken(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id": "123",
		"email":   "test@example.com",
	}

	token, err := CreateJWTToken(claims)
	if err != nil {
		t.Fatalf("CreateJWTToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Token should have 3 parts separated by dots
	parts := 0
	for _, char := range token {
		if char == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("JWT should have 3 parts (2 dots), got %d dots", parts)
	}
}

// TestDecodeJWTToken tests JWT token decoding
func TestDecodeJWTToken(t *testing.T) {
	// Create a token
	claims := jwt.MapClaims{
		"user_id": "456",
		"role":    "admin",
	}

	tokenString, err := CreateJWTToken(claims)
	if err != nil {
		t.Fatalf("CreateJWTToken failed: %v", err)
	}

	// Decode the token
	token, err := DecodeJWTToken(tokenString)
	if err != nil {
		t.Fatalf("DecodeJWTToken failed: %v", err)
	}

	if !token.Valid {
		t.Error("Token should be valid")
	}

	// Verify claims
	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}

	if tokenClaims["user_id"] != "456" {
		t.Errorf("user_id mismatch: got %v, want %v", tokenClaims["user_id"], "456")
	}

	if tokenClaims["role"] != "admin" {
		t.Errorf("role mismatch: got %v, want %v", tokenClaims["role"], "admin")
	}
}

// TestDecodeJWTTokenInvalid tests decoding invalid tokens
func TestDecodeJWTTokenInvalid(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"invalid format", "invalid.token.here"},
		{"wrong signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIn0.wrong_signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := DecodeJWTToken(tt.token)
			if err == nil && (token == nil || token.Valid) {
				t.Error("Expected error or invalid token for invalid input")
			}
		})
	}
}

// TestJWTRoundTrip tests creating and decoding a token
func TestJWTRoundTrip(t *testing.T) {
	originalClaims := jwt.MapClaims{
		"user_id": "789",
		"email":   "roundtrip@test.com",
		"active":  true,
	}

	// Create token
	tokenString, err := CreateJWTToken(originalClaims)
	if err != nil {
		t.Fatalf("CreateJWTToken failed: %v", err)
	}

	// Decode token
	token, err := DecodeJWTToken(tokenString)
	if err != nil {
		t.Fatalf("DecodeJWTToken failed: %v", err)
	}

	// Verify claims
	decodedClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}

	if decodedClaims["user_id"] != originalClaims["user_id"] {
		t.Error("user_id mismatch after round trip")
	}

	if decodedClaims["email"] != originalClaims["email"] {
		t.Error("email mismatch after round trip")
	}

	if decodedClaims["active"] != originalClaims["active"] {
		t.Error("active mismatch after round trip")
	}
}

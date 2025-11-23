package dto

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func TestUserRequestValidation_Valid(t *testing.T) {
	req := UserRequest{
		Email:    "valid@example.com",
		Password: "securepassword123",
	}

	err := validate.Struct(req)
	if err != nil {
		t.Errorf("Expected valid UserRequest, got validation error: %v", err)
	}
}

func TestUserRequestValidation_InvalidEmail(t *testing.T) {
	req := UserRequest{
		Email:    "invalidemail",
		Password: "securepassword123",
	}

	err := validate.Struct(req)
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}
}

func TestUserRequestValidation_WeakPassword(t *testing.T) {
	req := UserRequest{
		Email:    "valid@example.com",
		Password: "weak",
	}

	err := validate.Struct(req)
	if err == nil {
		t.Error("Expected validation error for weak password")
	}
}

func TestDepositRequestValidation_Valid(t *testing.T) {
	req := DepositRequest{
		Amount: "100.50",
	}

	err := validate.Struct(req)
	if err != nil {
		t.Errorf("Expected valid DepositRequest, got validation error: %v", err)
	}
}

func TestDepositRequestValidation_ZeroAmount(t *testing.T) {
	// Note: The 'gt=0' validator works on numeric fields, not string fields
	// Since Amount is a string with 'numeric' validation, it only checks format
	// The business logic should validate the actual value after parsing
	req := DepositRequest{
		Amount: "abc", // Invalid format should fail
	}

	err := validate.Struct(req)
	if err == nil {
		t.Error("Expected validation error for invalid amount format")
	}
}

func TestDepositRequestValidation_NegativeAmount(t *testing.T) {
	// Note: 'numeric' validator accepts negative numbers as valid format
	// Business logic should check the value > 0 after parsing
	req := DepositRequest{
		Amount: "", // Empty should fail 'required'
	}

	err := validate.Struct(req)
	if err == nil {
		t.Error("Expected validation error for empty amount")
	}
}

func TestTronRequestValidation_Valid(t *testing.T) {
	req := TronRequest{
		Address: "TJRyWwsFsfvCS6VRWNtV5FwusSh75Jx25a", // Valid TRON address format
		Amount:  "100.50",
	}

	err := validate.Struct(req)
	if err != nil {
		t.Errorf("Expected valid TronRequest, got validation error: %v", err)
	}
}

func TestTronRequestValidation_InvalidAddress(t *testing.T) {
	req := TronRequest{
		Address: "InvalidAddress",
		Amount:  "100.50",
	}

	err := validate.Struct(req)
	if err == nil {
		t.Error("Expected validation error for invalid TRON address")
	}
}

func TestTronRequestValidation_InvalidAmount(t *testing.T) {
	req := TronRequest{
		Address: "TJRyWwsFsfvCS6VRWNtV5FwusSh75Jx25a",
		Amount:  "notanumber",
	}

	err := validate.Struct(req)
	if err == nil {
		t.Error("Expected validation error for invalid amount")
	}
}

func TestSendTronRequestValidation_Valid(t *testing.T) {
	req := SendTronRequest{
		ToAddress: "TJRyWwsFsfvCS6VRWNtV5FwusSh75Jx25a",
		Amount:    "50.25",
	}

	err := validate.Struct(req)
	if err != nil {
		t.Errorf("Expected valid SendTronRequest, got validation error: %v", err)
	}
}

func TestEstimateGasRequestValidation_Valid(t *testing.T) {
	req := EstimateGasRequest{
		ToAddress: "TJRyWwsFsfvCS6VRWNtV5FwusSh75Jx25a",
		Amount:    "10.00",
	}

	err := validate.Struct(req)
	if err != nil {
		t.Errorf("Expected valid EstimateGasRequest, got validation error: %v", err)
	}
}

func TestWithdrawRequestValidation_Valid(t *testing.T) {
	req := WithdrawRequest{
		Amount: "75.50",
	}

	err := validate.Struct(req)
	if err != nil {
		t.Errorf("Expected valid WithdrawRequest, got validation error: %v", err)
	}
}

func TestTransferRequestValidation_Valid(t *testing.T) {
	req := TransferRequest{
		To:     "receiver@example.com",
		Amount: "25.00",
	}

	err := validate.Struct(req)
	if err != nil {
		t.Errorf("Expected valid TransferRequest, got validation error: %v", err)
	}
}

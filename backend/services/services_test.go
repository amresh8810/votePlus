package services_test

import (
	"testing"
	"time"

	"github.com/live-polling-app/backend/services"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	rawPassword := "SuperSecretPassword123!"

	hash, err := services.HashPassword(rawPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == rawPassword {
		t.Fatalf("password hash must not match raw password")
	}

	// Correct password check
	if !services.CheckPassword(hash, rawPassword) {
		t.Errorf("expected password check to succeed for correct password")
	}

	// Incorrect password check
	if services.CheckPassword(hash, "WrongPassword!") {
		t.Errorf("expected password check to fail for incorrect password")
	}
}

func TestJWTGenerationAndValidation(t *testing.T) {
	userID := "650000000000000000000001"
	secret := "test_secret_key_for_unit_testing_32b"
	expiration := 1 * time.Hour

	tokenStr, err := services.GenerateToken(userID, secret, expiration)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := services.ValidateToken(tokenStr, secret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("expected subject '%s', got '%s'", userID, claims.Subject)
	}
}

func TestExpiredJWTValidationFails(t *testing.T) {
	userID := "650000000000000000000001"
	secret := "test_secret_key_for_unit_testing_32b"
	// Negative duration means token is already expired when created
	expiration := -1 * time.Second

	tokenStr, err := services.GenerateToken(userID, secret, expiration)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	_, err = services.ValidateToken(tokenStr, secret)
	if err == nil {
		t.Errorf("expected validation error for expired token, got nil")
	}
}

func TestInvalidSecretJWTValidationFails(t *testing.T) {
	userID := "650000000000000000000001"
	tokenStr, err := services.GenerateToken(userID, "correct_secret_key_32bytes_long", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = services.ValidateToken(tokenStr, "wrong_secret_key_32bytes_long___")
	if err == nil {
		t.Errorf("expected validation error for wrong secret, got nil")
	}
}

func TestSignupValidation(t *testing.T) {
	// Valid request
	valid := services.SignupRequest{Name: "Alice", Email: "Alice@Example.COM  ", Password: "securepassword"}
	if err := services.ValidateSignup(valid); err != nil {
		t.Errorf("expected valid signup request to pass, got error: %v", err)
	}

	// Missing name
	missingName := services.SignupRequest{Name: "", Email: "alice@example.com", Password: "securepassword"}
	if err := services.ValidateSignup(missingName); err == nil {
		t.Errorf("expected error for missing name")
	}

	// Invalid email
	invalidEmail := services.SignupRequest{Name: "Alice", Email: "invalid-email-string", Password: "securepassword"}
	if err := services.ValidateSignup(invalidEmail); err == nil {
		t.Errorf("expected error for invalid email format")
	}

	// Short password
	shortPass := services.SignupRequest{Name: "Alice", Email: "alice@example.com", Password: "123"}
	if err := services.ValidateSignup(shortPass); err == nil {
		t.Errorf("expected error for password shorter than 6 characters")
	}
}

func TestEmailNormalization(t *testing.T) {
	raw := "   User.Name+Tag@Domain.COM   "
	expected := "user.name+tag@domain.com"
	if normalized := services.NormalizeEmail(raw); normalized != expected {
		t.Errorf("expected '%s', got '%s'", expected, normalized)
	}
}

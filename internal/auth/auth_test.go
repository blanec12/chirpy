package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	// Generate JWT
	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() failed: %v", err)
	}

	// Validate JWT
	validatedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() failed: %v", err)
	}

	if validatedID != userID {
		t.Errorf("expected userID %s, got %s", userID, validatedID)
	}
}

func TestValidateExpiredJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() failed: %v", err)
	}

	_, err = ValidateJWT(token, secret)

	if err == nil {
		t.Fatal("expected an error when validating an expired JWT")
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, "correct-secret", -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() failed: %v", err)
	}

	_, err = ValidateJWT(token, "wrong-secret")

	if err == nil {
		t.Fatal("expected an error when validating JWT with wrong secret")
	}
}

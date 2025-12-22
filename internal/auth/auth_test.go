package auth_test

import (
	"chirpy/internal/auth"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "securepassword123"

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	match, err := auth.CheckPassword(password, hashedPassword)
	if err != nil {
		t.Fatalf("Error checking password: %v", err)
	}

	if !match {
		t.Fatalf("Password does not match hashed password")
	}
}

package utils

import (
	"errors"
	"testing"
)

func TestHashPasswordAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == "secret123" {
		t.Fatal("HashPassword should not return plain text password")
	}

	if !CheckPassword("secret123", hash) {
		t.Fatal("CheckPassword should accept the correct password")
	}

	if CheckPassword("wrong-password", hash) {
		t.Fatal("CheckPassword should reject an incorrect password")
	}
}

func TestHashPasswordRejectsLongPasswords(t *testing.T) {
	longPassword := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	if len([]byte(longPassword)) <= maxBcryptPasswordBytes {
		t.Fatal("test password must exceed 72 bytes")
	}

	hash, err := HashPassword(longPassword)
	if err == nil {
		t.Fatal("HashPassword should reject passwords longer than 72 bytes")
	}
	if !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("HashPassword should return ErrPasswordTooLong, got %v", err)
	}

	if hash != "" {
		t.Fatal("HashPassword should return an empty string when hashing fails")
	}
}

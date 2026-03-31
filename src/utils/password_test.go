package utils

import "testing"

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

/*
 * Copyright (c) 2026 KAnggara
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * See <https://www.gnu.org/licenses/gpl-3.0.html>.
 *
 * @author KAnggara on Saturday 08/02/2026 10.57
 * @project pp
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/security/password
 */

package password

import (
	"strings"
	"testing"
)

func TestHash_Success(t *testing.T) {
	plainPassword := "mySecurePassword123"

	hashed, err := Hash(plainPassword)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hashed == "" {
		t.Error("Expected non-empty hash")
	}

	if hashed == plainPassword {
		t.Error("Hash should not equal plain password")
	}

	// Bcrypt hashes start with $2a$ or $2b$
	if !strings.HasPrefix(hashed, "$2a$") && !strings.HasPrefix(hashed, "$2b$") {
		t.Errorf("Expected bcrypt hash format, got %s", hashed)
	}
}

func TestHash_DifferentHashesForSamePassword(t *testing.T) {
	plainPassword := "mySecurePassword123"

	hash1, err1 := Hash(plainPassword)
	if err1 != nil {
		t.Fatalf("Expected no error for first hash, got %v", err1)
	}

	hash2, err2 := Hash(plainPassword)
	if err2 != nil {
		t.Fatalf("Expected no error for second hash, got %v", err2)
	}

	// Bcrypt should generate different hashes for the same password (due to salt)
	if hash1 == hash2 {
		t.Error("Expected different hashes for same password (bcrypt uses salt)")
	}
}

func TestHash_EmptyPassword(t *testing.T) {
	hashed, err := Hash("")
	if err != nil {
		t.Fatalf("Expected no error for empty password, got %v", err)
	}

	if hashed == "" {
		t.Error("Expected non-empty hash even for empty password")
	}
}

func TestHash_LongPassword(t *testing.T) {
	// Bcrypt has a maximum password length of 72 bytes
	// Passwords within 72 bytes should work fine
	password72 := strings.Repeat("a", 72)

	hashed, err := Hash(password72)
	if err != nil {
		t.Fatalf("Expected no error for 72-byte password, got %v", err)
	}

	if hashed == "" {
		t.Error("Expected non-empty hash")
	}

	// Passwords longer than 72 bytes will be truncated by bcrypt
	// This is a known limitation, but the function should still work
	longPassword := strings.Repeat("a", 100)
	hashed2, err2 := Hash(longPassword)
	if err2 != nil {
		// Some bcrypt implementations may error on >72 bytes
		// This is acceptable behavior
		return
	}

	if hashed2 == "" {
		t.Error("Expected non-empty hash")
	}
}

func TestCompare_Success(t *testing.T) {
	plainPassword := "mySecurePassword123"

	hashed, err := Hash(plainPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	result := Compare(hashed, plainPassword)
	if !result {
		t.Error("Expected Compare to return true for correct password")
	}
}

func TestCompare_WrongPassword(t *testing.T) {
	plainPassword := "mySecurePassword123"
	wrongPassword := "wrongPassword456"

	hashed, err := Hash(plainPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	result := Compare(hashed, wrongPassword)
	if result {
		t.Error("Expected Compare to return false for wrong password")
	}
}

func TestCompare_EmptyPassword(t *testing.T) {
	plainPassword := "mySecurePassword123"

	hashed, err := Hash(plainPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	result := Compare(hashed, "")
	if result {
		t.Error("Expected Compare to return false for empty password")
	}
}

func TestCompare_InvalidHash(t *testing.T) {
	result := Compare("invalid-hash", "password")
	if result {
		t.Error("Expected Compare to return false for invalid hash")
	}
}

func TestCompare_EmptyHash(t *testing.T) {
	result := Compare("", "password")
	if result {
		t.Error("Expected Compare to return false for empty hash")
	}
}

func TestCompare_CaseSensitive(t *testing.T) {
	plainPassword := "MyPassword123"

	hashed, err := Hash(plainPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test with different case
	result := Compare(hashed, "mypassword123")
	if result {
		t.Error("Expected Compare to be case-sensitive")
	}

	// Test with correct case
	result = Compare(hashed, plainPassword)
	if !result {
		t.Error("Expected Compare to return true for correct password")
	}
}

func TestHashAndCompare_MultiplePasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple", "password"},
		{"with numbers", "pass123"},
		{"with special chars", "p@ssw0rd!"},
		{"unicode", "пароль123"},
		{"spaces", "my password"},
		{"long", strings.Repeat("a", 60)}, // Use 60 bytes to avoid bcrypt's 72-byte truncation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := Hash(tt.password)
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}

			if !Compare(hashed, tt.password) {
				t.Error("Expected Compare to return true for correct password")
			}

			if Compare(hashed, tt.password+"wrong") {
				t.Error("Expected Compare to return false for wrong password")
			}
		})
	}
}

func TestHash_Consistency(t *testing.T) {
	plainPassword := "testPassword123"

	hashed, err := Hash(plainPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify the same hash can be used multiple times
	for i := 0; i < 5; i++ {
		if !Compare(hashed, plainPassword) {
			t.Errorf("Iteration %d: Expected Compare to return true", i)
		}
	}
}

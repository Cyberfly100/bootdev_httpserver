package auth

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestComplexPasswordHashing(t *testing.T) {
	password := "mYs3cret!p@ssword##"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("Error comparing password and hash: %v", err)
	}
	if !match {
		t.Errorf("Expected password to match hash, but it did not")
	}
}

func TestWrongPasswordHashing(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	wrongPassword := "wrongpassword"
	match, err := CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Fatalf("Error comparing wrong password and hash: %v", err)
	}
	if match {
		t.Errorf("Expected wrong password to not match hash, but it did")
	}
}

func TestHashVariability(t *testing.T) {
	// Test that hashing the same password multiple times produces different hashes
	password := "mysecretpassword"
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password a second time: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("Expected different hashes for the same password, but got the same")
	}
}

func TestHashingEmptyPassword(t *testing.T) {
	// Test that the function handles empty passwords
	emptyPassword := ""
	hashEmpty, err := HashPassword(emptyPassword)
	if err != nil {
		t.Fatalf("Failed to hash empty password: %v", err)
	}
	match, err := CheckPasswordHash(emptyPassword, hashEmpty)
	if err != nil {
		t.Fatalf("Error comparing empty password and hash: %v", err)
	}
	if !match {
		t.Errorf("Expected empty password to match its hash, but it did not")
	}
}

func TestHashingLongPassword(t *testing.T) {
	// Test that the function handles long passwords
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	longPassword := make([]rune, 1000)
	for i := range 1000 {
		longPassword[i] = letters[rand.Intn(len(letters))]
	}
	hashLong, err := HashPassword(string(longPassword))
	if err != nil {
		t.Fatalf("Failed to hash long password: %v", err)
	}
	match, err := CheckPasswordHash(string(longPassword), hashLong)
	if err != nil {
		t.Fatalf("Error comparing long password and hash: %v", err)
	}
	if !match {
		t.Errorf("Expected long password to match its hash, but it did not")
	}
}

func TestJWTValidation(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecretkey"
	token, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	validatedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}
	if validatedUserID != userID {
		t.Errorf("Expected validated user ID to match original user ID, but it did not")
	}
}

func TestJWTValidationWithWrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecretkey"
	token, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	wrongSecret := "wrongsecretkey"
	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Errorf("Expected validation to fail with wrong secret, but it succeeded")
	}
}

func TestJWTValidationWithExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecretkey"
	token, err := MakeJWT(userID, tokenSecret, -time.Hour) // Token that expired an hour ago
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Errorf("Expected validation to fail with expired token, but it succeeded")
	}
}

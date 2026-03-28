package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using Argon2
func (a *AuthService) HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Use configurable time and memory costs with secure defaults
	// OWASP recommends timeCost >= 3 for sensitive applications (previously was 1)
	timeCost := uint32(getTimeCost())
	memoryCost := uint32(getMemoryCost())

	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, keyLength)

	// Encode salt and hash as base64
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	// Note: timeCost now defaults to 3 (OWASP recommendation) instead of 1
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memoryCost, timeCost, threads, saltB64, hashB64), nil
}

// VerifyPassword verifies a password against a password hash (Argon2 or bcrypt).
func (a *AuthService) VerifyPassword(password, hash string) (bool, error) {
	// Bcrypt hashes start with $2a$, $2b$, $2y$
	if strings.HasPrefix(hash, "$2") {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err == nil {
			return true, nil
		}
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("bcrypt verify: %w", err)
	}

	// Argon2id format: $argon2id$v=19$m=...,t=...,p=...$salt$hash (6 parts)
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var memory, time, threads uint32
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, fmt.Errorf("invalid hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, uint8(threads), uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1, nil
}

// Package vault is the official Go SDK for the FunctionFly zero-knowledge
// secrets vault. See vault/client.go for usage.
package vault

import "fmt"

// Argon2id parameters — OWASP 2023 recommended minimums.
// Memory: 64 MiB, Iterations: 3, Parallelism: 4, Salt: 16 bytes, Key: 32 bytes.
const (
	Argon2MemoryKiB   uint32 = 64 * 1024
	Argon2Iterations  uint32 = 3
	Argon2Parallelism uint8  = 4
	Argon2SaltBytes   int    = 16
	Argon2KeyBytes    uint32 = 32
)

// KDFParams is the serialised KDF parameter set that travels alongside
// the ciphertext.
type KDFParams struct {
	Method      string `json:"method"`
	MemoryKiB   uint32 `json:"memory_kib,omitempty"`
	Iterations  uint32 `json:"iterations"`
	Parallelism uint8  `json:"parallelism,omitempty"`
	KeyBytes    uint32 `json:"key_bytes"`
	SaltBytes   int    `json:"salt_bytes,omitempty"`
}

// DefaultArgon2Params returns OWASP 2023 recommended Argon2id
// parameters.
func DefaultArgon2Params() KDFParams {
	return KDFParams{
		Method:      "argon2id",
		MemoryKiB:   Argon2MemoryKiB,
		Iterations:  Argon2Iterations,
		Parallelism: Argon2Parallelism,
		KeyBytes:    Argon2KeyBytes,
		SaltBytes:   Argon2SaltBytes,
	}
}

// NewSalt returns a cryptographically random salt of the given size.
// Sizes <= 0 are rejected.
func NewSalt(size int) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("salt size must be positive")
	}
	return newRandomBytes(size)
}

// NewPassword returns a URL-safe random password of the given length.
// The alphabet omits 0/O/1/l/I to avoid visual confusion in logs.
func NewPassword(length int) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("password length must be at least 8")
	}
	return randomString(length, passwordAlphabet)
}

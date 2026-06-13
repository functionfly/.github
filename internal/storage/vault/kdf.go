package vault

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters — OWASP 2023 recommended minimums.
// Memory: 64 MiB, Iterations: 3, Parallelism: 4, Salt: 16 bytes, Key: 32 bytes.
const (
	Argon2MemoryKiB   uint32 = 64 * 1024
	Argon2Iterations  uint32 = 3
	Argon2Parallelism uint8  = 4
	Argon2SaltBytes   int    = 16
	Argon2KeyBytes    uint32 = 32
)

// PBKDF2 defaults for backwards compatibility with v1 secrets.
const (
	PBKDF2Iterations = 100_000
	PBKDF2KeyBytes   = 32
	PBKDF2SaltBytes  = 16
)

// KDFParams describes the parameters of a key derivation function in a
// way that can be serialised to JSONB alongside the ciphertext.
type KDFParams struct {
	Method      KDFMethod `json:"method"`
	MemoryKiB   uint32    `json:"memory_kib,omitempty"`
	Iterations  uint32    `json:"iterations"`
	Parallelism uint8     `json:"parallelism,omitempty"`
	KeyBytes    uint32    `json:"key_bytes"`
	SaltBytes   int       `json:"salt_bytes,omitempty"`
}

// KeyDeriver is the contract every supported KDF must satisfy.
// DeriveKey returns a deterministic key for a given (passphrase, salt) pair.
type KeyDeriver interface {
	DeriveKey(passphrase string, salt []byte) ([]byte, error)
	DefaultParams() KDFParams
	Method() KDFMethod
}

// Argon2idDeriver is the default deriver for new secrets.
type Argon2idDeriver struct{}

func (Argon2idDeriver) DeriveKey(passphrase string, salt []byte) ([]byte, error) {
	if len(salt) < Argon2SaltBytes {
		return nil, fmt.Errorf("argon2id: salt must be at least %d bytes", Argon2SaltBytes)
	}
	return argon2.IDKey([]byte(passphrase), salt, Argon2Iterations, Argon2MemoryKiB, Argon2Parallelism, Argon2KeyBytes), nil
}

func (Argon2idDeriver) DefaultParams() KDFParams {
	return KDFParams{
		Method:      KDFMethodArgon2id,
		MemoryKiB:   Argon2MemoryKiB,
		Iterations:  Argon2Iterations,
		Parallelism: Argon2Parallelism,
		KeyBytes:    Argon2KeyBytes,
		SaltBytes:   Argon2SaltBytes,
	}
}

func (Argon2idDeriver) Method() KDFMethod { return KDFMethodArgon2id }

// NewDeriver returns the appropriate deriver for a given stored key_version.
// Unknown versions return an error so misconfigured rows fail loudly.
func NewDeriver(keyVersion int) (KeyDeriver, error) {
	switch keyVersion {
	case KeyVersionArgon2:
		return Argon2idDeriver{}, nil
	case KeyVersionPBKDF2:
		return PBKDF2Deriver{}, nil
	default:
		return nil, fmt.Errorf("vault: unsupported key_version %d", keyVersion)
	}
}

// GenerateSalt returns a cryptographically random salt of the requested size.
func GenerateSalt(size int) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New("vault: salt size must be positive")
	}
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("vault: failed to generate salt: %w", err)
	}
	return b, nil
}

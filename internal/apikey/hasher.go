package apikey

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// Hasher handles API key hashing and verification
type Hasher struct {
	// salt can be used to add additional security (optional)
	salt string
}

// NewHasher creates a new key hasher
func NewHasher() *Hasher {
	return &Hasher{}
}

// NewHasherWithSalt creates a new key hasher with a salt
func NewHasherWithSalt(salt string) *Hasher {
	return &Hasher{salt: salt}
}

// Hash returns the SHA-256 hash of the key
// The hash is used for secure storage in the database
func (h *Hasher) Hash(key string) string {
	data := key
	if h.salt != "" {
		data = h.salt + key
	}
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HashBytes returns the SHA-256 hash of the given bytes
func (h *Hasher) HashBytes(data []byte) string {
	if h.salt != "" {
		data = append([]byte(h.salt), data...)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Verify compares a plaintext key with its hash in constant time
// This prevents timing attacks by using subtle.ConstantTimeCompare
func (h *Hasher) Verify(plaintext, hash string) bool {
	computedHash := h.Hash(plaintext)
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(hash)) == 1
}

// VerifyPrefix verifies that the key has the correct prefix
// This is useful for quickly identifying key type before doing a full hash lookup
func VerifyPrefix(key, expectedPrefix string) bool {
	return len(key) > len(expectedPrefix) && key[:len(expectedPrefix)] == expectedPrefix
}

// VerifyMultiplePrefixes checks if the key matches any of the given prefixes
func VerifyMultiplePrefixes(key string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if VerifyPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// ShortHash returns a shortened hash for display/logging purposes
// It returns the first 8 characters of the hash
func ShortHash(hash string) string {
	if len(hash) < 8 {
		return hash
	}
	return hash[:8]
}

// MaskKey returns a masked version of the key for display
// Example: ffp_v1_xxxx_xxxx_xxxx_xxxx_xxxx_xxxx_xxxx
func MaskKey(key string) string {
	if len(key) < 12 {
		return "****"
	}

	// Show first 8 characters (prefix + version + first part)
	// and last 4 characters
	prefix := key[:8]
	suffix := key[len(key)-4:]

	return prefix + "_xxxx_xxxx_" + suffix
}

// HashKey is a convenience function that hashes a key using the default hasher
func HashKey(key string) string {
	return NewHasher().Hash(key)
}

// VerifyKey is a convenience function that verifies a key against a hash
func VerifyKey(plaintext, hash string) bool {
	return NewHasher().Verify(plaintext, hash)
}

// PrepareKeyForStorage prepares a key for secure storage
// This includes normalizing the key and generating the hash
func PrepareKeyForStorage(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	// Validate the key format first
	if err := ValidateKeyFormat(key); err != nil {
		return "", fmt.Errorf("invalid key format: %w", err)
	}

	return HashKey(key), nil
}

// KeyInfo contains parsed information about a key
type KeyInfo struct {
	Prefix    string
	Version   string
	RandomHex string
	Checksum  string
}

// ExtractKeyInfo extracts information from a key without validating
// Returns nil if the key format is invalid
func ExtractKeyInfo(key string) *KeyInfo {
	prefix, version, randomHex, checksumHex, err := ParseKey(key)
	if err != nil {
		return nil
	}
	return &KeyInfo{
		Prefix:    prefix,
		Version:   version,
		RandomHex: randomHex,
		Checksum:  checksumHex,
	}
}

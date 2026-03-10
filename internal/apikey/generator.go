package apikey

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// Version is the current key format version
	Version = "v1"
	// Separator separates key components
	Separator = "_"
	// CRC8Length is the length of the CRC checksum in hex characters
	CRC8Length = 2
)

// Generator handles API key generation
type Generator struct {
	// randomSource can be replaced for testing
	randomSource func([]byte) (int, error)
}

// NewGenerator creates a new key generator
func NewGenerator() *Generator {
	return &Generator{
		randomSource: rand.Read,
	}
}

// Generate creates a new API key with the given type
// The key format is: {prefix}_{version}_{random16bytes}_{crc8checksum}
// Example: ffp_v1_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6_a1b2
func (g *Generator) Generate(keyType KeyType) (string, error) {
	prefix := GetPrefixForKeyType(keyType)
	return g.GenerateWithPrefix(prefix)
}

// GenerateWithPrefix creates a new API key with the given prefix
func (g *Generator) GenerateWithPrefix(prefix string) (string, error) {
	// Generate 16 random bytes (32 hex characters)
	randomBytes := make([]byte, RandomBytesLength)
	n, err := g.randomSource(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	if n != RandomBytesLength {
		return "", fmt.Errorf("failed to generate enough random bytes: got %d, want %d", n, RandomBytesLength)
	}

	randomHex := hex.EncodeToString(randomBytes)

	// Calculate CRC8 checksum of the prefix + version + random
	checksum := calculateCRC8(prefix + Version + Separator + randomHex)
	checksumHex := hex.EncodeToString([]byte{checksum})

	// Construct the key: prefix_v1_randomHex_checksum
	key := prefix + Version + Separator + randomHex + Separator + checksumHex

	return key, nil
}

// GenerateForType is a convenience function that creates a new API key for the given type
func GenerateForType(keyType KeyType) (string, error) {
	return NewGenerator().Generate(keyType)
}

// ParseKey parses an API key and returns its components
func ParseKey(key string) (prefix, version, randomHex, checksumHex string, err error) {
	parts := strings.Split(key, Separator)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid key format: expected 4 parts, got %d", len(parts))
	}

	prefix = parts[0]
	version = parts[1]
	randomHex = parts[2]
	checksumHex = parts[3]

	// Validate version format
	if version != Version {
		return "", "", "", "", fmt.Errorf("unsupported key version: %s", version)
	}

	// Validate random hex length
	if len(randomHex) != RandomBytesLength*2 {
		return "", "", "", "", fmt.Errorf("invalid random bytes length: expected %d, got %d", RandomBytesLength*2, len(randomHex))
	}

	// Validate checksum hex length
	if len(checksumHex) != CRC8Length {
		return "", "", "", "", fmt.Errorf("invalid checksum length: expected %d, got %d", CRC8Length, len(checksumHex))
	}

	return prefix, version, randomHex, checksumHex, nil
}

// ValidateKeyFormat validates the key format (prefix, version, random, checksum)
// This does NOT validate the key against stored hashes - use the Validator for that
func ValidateKeyFormat(key string) error {
	_, _, _, _, err := ParseKey(key)
	return err
}

// calculateCRC8 calculates a CRC8 checksum of the input data
// This is a simple checksum for quick validation, not cryptographic
func calculateCRC8(data string) byte {
	crc := byte(0)
	for i := 0; i < len(data); i++ {
		crc ^= byte(data[i])
		// Simple polynomial: x^8 + x^2 + x + 1 (0x07)
		for j := 0; j < 8; j++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x07
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// VerifyChecksum verifies that the checksum in the key is correct
func VerifyChecksum(key string) bool {
	prefix, version, randomHex, checksumHex, err := ParseKey(key)
	if err != nil {
		return false
	}

	expectedChecksum := calculateCRC8(prefix + version + Separator + randomHex)
	expectedChecksumHex := hex.EncodeToString([]byte{expectedChecksum})

	return strings.EqualFold(checksumHex, expectedChecksumHex)
}

// GetKeyTypeFromPrefix returns the key type for a given prefix
func GetKeyTypeFromPrefix(prefix string) (KeyType, error) {
	switch prefix {
	case PrefixPlatform:
		return KeyTypePlatform, nil
	case PrefixFunction:
		return KeyTypeFunction, nil
	case PrefixAgent:
		return KeyTypeAgent, nil
	case PrefixEnvironment:
		return KeyTypeEnvironment, nil
	case PrefixOAuth:
		return KeyTypeOAuth, nil
	default:
		return "", fmt.Errorf("unknown key prefix: %s", prefix)
	}
}

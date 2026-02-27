package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

// NormalizeInput canonicalizes JSON input for consistent hashing
// This ensures identical inputs produce identical cache keys regardless of
// formatting differences like key order, whitespace, or number representation
func NormalizeInput(input []byte) ([]byte, error) {
	// First pass: unmarshal to interface{}
	var raw interface{}
	if err := json.Unmarshal(input, &raw); err != nil {
		return nil, err
	}

	// Recursively normalize
	normalized := normalizeValue(raw)

	// Marshal with sorted keys (ensure map keys are ordered)
	return json.Marshal(normalized)
}

// normalizeValue recursively normalizes a JSON value
func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// Sort keys alphabetically for consistent ordering
		sorted := make(map[string]interface{})
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sorted[k] = normalizeValue(val[k])
		}
		return sorted

	case []interface{}:
		// Normalize each array element
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = normalizeValue(item)
		}
		return result

	case string:
		// Trim whitespace, collapse interior spaces to single space
		trimmed := strings.TrimSpace(val)
		return strings.Join(strings.Fields(trimmed), " ")

	case float64:
		// JSON unmarshals all numbers as float64
		// This preserves canonical form (no trailing zeros in marshaled output)
		return val

	case bool:
		// Already canonical
		return val

	case nil:
		return nil

	default:
		return v
	}
}

// GenerateCacheKey creates a deterministic cache key from function metadata and normalized input
// Format: fx:cache:{function_id}:{version}:{hash16}
func GenerateCacheKey(functionID, version string, normalizedInput []byte) string {
	hasher := sha256.New()
	hasher.Write([]byte(functionID))
	hasher.Write([]byte("::"))
	hasher.Write([]byte(version))
	hasher.Write([]byte("::"))
	hasher.Write(normalizedInput)

	hash := hex.EncodeToString(hasher.Sum(nil))
	// Use first 16 chars of hash for reasonable collision resistance
	return "fx:cache:" + functionID + ":" + version + ":" + hash[:16]
}

// HashInput returns just the SHA-256 hash of normalized input
func HashInput(normalizedInput []byte) string {
	hasher := sha256.New()
	hasher.Write(normalizedInput)
	return hex.EncodeToString(hasher.Sum(nil))
}

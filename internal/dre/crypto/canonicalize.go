package crypto

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Canonicalize serializes any value to RFC-8785 canonical JSON:
//   - Sorted keys (lexicographic)
//   - No whitespace
//   - UTF-8 encoded
//   - Floats normalized (no trailing zeros)
//   - Timestamps should be pre-formatted as ISO-8601 UTC by the caller
//   - Binary data should be pre-encoded as base64 by the caller
//
// This implementation provides deterministic JSON serialization suitable
// for cryptographic hashing. It follows the key ordering rules of RFC-8785.
func Canonicalize(v interface{}) ([]byte, error) {
	// First marshal to get a normalized representation
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("canonicalize: marshal failed: %w", err)
	}

	// Re-parse and re-serialize with sorted keys
	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("canonicalize: unmarshal failed: %w", err)
	}

	return canonicalizeValue(parsed)
}

// canonicalizeValue recursively canonicalizes a parsed JSON value.
func canonicalizeValue(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case nil:
		return []byte("null"), nil

	case bool:
		if val {
			return []byte("true"), nil
		}
		return []byte("false"), nil

	case float64:
		// Use standard JSON encoding for numbers
		b, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("canonicalize number: %w", err)
		}
		return b, nil

	case string:
		b, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("canonicalize string: %w", err)
		}
		return b, nil

	case []interface{}:
		result := []byte("[")
		for i, item := range val {
			itemBytes, err := canonicalizeValue(item)
			if err != nil {
				return nil, fmt.Errorf("canonicalize array[%d]: %w", i, err)
			}
			if i > 0 {
				result = append(result, ',')
			}
			result = append(result, itemBytes...)
		}
		result = append(result, ']')
		return result, nil

	case map[string]interface{}:
		// Sort keys lexicographically (RFC-8785 requirement)
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		result := []byte("{")
		for i, k := range keys {
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, fmt.Errorf("canonicalize key %q: %w", k, err)
			}
			valBytes, err := canonicalizeValue(val[k])
			if err != nil {
				return nil, fmt.Errorf("canonicalize value for key %q: %w", k, err)
			}
			if i > 0 {
				result = append(result, ',')
			}
			result = append(result, keyBytes...)
			result = append(result, ':')
			result = append(result, valBytes...)
		}
		result = append(result, '}')
		return result, nil

	default:
		// Fallback for any other types
		b, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("canonicalize unknown type: %w", err)
		}
		return b, nil
	}
}

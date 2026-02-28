package cache

import (
	"encoding/json"
	"testing"
)

// TestNormalizeInput_KeyOrdering tests that key ordering doesn't affect the normalized output
func TestNormalizeInput_KeyOrdering(t *testing.T) {
	input1 := []byte(`{"b":2,"a":1}`)
	input2 := []byte(`{"a":1,"b":2}`)

	norm1, err := NormalizeInput(input1)
	if err != nil {
		t.Fatalf("Failed to normalize input1: %v", err)
	}

	norm2, err := NormalizeInput(input2)
	if err != nil {
		t.Fatalf("Failed to normalize input2: %v", err)
	}

	if string(norm1) != string(norm2) {
		t.Errorf("Key ordering should not affect normalization.\nGot:      %s\nExpected: %s", norm1, norm2)
	}
}

// TestNormalizeInput_Whitespace tests that whitespace is normalized
func TestNormalizeInput_Whitespace(t *testing.T) {
	input1 := []byte(`{"a": "hello   world"}`)
	input2 := []byte(`{"a":"hello world"}`)

	norm1, err := NormalizeInput(input1)
	if err != nil {
		t.Fatalf("Failed to normalize input1: %v", err)
	}

	norm2, err := NormalizeInput(input2)
	if err != nil {
		t.Fatalf("Failed to normalize input2: %v", err)
	}

	if string(norm1) != string(norm2) {
		t.Errorf("Whitespace should be normalized.\nGot:      %s\nExpected: %s", norm1, norm2)
	}
}

// TestNormalizeInput_NestedObjects tests that nested objects are normalized recursively
func TestNormalizeInput_NestedObjects(t *testing.T) {
	input1 := []byte(`{"z":{"b":2,"a":1},"a":1}`)
	input2 := []byte(`{"a":1,"z":{"a":1,"b":2}}`)

	norm1, err := NormalizeInput(input1)
	if err != nil {
		t.Fatalf("Failed to normalize input1: %v", err)
	}

	norm2, err := NormalizeInput(input2)
	if err != nil {
		t.Fatalf("Failed to normalize input2: %v", err)
	}

	if string(norm1) != string(norm2) {
		t.Errorf("Nested objects should be normalized consistently.\nGot:      %s\nExpected: %s", norm1, norm2)
	}
}

// TestNormalizeInput_ArrayOrder tests that array order is preserved
func TestNormalizeInput_ArrayOrder(t *testing.T) {
	input1 := []byte(`{"a":[1,2,3]}`)
	input2 := []byte(`{"a":[3,2,1]}`)

	norm1, err := NormalizeInput(input1)
	if err != nil {
		t.Fatalf("Failed to normalize input1: %v", err)
	}

	norm2, err := NormalizeInput(input2)
	if err != nil {
		t.Fatalf("Failed to normalize input2: %v", err)
	}

	if string(norm1) == string(norm2) {
		t.Error("Array order should be preserved and produce different normalizations")
	}
}

// TestNormalizeInput_BooleanStrictTyping tests that booleans and strings are distinct
func TestNormalizeInput_BooleanStrictTyping(t *testing.T) {
	input1 := []byte(`{"a":true}`)
	input2 := []byte(`{"a":"true"}`)

	norm1, err := NormalizeInput(input1)
	if err != nil {
		t.Fatalf("Failed to normalize input1: %v", err)
	}

	norm2, err := NormalizeInput(input2)
	if err != nil {
		t.Fatalf("Failed to normalize input2: %v", err)
	}

	if string(norm1) == string(norm2) {
		t.Error("Boolean true should be different from string 'true'")
	}
}

// TestNormalizeInput_NumberNormalization tests that equivalent numbers are normalized the same
func TestNormalizeInput_NumberNormalization(t *testing.T) {
	input1 := []byte(`{"a":1.0}`)
	input2 := []byte(`{"a":1}`)

	norm1, err := NormalizeInput(input1)
	if err != nil {
		t.Fatalf("Failed to normalize input1: %v", err)
	}

	norm2, err := NormalizeInput(input2)
	if err != nil {
		t.Fatalf("Failed to normalize input2: %v", err)
	}

	// JSON unmarshals both as float64, so they should be equal
	if string(norm1) != string(norm2) {
		t.Errorf("Numeric values 1.0 and 1 should be equivalent.\nGot:      %s\nExpected: %s", norm1, norm2)
	}
}

// TestNormalizeInput_InvalidJSON tests error handling for invalid JSON
func TestNormalizeInput_InvalidJSON(t *testing.T) {
	input := []byte(`{invalid json}`)

	_, err := NormalizeInput(input)
	if err == nil {
		t.Error("Should return error for invalid JSON")
	}
}

// TestNormalizeInput_EmptyObject tests empty object handling
func TestNormalizeInput_EmptyObject(t *testing.T) {
	input := []byte(`{}`)

	norm, err := NormalizeInput(input)
	if err != nil {
		t.Fatalf("Failed to normalize empty object: %v", err)
	}

	if string(norm) != "{}" {
		t.Errorf("Empty object should normalize to {}. Got: %s", norm)
	}
}

// TestNormalizeInput_NullValue tests null value handling
func TestNormalizeInput_NullValue(t *testing.T) {
	input1 := []byte(`{"a":null}`)
	input2 := []byte(`{"a":null}`)

	norm1, err := NormalizeInput(input1)
	if err != nil {
		t.Fatalf("Failed to normalize input1: %v", err)
	}

	norm2, err := NormalizeInput(input2)
	if err != nil {
		t.Fatalf("Failed to normalize input2: %v", err)
	}

	if string(norm1) != string(norm2) {
		t.Errorf("Null values should normalize consistently.\nGot:      %s\nExpected: %s", norm1, norm2)
	}
}

// TestGenerateCacheKey_VersionIsolation tests that different versions produce different keys
func TestGenerateCacheKey_VersionIsolation(t *testing.T) {
	functionID := "550e8400-e29b-41d4-a716-446655440000"
	input := []byte(`{"test":"value"}`)

	key1 := GenerateCacheKey(functionID, "v1.0.0", input)
	key2 := GenerateCacheKey(functionID, "v1.0.1", input)

	if key1 == key2 {
		t.Error("Different versions should produce different cache keys")
	}
}

// TestGenerateCacheKey_FunctionIDIsolation tests that different function IDs produce different keys
func TestGenerateCacheKey_FunctionIDIsolation(t *testing.T) {
	version := "v1.0.0"
	input := []byte(`{"test":"value"}`)

	key1 := GenerateCacheKey("550e8400-e29b-41d4-a716-446655440000", version, input)
	key2 := GenerateCacheKey("550e8400-e29b-41d4-a716-446655440001", version, input)

	if key1 == key2 {
		t.Error("Different function IDs should produce different cache keys")
	}
}

// TestGenerateCacheKey_InputIsolation tests that different inputs produce different keys
func TestGenerateCacheKey_InputIsolation(t *testing.T) {
	functionID := "550e8400-e29b-41d4-a716-446655440000"
	version := "v1.0.0"

	input1 := []byte(`{"test":"value1"}`)
	input2 := []byte(`{"test":"value2"}`)

	key1 := GenerateCacheKey(functionID, version, input1)
	key2 := GenerateCacheKey(functionID, version, input2)

	if key1 == key2 {
		t.Error("Different inputs should produce different cache keys")
	}
}

// TestGenerateCacheKey_KeyFormat tests that the generated key has the expected format
func TestGenerateCacheKey_KeyFormat(t *testing.T) {
	functionID := "550e8400-e29b-41d4-a716-446655440000"
	version := "v1.0.0"
	input := []byte(`{"test":"value"}`)

	key := GenerateCacheKey(functionID, version, input)

	// Check prefix
	expectedPrefix := "fx:cache:" + functionID + ":" + version + ":"
	if len(key) < len(expectedPrefix) || key[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Cache key should start with %s. Got: %s", expectedPrefix, key)
	}

	// Check total length (prefix + 16 char hash)
	expectedLen := len(expectedPrefix) + 16
	if len(key) != expectedLen {
		t.Errorf("Cache key should be %d characters. Got: %d (%s)", expectedLen, len(key), key)
	}
}

// TestHashInput_Deterministic tests that hash generation is deterministic
func TestHashInput_Deterministic(t *testing.T) {
	input := []byte(`{"test":"value"}`)

	hash1 := HashInput(input)
	hash2 := HashInput(input)

	if hash1 != hash2 {
		t.Error("Hash should be deterministic for same input")
	}
}

// TestHashInput_DifferentInputs tests that different inputs produce different hashes
func TestHashInput_DifferentInputs(t *testing.T) {
	input1 := []byte(`{"test":"value1"}`)
	input2 := []byte(`{"test":"value2"}`)

	hash1 := HashInput(input1)
	hash2 := HashInput(input2)

	if hash1 == hash2 {
		t.Error("Different inputs should produce different hashes")
	}
}

// TestHashInput_Length tests that generated hash has expected length (SHA-256 = 64 hex chars)
func TestHashInput_Length(t *testing.T) {
	input := []byte(`{"test":"value"}`)

	hash := HashInput(input)

	if len(hash) != 64 {
		t.Errorf("SHA-256 hash should be 64 hex characters. Got: %d (%s)", len(hash), hash)
	}
}

// TestNormalizeValue_StringTrimming tests string whitespace trimming
func TestNormalizeValue_StringTrimming(t *testing.T) {
	input := []byte(`{"a":"  hello   world  "}`)

	norm, err := NormalizeInput(input)
	if err != nil {
		t.Fatalf("Failed to normalize: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(norm, &result); err != nil {
		t.Fatalf("Failed to unmarshal normalized: %v", err)
	}

	if result["a"] != "hello world" {
		t.Errorf("String should be trimmed and whitespace collapsed. Got: %q", result["a"])
	}
}

// TestNormalizeValue_DeepNesting tests deeply nested object normalization
func TestNormalizeValue_DeepNesting(t *testing.T) {
	input1 := []byte(`{"level1":{"level2":{"level3":{"b":2,"a":1}}}}`)
	input2 := []byte(`{"level1":{"level2":{"level3":{"a":1,"b":2}}}}`)

	norm1, err := NormalizeInput(input1)
	if err != nil {
		t.Fatalf("Failed to normalize input1: %v", err)
	}

	norm2, err := NormalizeInput(input2)
	if err != nil {
		t.Fatalf("Failed to normalize input2: %v", err)
	}

	if string(norm1) != string(norm2) {
		t.Errorf("Deeply nested objects should normalize consistently.\nGot:      %s\nExpected: %s", norm1, norm2)
	}
}

// TestNormalizeValue_EmptyArray tests empty array handling
func TestNormalizeValue_EmptyArray(t *testing.T) {
	input := []byte(`{"a":[]}`)

	norm, err := NormalizeInput(input)
	if err != nil {
		t.Fatalf("Failed to normalize: %v", err)
	}

	expected := `{"a":[]}`
	if string(norm) != expected {
		t.Errorf("Empty array should normalize consistently. Got: %s, Expected: %s", norm, expected)
	}
}

// TestNormalizeValue_MixedTypes tests mixed type handling
func TestNormalizeValue_MixedTypes(t *testing.T) {
	input := []byte(`{"string":"value","number":42,"bool":true,"null":null,"array":[1,2],"object":{"a":1}}`)

	norm, err := NormalizeInput(input)
	if err != nil {
		t.Fatalf("Failed to normalize: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(norm, &result); err != nil {
		t.Fatalf("Normalized result should be valid JSON: %v", err)
	}

	// Check all types are preserved
	if result["string"] != "value" {
		t.Errorf("String value not preserved")
	}
	if result["number"] != 42.0 { // JSON numbers unmarshal as float64
		t.Errorf("Number value not preserved")
	}
	if result["bool"] != true {
		t.Errorf("Boolean value not preserved")
	}
	if result["null"] != nil {
		t.Errorf("Null value not preserved")
	}
}

// BenchmarkNormalizeInput benchmarks the normalization function
func BenchmarkNormalizeInput(b *testing.B) {
	input := []byte(`{"name":"test","values":[1,2,3,4,5],"nested":{"a":1,"b":2},"active":true}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NormalizeInput(input)
		if err != nil {
			b.Fatalf("Failed to normalize: %v", err)
		}
	}
}

// BenchmarkGenerateCacheKey benchmarks the cache key generation
func BenchmarkGenerateCacheKey(b *testing.B) {
	functionID := "550e8400-e29b-41d4-a716-446655440000"
	version := "v1.0.0"
	input := []byte(`{"name":"test","values":[1,2,3,4,5]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCacheKey(functionID, version, input)
	}
}

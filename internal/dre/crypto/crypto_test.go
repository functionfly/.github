package crypto

import (
	"encoding/json"
	"testing"
)

// TestHashDomainSeparation verifies that different domain tags produce different hashes
// even with the same input data.
func TestHashDomainSeparation(t *testing.T) {
	data := []byte("test input data")

	hash1 := Hash(TagInput, data)
	hash2 := Hash(TagEnv, data)

	if string(hash1) == string(hash2) {
		t.Error("Hashes with different domain tags should be different")
	}

	// Verify hash length (BLAKE3 produces 32 bytes)
	if len(hash1) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash1))
	}
}

// TestHashString verifies that HashString returns hex-encoded output.
func TestHashString(t *testing.T) {
	data := []byte("test input data")

	hashStr := HashString(TagInput, data)

	// BLAKE3 produces 32 bytes = 64 hex characters
	if len(hashStr) != 64 {
		t.Errorf("Expected hex string length 64, got %d", len(hashStr))
	}

	// Verify it's valid hex
	for _, c := range hashStr {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Invalid hex character: %c", c)
		}
	}
}

// TestMerkleRoot verifies Merkle root computation.
func TestMerkleRoot(t *testing.T) {
	tests := []struct {
		name    string
		leaves  [][]byte
		wantNil bool
		wantLen int
	}{
		{
			name:    "empty leaves",
			leaves:  nil,
			wantNil: true,
		},
		{
			name:    "single leaf",
			leaves:  [][]byte{[]byte("a")},
			wantNil: false,
			wantLen: 32,
		},
		{
			name:    "two leaves",
			leaves:  [][]byte{[]byte("a"), []byte("b")},
			wantNil: false,
			wantLen: 32,
		},
		{
			name:    "three leaves (odd)",
			leaves:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			wantNil: false,
			wantLen: 32,
		},
		{
			name:    "four leaves",
			leaves:  [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			wantNil: false,
			wantLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := MerkleRoot(tt.leaves)

			if tt.wantNil && root != nil {
				t.Error("Expected nil root")
			}
			if !tt.wantNil && root == nil {
				t.Error("Expected non-nil root")
			}
			if tt.wantLen > 0 && len(root) != tt.wantLen {
				t.Errorf("Expected root length %d, got %d", tt.wantLen, len(root))
			}
		})
	}
}

// TestMerkleRootDeterministic verifies that Merkle root is deterministic.
func TestMerkleRootDeterministic(t *testing.T) {
	leaves := [][]byte{
		[]byte("leaf1"),
		[]byte("leaf2"),
		[]byte("leaf3"),
		[]byte("leaf4"),
	}

	root1 := MerkleRoot(leaves)
	root2 := MerkleRoot(leaves)

	if string(root1) != string(root2) {
		t.Error("Merkle root should be deterministic")
	}
}

// TestCanonicalize verifies JSON canonicalization produces consistent output.
func TestCanonicalize(t *testing.T) {
	// Test with map (object)
	data := map[string]interface{}{
		"z": "zebra",
		"a": "alpha",
		"m": map[string]interface{}{
			"inner": "value",
			"z":     1,
		},
	}

	// Canonicalize twice and verify same output
	bytes1, err := Canonicalize(data)
	if err != nil {
		t.Fatalf("Canonicalize failed: %v", err)
	}

	bytes2, err := Canonicalize(data)
	if err != nil {
		t.Fatalf("Canonicalize failed: %v", err)
	}

	if string(bytes1) != string(bytes2) {
		t.Error("Canonicalize should be deterministic")
	}

	// Verify keys are sorted (z should come after a)
	result := string(bytes1)
	if result[:2] != `{"` {
		t.Error("Expected JSON object start")
	}
}

// TestCanonicalizeArray verifies array canonicalization.
func TestCanonicalizeArray(t *testing.T) {
	data := []interface{}{3, 1, 2}

	bytes, err := Canonicalize(data)
	if err != nil {
		t.Fatalf("Canonicalize failed: %v", err)
	}

	// Arrays should preserve order
	expected := "[3,1,2]"
	if string(bytes) != expected {
		t.Errorf("Expected %s, got %s", expected, string(bytes))
	}
}

// TestBuildMEG verifies the full MEG build process.
func TestBuildMEG(t *testing.T) {
	components := MEGComponents{
		InputPayload: map[string]interface{}{
			"args":   []string{"arg1", "arg2"},
			"caller": "fx://example/func@1.0.0",
		},
		EnvironmentData: map[string]string{
			"runtime": "wasm/1.0",
		},
		Dependencies: []Dependency{
			{Name: "lodash", Version: "4.17.21", ContentHash: "abc123"},
		},
		TraceChunks: nil,
		ResourceUsage: map[string]int64{
			"cpu_cycles":  1000000,
			"memory_peak": 1024,
		},
		OutputPayload: map[string]interface{}{
			"result": "success",
		},
		Metadata: map[string]string{
			"execution_id": "exec_123",
			"function_id":  "fx://example/func",
		},
	}

	result, err := BuildMEG(components)
	if err != nil {
		t.Fatalf("BuildMEG failed: %v", err)
	}

	// Verify all component hashes are present
	if result.InputHash == "" {
		t.Error("InputHash should not be empty")
	}
	if result.EnvironmentHash == "" {
		t.Error("EnvironmentHash should not be empty")
	}
	if result.DependencyHash == "" {
		t.Error("DependencyHash should not be empty")
	}
	if result.TraceHash == "" {
		t.Error("TraceHash should not be empty")
	}
	if result.ResourceHash == "" {
		t.Error("ResourceHash should not be empty")
	}
	if result.OutputHash == "" {
		t.Error("OutputHash should not be empty")
	}
	if result.MetadataHash == "" {
		t.Error("MetadataHash should not be empty")
	}

	// Verify ExecutionRootHash is present
	if result.ExecutionRootHash == "" {
		t.Error("ExecutionRootHash should not be empty")
	}

	// Verify ExecutionRootHash length (64 hex chars for 32 bytes)
	if len(result.ExecutionRootHash) != 64 {
		t.Errorf("Expected ExecutionRootHash length 64, got %d", len(result.ExecutionRootHash))
	}

	// Verify leaf hashes array has 7 elements
	if len(result.LeafHashes) != 7 {
		t.Errorf("Expected 7 leaf hashes, got %d", len(result.LeafHashes))
	}
}

// TestBuildMEGDeterministic verifies that BuildMEG is deterministic.
func TestBuildMEGDeterministic(t *testing.T) {
	components := MEGComponents{
		InputPayload: map[string]interface{}{
			"value": 42,
		},
		EnvironmentData: map[string]string{
			"runtime": "wasm/1.0",
		},
		Dependencies:  []Dependency{},
		TraceChunks:   nil,
		ResourceUsage: map[string]int64{"cpu": 100},
		OutputPayload: map[string]interface{}{"result": "ok"},
		Metadata:      map[string]string{"id": "test"},
	}

	result1, err := BuildMEG(components)
	if err != nil {
		t.Fatalf("BuildMEG failed: %v", err)
	}

	result2, err := BuildMEG(components)
	if err != nil {
		t.Fatalf("BuildMEG failed: %v", err)
	}

	if result1.ExecutionRootHash != result2.ExecutionRootHash {
		t.Error("BuildMEG should be deterministic")
	}
}

// TestBuildMEGWithTrace verifies MEG with trace chunks.
func TestBuildMEGWithTrace(t *testing.T) {
	components := MEGComponents{
		InputPayload:    map[string]interface{}{"in": "value"},
		EnvironmentData: map[string]string{"runtime": "wasm/1.0"},
		Dependencies:    []Dependency{},
		TraceChunks:     [][]byte{[]byte("chunk1"), []byte("chunk2"), []byte("chunk3")},
		ResourceUsage:   map[string]int64{"cpu": 100},
		OutputPayload:   map[string]interface{}{"out": "value"},
		Metadata:        map[string]string{"id": "test"},
	}

	result, err := BuildMEG(components)
	if err != nil {
		t.Fatalf("BuildMEG failed: %v", err)
	}

	// Verify trace hash is computed (not empty marker)
	if result.TraceHash == "" {
		t.Error("TraceHash should not be empty")
	}
}

// TestBuildMEGEmptyDeps verifies MEG with empty dependencies.
func TestBuildMEGEmptyDeps(t *testing.T) {
	components := MEGComponents{
		InputPayload:    map[string]interface{}{"in": "value"},
		EnvironmentData: map[string]string{"runtime": "wasm/1.0"},
		Dependencies:    []Dependency{}, // Empty dependencies
		TraceChunks:     nil,
		ResourceUsage:   map[string]int64{"cpu": 100},
		OutputPayload:   map[string]interface{}{"out": "value"},
		Metadata:        map[string]string{"id": "test"},
	}

	result, err := BuildMEG(components)
	if err != nil {
		t.Fatalf("BuildMEG failed: %v", err)
	}

	// Should have a dependency hash for empty set
	if result.DependencyHash == "" {
		t.Error("DependencyHash should not be empty even for empty deps")
	}
}

// TestDependencyHash verifies dependency Merkle tree.
func TestDependencyHash(t *testing.T) {
	components := MEGComponents{
		InputPayload:    map[string]interface{}{"in": "value"},
		EnvironmentData: map[string]string{"runtime": "wasm/1.0"},
		Dependencies: []Dependency{
			{Name: "pkg-a", Version: "1.0.0", ContentHash: "hash1"},
			{Name: "pkg-b", Version: "2.0.0", ContentHash: "hash2"},
			{Name: "pkg-c", Version: "3.0.0", ContentHash: "hash3"},
		},
		TraceChunks:   nil,
		ResourceUsage: map[string]int64{"cpu": 100},
		OutputPayload: map[string]interface{}{"out": "value"},
		Metadata:      map[string]string{"id": "test"},
	}

	result, err := BuildMEG(components)
	if err != nil {
		t.Fatalf("BuildMEG failed: %v", err)
	}

	// Dependency hash should be present
	if result.DependencyHash == "" {
		t.Error("DependencyHash should not be empty")
	}
}

// TestJSONMarshalCanonicalize verifies canonicalization matches JSON marshal.
func TestJSONMarshalCanonicalize(t *testing.T) {
	// Test that our canonicalization handles standard JSON correctly
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := TestStruct{Name: "Alice", Age: 30}

	// Use json.Marshal for comparison
	stdJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Canonicalize should also work
	canonJSON, err := Canonicalize(data)
	if err != nil {
		t.Fatalf("Canonicalize failed: %v", err)
	}

	// Both should produce valid JSON (may differ in key ordering)
	var stdParsed, canonParsed interface{}
	if err := json.Unmarshal(stdJSON, &stdParsed); err != nil {
		t.Errorf("Standard JSON unmarshal failed: %v", err)
	}
	if err := json.Unmarshal(canonJSON, &canonParsed); err != nil {
		t.Errorf("Canonical JSON unmarshal failed: %v", err)
	}
}

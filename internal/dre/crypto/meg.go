package crypto

import (
	"encoding/hex"
	"fmt"
)

// MEGComponents holds all the inputs needed to build a Merkle Execution Graph.
// Each field corresponds to one leaf in the Merkle tree.
type MEGComponents struct {
	// InputPayload contains the function arguments, caller identity, fx:// URI, and seed.
	InputPayload interface{}
	// EnvironmentData contains runtime versions and the capsule descriptor.
	EnvironmentData interface{}
	// Dependencies lists all function dependencies with their content hashes.
	Dependencies []Dependency
	// TraceChunks contains raw execution trace segments (nil in lite tier).
	TraceChunks [][]byte
	// ResourceUsage contains cpu_cycles, memory_peak, wall_time, etc.
	ResourceUsage interface{}
	// OutputPayload contains the return value, exit code, and events.
	OutputPayload interface{}
	// Metadata contains execution_id, function_id, nonce, protocol_version, etc.
	Metadata interface{}
}

// Dependency represents a single function dependency.
type Dependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	ContentHash string `json:"content_hash"`
}

// MEGResult contains all computed hashes from a Merkle Execution Graph build.
type MEGResult struct {
	// Component hashes (leaf nodes)
	InputHash       string `json:"input_hash"`
	EnvironmentHash string `json:"environment_hash"`
	DependencyHash  string `json:"dependency_hash"`
	TraceHash       string `json:"trace_hash"`
	ResourceHash    string `json:"resource_hash"`
	OutputHash      string `json:"output_hash"`
	MetadataHash    string `json:"metadata_hash"`

	// ExecutionRootHash is the canonical identity of this execution.
	// It is the Merkle root of the 7 component hashes in fixed order.
	ExecutionRootHash string `json:"execution_root_hash"`

	// LeafHashes contains the ordered leaf hashes for partial verification.
	// Order: [Input, Environment, Dependency, Trace, Resource, Output, Metadata]
	LeafHashes []string `json:"leaf_hashes"`
}

// BuildMEG computes the full Merkle Execution Graph from the provided components.
//
// Leaf ordering is fixed forever (DRE/1.0):
//  1. InputHash
//  2. EnvironmentHash
//  3. DependencyHash
//  4. TraceHash
//  5. ResourceHash
//  6. OutputHash
//  7. MetadataHash
func BuildMEG(components MEGComponents) (*MEGResult, error) {
	result := &MEGResult{}

	// 1. Input hash
	inputBytes, err := Canonicalize(components.InputPayload)
	if err != nil {
		return nil, fmt.Errorf("meg: canonicalize input: %w", err)
	}
	result.InputHash = HashString(TagInput, inputBytes)

	// 2. Environment hash
	envBytes, err := Canonicalize(components.EnvironmentData)
	if err != nil {
		return nil, fmt.Errorf("meg: canonicalize environment: %w", err)
	}
	result.EnvironmentHash = HashString(TagEnv, envBytes)

	// 3. Dependency hash — hash each dep node, then hash the list
	depLeaves := make([][]byte, len(components.Dependencies))
	for i, dep := range components.Dependencies {
		depBytes, err := Canonicalize(dep)
		if err != nil {
			return nil, fmt.Errorf("meg: canonicalize dep[%d]: %w", i, err)
		}
		depLeaves[i] = Hash(TagDepNode, depBytes)
	}
	if len(depLeaves) == 0 {
		// Empty dependency set — hash an empty marker
		result.DependencyHash = HashString(TagDeps, []byte("[]"))
	} else {
		depRoot := MerkleRoot(depLeaves)
		result.DependencyHash = HashString(TagDeps, depRoot)
	}

	// 4. Trace hash — hash each chunk, then Merkle-root them
	if len(components.TraceChunks) == 0 {
		// Lite tier or no trace — use empty marker
		result.TraceHash = HashString(TagTrace, []byte(""))
	} else {
		traceLeaves := make([][]byte, len(components.TraceChunks))
		for i, chunk := range components.TraceChunks {
			traceLeaves[i] = Hash(TagTraceChunk, chunk)
		}
		traceRoot := MerkleRoot(traceLeaves)
		result.TraceHash = HashString(TagTrace, traceRoot)
	}

	// 5. Resource hash
	resBytes, err := Canonicalize(components.ResourceUsage)
	if err != nil {
		return nil, fmt.Errorf("meg: canonicalize resource: %w", err)
	}
	result.ResourceHash = HashString(TagResource, resBytes)

	// 6. Output hash
	outBytes, err := Canonicalize(components.OutputPayload)
	if err != nil {
		return nil, fmt.Errorf("meg: canonicalize output: %w", err)
	}
	result.OutputHash = HashString(TagOutput, outBytes)

	// 7. Metadata hash
	metaBytes, err := Canonicalize(components.Metadata)
	if err != nil {
		return nil, fmt.Errorf("meg: canonicalize metadata: %w", err)
	}
	result.MetadataHash = HashString(TagMeta, metaBytes)

	// Build leaf array in fixed order
	leafHashes := []string{
		result.InputHash,
		result.EnvironmentHash,
		result.DependencyHash,
		result.TraceHash,
		result.ResourceHash,
		result.OutputHash,
		result.MetadataHash,
	}
	result.LeafHashes = leafHashes

	// Convert hex strings to bytes for Merkle tree
	leaves := make([][]byte, len(leafHashes))
	for i, h := range leafHashes {
		b, err := hex.DecodeString(h)
		if err != nil {
			return nil, fmt.Errorf("meg: decode leaf hash[%d]: %w", i, err)
		}
		leaves[i] = b
	}

	// Compute Merkle root — this is the ExecutionRootHash
	root := MerkleRoot(leaves)
	result.ExecutionRootHash = hex.EncodeToString(root)

	return result, nil
}

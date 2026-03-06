// Package crypto implements the Merkle Execution Graph (MEG) cryptographic primitives
// for the Deterministic Replay Engine (DRE) 2.0 protocol.
//
// Hash algorithm: BLAKE3 - extremely fast, tree-native, parallelizable,
// with strong cryptographic guarantees. Ideal for large trace hashing.
//
// BLAKE3 produces a 32-byte (256-bit) output by default, which is used
// for all hash computations in the MEG protocol.
package crypto

import (
	"encoding/hex"

	"lukechampine.com/blake3"
)

// Domain separation tags — fixed forever in protocol (DRE/1.0)
const (
	TagInput        = "FX_INPUT"
	TagEnv          = "FX_ENV"
	TagDeps         = "FX_DEPS"
	TagDepNode      = "FX_DEP_NODE"
	TagTrace        = "FX_TRACE"
	TagTraceChunk   = "FX_TRACE_CHUNK"
	TagResource     = "FX_RES"
	TagOutput       = "FX_OUT"
	TagMeta         = "FX_META"
	TagNode         = "FX_NODE"
	TagSyscall      = "FX_SYSCALL"
	TagCert         = "FX_CERT"
	TagReplayProof  = "FX_REPLAY_PROOF"
)

// Hash computes BLAKE3(tag || data) with domain separation.
// The tag is prepended as raw bytes (no length prefix) to ensure
// domain separation between different hash contexts.
//
// BLAKE3 is used as the default hash algorithm per the ExecutionRootHash v1.0 protocol.
func Hash(tag string, data []byte) []byte {
	h := blake3.New(32, nil) // 32-byte digest, unkeyed
	h.Write([]byte(tag))
	h.Write(data)
	return h.Sum(nil)
}

// HashString returns the hex-encoded hash of BLAKE3(tag || data).
func HashString(tag string, data []byte) string {
	return hex.EncodeToString(Hash(tag, data))
}

// MerkleRoot computes the Merkle root of a list of leaf hashes.
// If the number of leaves is odd, the last leaf is duplicated (Bitcoin-style).
// Returns nil for an empty leaf list.
func MerkleRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return nil
	}
	if len(leaves) == 1 {
		return Hash(TagNode, leaves[0])
	}

	// Work on a copy to avoid mutating the input
	current := make([][]byte, len(leaves))
	copy(current, leaves)

	for len(current) > 1 {
		// Pad to even length
		if len(current)%2 != 0 {
			current = append(current, current[len(current)-1])
		}

		next := make([][]byte, len(current)/2)
		for i := 0; i < len(current); i += 2 {
			combined := append(current[i], current[i+1]...)
			next[i/2] = Hash(TagNode, combined)
		}
		current = next
	}

	return current[0]
}

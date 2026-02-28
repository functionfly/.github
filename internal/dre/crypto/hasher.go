// Package crypto implements the Merkle Execution Graph (MEG) cryptographic primitives
// for the Deterministic Replay Engine (DRE) 2.0 protocol.
//
// Hash algorithm: SHA-256 with domain separation (FIPS-compatible default).
// To use BLAKE3, build with the blake3 build tag and ensure lukechampine.com/blake3
// is available: go get lukechampine.com/blake3
package crypto

import (
	"crypto/sha256"
	"encoding/hex"
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

// Hash computes SHA-256(tag || data) with domain separation.
// The tag is prepended as raw bytes (no length prefix) to ensure
// domain separation between different hash contexts.
func Hash(tag string, data []byte) []byte {
	h := sha256.New()
	h.Write([]byte(tag))
	h.Write(data)
	return h.Sum(nil)
}

// HashString returns the hex-encoded hash of SHA-256(tag || data).
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
		return leaves[0]
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

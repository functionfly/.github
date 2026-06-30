package trustapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/bits"
	"time"

	"github.com/google/uuid"
)

// Merkle domain separation prefixes per RFC 6962.
var (
	merkleLeafPrefix = []byte{0x00}
	merkleNodePrefix = []byte{0x01}
)

// MerkleTreeHead represents a signed snapshot of the Merkle log at a point in time.
type MerkleTreeHead struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TreeSize     int64       `json:"tree_size" gorm:"not null;index:idx_merkle_heads_size"`
	RootHash     string      `json:"root_hash" gorm:"size:64;not null"`
	PreviousHash string      `json:"previous_hash,omitempty" gorm:"size:64"`
	Timestamp    time.Time   `json:"timestamp" gorm:"not null;index:idx_merkle_heads_ts"`
	Signature    string      `json:"signature,omitempty" gorm:"size:512"`
	PublicKeyID  string      `json:"public_key_id,omitempty" gorm:"size:100"`
	Metadata     json.RawMessage `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

func (MerkleTreeHead) TableName() string { return "merkle_tree_heads" }

// MerkleNode is a persisted node in the Merkle tree.
// Level 0 = leaves (Index = leaf position). Higher levels = interior.
type MerkleNode struct {
	ID     int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Level  int    `json:"level" gorm:"not null;uniqueIndex:idx_merkle_node_pos,priority:1"`
	Index  int64  `json:"index" gorm:"not null;uniqueIndex:idx_merkle_node_pos,priority:2"`
	Hash   string `json:"hash" gorm:"size:64;not null;uniqueIndex:idx_merkle_node_hash"`
	LeafID string `json:"leaf_id,omitempty" gorm:"size:32;index:idx_merkle_node_leaf"`
}

func (MerkleNode) TableName() string { return "merkle_nodes" }

// MerkleInclusionProof proves a specific leaf is in the tree.
type MerkleInclusionProof struct {
	LeafIndex int64    `json:"leaf_index"`
	LeafHash  string   `json:"leaf_hash"`
	TreeSize  int64    `json:"tree_size"`
	RootHash  string   `json:"root_hash"`
	Path      []string `json:"path"` // Sibling hashes from leaf to root
}

// MerkleConsistencyProof proves the log is append-only between two tree sizes.
type MerkleConsistencyProof struct {
	OldSize int64    `json:"old_size"`
	NewSize int64    `json:"new_size"`
	OldRoot string   `json:"old_root"`
	NewRoot string   `json:"new_root"`
	Path    []string `json:"path"`
}

// ============================================
// Pure hash helpers
// ============================================

// MerkleLeafHash computes the hash of a leaf per RFC 6962 §2.1:
//
//	leaf_hash = SHA256(0x00 || data)
func MerkleLeafHash(data []byte) string {
	h := sha256.New()
	h.Write(merkleLeafPrefix)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// MerkleNodeHash computes the hash of an interior node:
//
//	node_hash = SHA256(0x01 || left_hash || right_hash)
func MerkleNodeHash(leftHex, rightHex string) (string, error) {
	left, err := hex.DecodeString(leftHex)
	if err != nil {
		return "", err
	}
	right, err := hex.DecodeString(rightHex)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(merkleNodePrefix)
	h.Write(left)
	h.Write(right)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ============================================
// Root computation from leaf hashes
// ============================================

// ComputeRoot builds the Merkle root from a slice of leaf hashes.
// Returns "" for empty input.
func ComputeRoot(leaves []string) string {
	if len(leaves) == 0 {
		return ""
	}
	if len(leaves) == 1 {
		return leaves[0]
	}
	current := make([]string, len(leaves))
	copy(current, leaves)
	for len(current) > 1 {
		var next []string
		for i := 0; i < len(current); i += 2 {
			if i+1 < len(current) {
				h, err := MerkleNodeHash(current[i], current[i+1])
				if err != nil {
					return ""
				}
				next = append(next, h)
			} else {
				next = append(next, current[i]) // promote odd node
			}
		}
		current = next
	}
	return current[0]
}

// ============================================
// Inclusion proof (audit path)
// ============================================

// BuildInclusionProof produces the audit path for leaf at `idx` in a tree
// built from `leaves`.  Uses the RFC 6962 algorithm.
func BuildInclusionProof(leaves []string, idx int64) ([]string, error) {
	n := int64(len(leaves))
	if idx < 0 || idx >= n {
		return nil, ErrIndexOutOfRange
	}
	if n == 1 {
		return nil, nil // single leaf — trivial proof
	}

	var path []string
	level := make([]string, n)
	copy(level, leaves)
	position := idx

	for len(level) > 1 {
		sz := int64(len(level))
		k := largestPowerOfTwoLessThan(sz)

		var left, right []string
		var leftIdx int64

		if position < k {
			left = level[:k]
			right = level[k:]
			leftIdx = position
		} else {
			left = level[:k]
			right = level[k:]
			leftIdx = position - k
		}

		if position < k {
			// Node is in left subtree — sibling is root of right subtree
			path = append(path, ComputeRoot(right))
			level = left
			position = leftIdx
		} else {
			// Node is in right subtree — sibling is root of left subtree
			path = append(path, ComputeRoot(left))
			level = right
			position = leftIdx
		}
	}

	return path, nil
}

// VerifyInclusion checks that `leafHash` at `leafIndex` is in a tree of `treeSize`
// whose root is `rootHash`, given the `proof` path.
func VerifyInclusion(leafHash string, leafIndex, treeSize int64, proof []string, rootHash string) bool {
	if leafIndex < 0 || leafIndex >= treeSize {
		return false
	}

	hash := leafHash
	idx := leafIndex
	size := treeSize

	for _, sib := range proof {
		k := largestPowerOfTwoLessThan(size)
		var err error
		if idx < k {
			hash, err = MerkleNodeHash(hash, sib)
			size = k
		} else {
			hash, err = MerkleNodeHash(sib, hash)
			idx -= k
			size -= k
		}
		if err != nil {
			return false
		}
	}

	return hash == rootHash
}

// ============================================
// Consistency proof
// ============================================

// BuildConsistencyProof proves that the first `m` leaves of the current tree
// (size `n`) are identical to the tree at size `m`.
func BuildConsistencyProof(leaves []string, m int64) ([]string, error) {
	n := int64(len(leaves))
	if m < 0 || m > n {
		return nil, ErrIndexOutOfRange
	}
	if m == 0 || m == n {
		return nil, nil
	}

	var proof []string

	// Find the prefix subtree for size m
	k := largestPowerOfTwoLessThan(m)

	// If m is not a power of 2, we need the old root first
	if m != k {
		proof = append(proof, ComputeRoot(leaves[:m]))
	}

	// Walk up the right edge of the old tree, collecting left-sibling hashes
	position := m
	for position > 1 {
		k = largestPowerOfTwoLessThan(position)
		proof = append(proof, ComputeRoot(leaves[position-k:position]))
		position -= k
	}

	return proof, nil
}

// ============================================
// Helpers
// ============================================

// largestPowerOfTwoLessThan returns the largest power of 2 strictly less than n.
// For n=1 returns 0 (no smaller power of 2 exists).
func largestPowerOfTwoLessThan(n int64) int64 {
	if n <= 1 {
		return 0
	}
	return int64(1) << (bits.Len64(uint64(n-1)) - 1)
}

// ErrIndexOutOfRange is returned when a Merkle operation receives an invalid index.
var ErrIndexOutOfRange = fmt.Errorf("index out of range")

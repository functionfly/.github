package trustapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MerkleRepository handles persistence of the Merkle audit trail.
type MerkleRepository struct {
	db *gorm.DB
}

// NewMerkleRepository creates a new Merkle repository.
func NewMerkleRepository(db *gorm.DB) *MerkleRepository {
	return &MerkleRepository{db: db}
}

// AppendLeaf adds an attestation to the Merkle log.
// It computes the leaf hash, persists the leaf node, recomputes affected
// interior nodes, and creates a new signed tree head.
// This MUST be called inside the same transaction as the attestation insert.
func (r *MerkleRepository) AppendLeaf(attestation *TrustAttestation, signer Signer) (*MerkleTreeHead, error) {
	// Serialize attestation data for hashing
	leafData, err := json.Marshal(map[string]interface{}{
		"attestation_id": attestation.AttestationID,
		"function_id":    attestation.FunctionID.String(),
		"type":           attestation.Type,
		"proof_hash":     attestation.ProofHash,
		"code_hash":      attestation.CodeHash,
		"input_hash":     attestation.InputHash,
		"output_hash":    attestation.OutputHash,
		"attested_at":    attestation.AttestedAt.UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal leaf data: %w", err)
	}

	leafHash := MerkleLeafHash(leafData)

	// Get current tree size
	size, err := r.TreeSize()
	if err != nil {
		return nil, fmt.Errorf("get tree size: %w", err)
	}

	// Persist leaf node
	leafNode := &MerkleNode{
		Level:  0,
		Index:  size,
		Hash:   leafHash,
		LeafID: attestation.AttestationID,
	}
	if err := r.db.Create(leafNode).Error; err != nil {
		return nil, fmt.Errorf("insert leaf node: %w", err)
	}

	// Recompute interior nodes up to root
	newSize := size + 1
	if err := r.updateInteriorNodes(size, leafHash); err != nil {
		return nil, fmt.Errorf("update interior nodes: %w", err)
	}

	// Compute new root
	root, err := r.ComputeRoot()
	if err != nil {
		return nil, fmt.Errorf("compute root: %w", err)
	}

	// Get previous tree head hash for chaining
	var prevHash string
	prevHead, err := r.LatestTreeHead()
	if err == nil && prevHead != nil {
		prevHash = treeHeadHash(prevHead)
	}

	// Build tree head
	head := &MerkleTreeHead{
		TreeSize:     newSize,
		RootHash:     root,
		PreviousHash: prevHash,
		Timestamp:    time.Now(),
	}

	// Sign the tree head
	if signer != nil {
		signedPayload := fmt.Sprintf("%d|%s|%d", head.TreeSize, head.RootHash, head.Timestamp.UnixNano())
		sig, err := signer.Sign([]byte(signedPayload))
		if err == nil {
			head.Signature = sig
			head.PublicKeyID = signer.KeyID()
		}
	}

	if err := r.db.Create(head).Error; err != nil {
		return nil, fmt.Errorf("insert tree head: %w", err)
	}

	return head, nil
}

// updateInteriorNodes recomputes interior nodes after appending a leaf at `newLeafIndex`.
func (r *MerkleRepository) updateInteriorNodes(newLeafIndex int64, newLeafHash string) error {
	level := 0
	currentHash := newLeafHash
	currentIndex := newLeafIndex

	for {
		// At each level, if the new node's index is even, it doesn't create
		// a new parent yet (it's the right child of an incomplete pair).
		// If it's odd, it pairs with the left sibling.
		if currentIndex%2 == 1 {
			// Get the left sibling
			leftIdx := currentIndex - 1
			leftNode, err := r.getNode(level, leftIdx)
			if err != nil {
				return fmt.Errorf("get left sibling at level %d idx %d: %w", level, leftIdx, err)
			}

			parentHash, err := MerkleNodeHash(leftNode.Hash, currentHash)
			if err != nil {
				return fmt.Errorf("compute parent hash: %w", err)
			}

			parentIndex := currentIndex / 2
			parentNode := &MerkleNode{
				Level: level + 1,
				Index: parentIndex,
				Hash:  parentHash,
			}

			// Upsert the parent node
			if err := r.upsertNode(parentNode); err != nil {
				return fmt.Errorf("upsert parent node level %d idx %d: %w", level+1, parentIndex, err)
			}

			currentHash = parentHash
			currentIndex = parentIndex
			level++
		} else {
			// Odd position — this node doesn't have a sibling yet.
			// It promotes as-is to the next level (like a lone node).
			parentIndex := currentIndex / 2
			parentNode := &MerkleNode{
				Level: level + 1,
				Index: parentIndex,
				Hash:  currentHash,
			}
			if err := r.upsertNode(parentNode); err != nil {
				return fmt.Errorf("promote node level %d idx %d: %w", level+1, parentIndex, err)
			}
			break // No further propagation needed for promoted nodes
		}
	}

	return nil
}

// ComputeRoot reads all leaf hashes and computes the root.
func (r *MerkleRepository) ComputeRoot() (string, error) {
	size, err := r.TreeSize()
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", nil
	}

	// Check if we have a cached root at the top level
	topLevel, err := r.maxLevel()
	if err != nil {
		return "", err
	}

	node, err := r.getNode(topLevel, 0)
	if err == nil && node != nil {
		return node.Hash, nil
	}

	// Fallback: rebuild from leaves
	leaves, err := r.allLeaves()
	if err != nil {
		return "", err
	}
	hashes := make([]string, len(leaves))
	for i, l := range leaves {
		hashes[i] = l.Hash
	}
	return ComputeRoot(hashes), nil
}

// TreeSize returns the number of leaves in the Merkle tree.
func (r *MerkleRepository) TreeSize() (int64, error) {
	var count int64
	err := r.db.Model(&MerkleNode{}).Where("level = 0").Count(&count).Error
	return count, err
}

// LatestTreeHead returns the most recent tree head, or nil if none exists.
func (r *MerkleRepository) LatestTreeHead() (*MerkleTreeHead, error) {
	var head MerkleTreeHead
	err := r.db.Order("tree_size DESC").First(&head).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &head, nil
}

// GetTreeHeadBySize returns the tree head at a specific tree size.
func (r *MerkleRepository) GetTreeHeadBySize(size int64) (*MerkleTreeHead, error) {
	var head MerkleTreeHead
	err := r.db.Where("tree_size = ?", size).First(&head).Error
	if err != nil {
		return nil, err
	}
	return &head, nil
}

// GetInclusionProof builds an inclusion proof for a leaf at the given index.
func (r *MerkleRepository) GetInclusionProof(leafIndex int64) (*MerkleInclusionProof, error) {
	size, err := r.TreeSize()
	if err != nil {
		return nil, err
	}
	if leafIndex < 0 || leafIndex >= size {
		return nil, ErrIndexOutOfRange
	}

	leaves, err := r.allLeaves()
	if err != nil {
		return nil, err
	}
	hashes := make([]string, len(leaves))
	for i, l := range leaves {
		hashes[i] = l.Hash
	}

	path, err := BuildInclusionProof(hashes, leafIndex)
	if err != nil {
		return nil, err
	}

	root := ComputeRoot(hashes)

	return &MerkleInclusionProof{
		LeafIndex: leafIndex,
		LeafHash:  hashes[leafIndex],
		TreeSize:  size,
		RootHash:  root,
		Path:      path,
	}, nil
}

// GetConsistencyProof proves the log is append-only between two tree sizes.
func (r *MerkleRepository) GetConsistencyProof(oldSize int64) (*MerkleConsistencyProof, error) {
	newSize, err := r.TreeSize()
	if err != nil {
		return nil, err
	}
	if oldSize < 0 || oldSize > newSize {
		return nil, ErrIndexOutOfRange
	}

	leaves, err := r.allLeaves()
	if err != nil {
		return nil, err
	}
	hashes := make([]string, len(leaves))
	for i, l := range leaves {
		hashes[i] = l.Hash
	}

	proof, err := BuildConsistencyProof(hashes, oldSize)
	if err != nil {
		return nil, err
	}

	oldRoot := ComputeRoot(hashes[:oldSize])
	newRoot := ComputeRoot(hashes)

	return &MerkleConsistencyProof{
		OldSize: oldSize,
		NewSize: newSize,
		OldRoot: oldRoot,
		NewRoot: newRoot,
		Path:    proof,
	}, nil
}

// ============================================
// Internal helpers
// ============================================

func (r *MerkleRepository) getNode(level int, index int64) (*MerkleNode, error) {
	var node MerkleNode
	err := r.db.Where("level = ? AND index = ?", level, index).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *MerkleRepository) upsertNode(node *MerkleNode) error {
	var existing MerkleNode
	err := r.db.Where("level = ? AND index = ?", node.Level, node.Index).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(node).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&existing).Update("hash", node.Hash).Error
}

func (r *MerkleRepository) allLeaves() ([]MerkleNode, error) {
	var leaves []MerkleNode
	err := r.db.Where("level = 0").Order("index ASC").Find(&leaves).Error
	return leaves, err
}

func (r *MerkleRepository) maxLevel() (int, error) {
	var result struct{ Max int }
	err := r.db.Model(&MerkleNode{}).Select("COALESCE(MAX(level), 0) as max").Scan(&result).Error
	return result.Max, err
}

// treeHeadHash produces a deterministic hash of a tree head for chaining.
func treeHeadHash(h *MerkleTreeHead) string {
	data := fmt.Sprintf("%d|%s|%d", h.TreeSize, h.RootHash, h.Timestamp.UnixNano())
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

// GetLeafIndexByAttestationID finds the leaf index for a given attestation ID.
func (r *MerkleRepository) GetLeafIndexByAttestationID(attestationID string) (int64, error) {
	var node MerkleNode
	err := r.db.Where("level = 0 AND leaf_id = ?", attestationID).First(&node).Error
	if err != nil {
		return -1, err
	}
	return node.Index, nil
}

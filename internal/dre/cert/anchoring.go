// Package cert implements the FXCERT execution certificate protocol.
// This file implements blockchain anchoring for execution certificates.
package cert

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Supported blockchain chains for anchoring.
const (
	ChainEthereum  = "ethereum"
	ChainPolygon   = "polygon"
	ChainArbitrum  = "arbitrum"
	ChainOptimism  = "optimism"
	ChainBase      = "base"
	ChainAvalanche = "avalanche"
)

// DefaultChain is the recommended chain for anchoring: lowest cost among major EVM L2s
// (typically $0.001–0.02 per simple tx). Use when no chain preference is specified.
const DefaultChain = ChainBase

// SupportedChains returns the list of supported blockchain chains, ordered by
// cost-efficiency (cheapest first): Base, Polygon, Arbitrum, Optimism, Avalanche, Ethereum.
var SupportedChains = []string{
	ChainBase,      // Best cost/safety: L2, ~$0.001–0.02/tx
	ChainPolygon,   // Very cheap sidechain, ~$0.001–0.01/tx
	ChainArbitrum,  // L2, ~$0.03–0.06/tx
	ChainOptimism,  // L2, ~$0.06–0.10/tx
	ChainAvalanche, // L1, variable
	ChainEthereum,  // L1, highest cost; use for max assurance only
}

// AnchorReceipt contains the blockchain anchoring confirmation.
type AnchorReceipt struct {
	Chain          string    `json:"chain"`
	BlockNumber    int64     `json:"block_number"`
	TxHash         string    `json:"tx_hash"`
	MerkleRoot     string    `json:"merkle_root"`     // The anchored merkle root
	AnchorHash     string    `json:"anchor_hash"`     // Hash of the anchoring data
	AnchoredAt     time.Time `json:"anchored_at"`     // Timestamp of anchoring
	Confirmations  int64     `json:"confirmations"`   // Number of block confirmations
	GasUsed        uint64    `json:"gas_used"`        // Gas used for the transaction
	ContractAddress string   `json:"contract_address"` // Address of the anchoring contract
}

// AnchorRequest contains the data needed to anchor a certificate.
type AnchorRequest struct {
	CertificateID   string   `json:"certificate_id"`
	ExecutionRootHash string `json:"execution_root_hash"`
	Chain           string   `json:"chain"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AnchoringService defines the interface for blockchain anchoring operations.
type AnchoringService interface {
	// Anchor submits the execution root hash to a blockchain.
	Anchor(ctx context.Context, req *AnchorRequest) (*AnchorReceipt, error)

	// VerifyAnchor verifies an existing anchor on the blockchain.
	VerifyAnchor(ctx context.Context, receipt *AnchorReceipt) (bool, error)

	// GetMerkleProof returns the Merkle proof for an anchored hash.
	GetMerkleProof(ctx context.Context, chain, txHash string) (*MerkleProof, error)

	// GetBlockNumber returns the current block number for a chain.
	GetBlockNumber(ctx context.Context, chain string) (int64, error)
}

// MerkleProof represents a Merkle proof for verification.
type MerkleProof struct {
	BlockNumber int64    `json:"block_number"`
	TxHash      string   `json:"tx_hash"`
	RootHash    string   `json:"root_hash"`
	Proof       []string `json:"proof"`
	LeafHash    string   `json:"leaf_hash"`
}

// EthereumAnchoringService implements AnchoringService for Ethereum-compatible chains.
type EthereumAnchoringService struct {
	// RPC endpoints for different chains
	rpcEndpoints map[string]string
	// Contract addresses for anchoring (set during initialization)
	contractAddresses map[string]string
	// Private key for signing transactions (should be from secure storage)
	privateKey string
	// clientCache caches ethclient by chain (used by production implementation)
	clientCache map[string]interface{}
	clientMu    sync.Mutex
}

// NewEthereumAnchoringService creates a new Ethereum anchoring service.
func NewEthereumAnchoringService(rpcEndpoints map[string]string) *EthereumAnchoringService {
	return &EthereumAnchoringService{
		rpcEndpoints:     rpcEndpoints,
		contractAddresses: make(map[string]string),
	}
}

// SetContractAddress sets the contract address for a specific chain.
func (s *EthereumAnchoringService) SetContractAddress(chain, address string) {
	s.contractAddresses[chain] = address
}

// Anchor submits the execution root hash to the blockchain.
// Production: requires SetSigningKey, and for the chosen chain an RPC endpoint and contract address.
func (s *EthereumAnchoringService) Anchor(ctx context.Context, req *AnchorRequest) (*AnchorReceipt, error) {
	chain := req.Chain
	if chain == "" {
		chain = DefaultChain
	}
	if !IsChainSupported(chain) {
		return nil, fmt.Errorf("cert: unsupported chain: %s", chain)
	}
	if req.ExecutionRootHash == "" {
		return nil, fmt.Errorf("cert: execution root hash is required")
	}

	contractAddr := s.contractAddresses[chain]
	if contractAddr == "" {
		return nil, fmt.Errorf("cert: no contract address for chain %q (call SetContractAddress)", chain)
	}
	if s.rpcEndpoints[chain] == "" {
		return nil, fmt.Errorf("cert: no RPC endpoint for chain %q", chain)
	}

	return s.anchorOnChain(ctx, chain, req.ExecutionRootHash, contractAddr)
}

// VerifyAnchor verifies that an anchor exists on the blockchain.
func (s *EthereumAnchoringService) VerifyAnchor(ctx context.Context, receipt *AnchorReceipt) (bool, error) {
	if receipt == nil {
		return false, fmt.Errorf("cert: nil receipt")
	}
	if receipt.TxHash == "" {
		return false, fmt.Errorf("cert: empty transaction hash")
	}
	if receipt.MerkleRoot == "" {
		return false, fmt.Errorf("cert: empty merkle root")
	}
	if s.rpcEndpoints[receipt.Chain] == "" {
		return false, fmt.Errorf("cert: no RPC endpoint for chain %q", receipt.Chain)
	}
	return s.verifyAnchorOnChain(ctx, receipt)
}

// GetMerkleProof returns the Merkle proof for an anchored hash from chain logs.
func (s *EthereumAnchoringService) GetMerkleProof(ctx context.Context, chain, txHash string) (*MerkleProof, error) {
	if s.rpcEndpoints[chain] == "" {
		return nil, fmt.Errorf("cert: no RPC endpoint for chain %q", chain)
	}
	return s.getMerkleProofOnChain(ctx, chain, txHash)
}

// GetBlockNumber returns the current block number for a chain.
func (s *EthereumAnchoringService) GetBlockNumber(ctx context.Context, chain string) (int64, error) {
	if s.rpcEndpoints[chain] == "" {
		return 0, fmt.Errorf("cert: no RPC endpoint for chain %q", chain)
	}
	return s.getBlockNumberOnChain(ctx, chain)
}

// IsChainSupported checks if a chain is supported.
func IsChainSupported(chain string) bool {
	for _, c := range SupportedChains {
		if c == chain {
			return true
		}
	}
	return false
}

// generateMockTxHash generates a mock transaction hash for testing.
func generateMockTxHash(rootHash string) string {
	data := []byte("anchor:" + rootHash)
	hash := sha256.Sum256(data)
	return "0x" + hex.EncodeToString(hash[:])
}

// generateAnchorHash generates a hash of the anchoring data.
func generateAnchorHash(rootHash string) string {
	data := []byte("fxcert:" + rootHash)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// AnchorCertificate updates an FXCert with anchoring information.
func AnchorCertificate(cert *FXCert, receipt *AnchorReceipt) error {
	if cert == nil {
		return fmt.Errorf("cert: nil certificate")
	}

	if receipt == nil {
		return fmt.Errorf("cert: nil anchor receipt")
	}

	// Update the anchoring section
	cert.Anchoring = AnchoringSection{
		Anchored:          true,
		AnchorChain:       receipt.Chain,
		AnchorBlockNumber: receipt.BlockNumber,
		AnchorTxHash:      receipt.TxHash,
		AnchorMerkleRoot: receipt.MerkleRoot,
		AnchoredAt:        receipt.AnchoredAt.Format(time.RFC3339),
	}

	// Note: We don't recompute the certificate hash here because
	// anchoring is optional and doesn't affect the execution proof.
	// In a stricter model, you might want to include anchoring in the hash.

	return nil
}

// IsAnchored returns true if the certificate has been anchored.
func IsAnchored(cert *FXCert) bool {
	return cert != nil && cert.Anchoring.Anchored
}

// GetAnchorInfo returns anchoring information for a certificate.
func GetAnchorInfo(cert *FXCert) map[string]interface{} {
	if cert == nil {
		return nil
	}

	return map[string]interface{}{
		"anchored":          cert.Anchoring.Anchored,
		"anchor_chain":      cert.Anchoring.AnchorChain,
		"anchor_block":      cert.Anchoring.AnchorBlockNumber,
		"anchor_tx_hash":    cert.Anchoring.AnchorTxHash,
		"anchor_merkle_root": cert.Anchoring.AnchorMerkleRoot,
		"anchored_at":       cert.Anchoring.AnchoredAt,
	}
}

// UnmarshalJSON implements custom JSON unmarshaling for AnchorReceipt.
func (r *AnchorReceipt) UnmarshalJSON(data []byte) error {
	type Alias AnchorReceipt
	aux := &struct {
		AnchoredAt string `json:"anchored_at"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.AnchoredAt != "" {
		var err error
		r.AnchoredAt, err = time.Parse(time.RFC3339, aux.AnchoredAt)
		if err != nil {
			return fmt.Errorf("cert: parse anchored_at: %w", err)
		}
	}

	return nil
}

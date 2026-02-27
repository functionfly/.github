// Package cert implements the FXCERT execution certificate protocol.
// An FXCERT is a legal-grade, cryptographically signed artifact that proves
// a specific function execution occurred with specific inputs and produced
// specific outputs in a sealed deterministic environment.
package cert

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
)

// FXCert is the legal-grade execution certificate.
// It is a self-contained, verifiable proof of a function execution.
type FXCert struct {
	FXCertVersion string            `json:"fxcert_version"` // "1.0"
	CertificateID string            `json:"certificate_id"` // "fxc_01H..."
	Execution     ExecutionSection  `json:"execution"`
	Capsule       CapsuleSection    `json:"capsule"`
	Integrity     IntegritySection  `json:"integrity"`
	Trust         TrustSection      `json:"trust"`
	Signatures    SignatureSection  `json:"signatures"`
	Anchoring     AnchoringSection  `json:"anchoring"`
	ReplayCert    *ReplayCertSection `json:"replay_certification,omitempty"`
}

// ExecutionSection identifies the execution context.
type ExecutionSection struct {
	ExecutionID      string `json:"execution_id"`
	FunctionID       string `json:"function_id"`        // "fx://acme/compute/1.2.4"
	OwnerID          string `json:"owner_id"`
	CallerID         string `json:"caller_id"`
	NodeID           string `json:"node_id"`
	Region           string `json:"region"`
	TimestampVirtual string `json:"timestamp_virtual"` // deterministic (from capsule seed)
	TimestampRealUTC string `json:"timestamp_real_utc"` // informational only
	ProtocolVersion  string `json:"protocol_version"`  // "dre/1.0"
}

// CapsuleSection describes the deterministic execution environment.
type CapsuleSection struct {
	CapsuleDescriptorHash string `json:"capsule_descriptor_hash"`
	DeterminismTier       string `json:"determinism_tier"` // "full"|"lite"
	ProtocolVersion       string `json:"protocol_version"` // "dcc/1.0"
}

// IntegritySection contains all cryptographic hashes for verification.
type IntegritySection struct {
	ExecutionRootHash string `json:"execution_root_hash"`
	InputHash         string `json:"input_hash"`
	EnvironmentHash   string `json:"environment_hash"`
	DependencyHash    string `json:"dependency_hash"`
	TraceHash         string `json:"trace_hash"`
	ResourceHash      string `json:"resource_hash"`
	OutputHash        string `json:"output_hash"`
	MetadataHash      string `json:"metadata_hash"`
	// CertificateHash = H("FX_CERT" || canonical_cert_without_this_field)
	CertificateHash string `json:"certificate_hash"`
}

// TrustSection contains the trust score snapshot at time of certification.
type TrustSection struct {
	TrustScore              float64 `json:"trust_score"`
	DeterminismScore        float64 `json:"determinism_score"`
	ReplayConsistencyScore  float64 `json:"replay_consistency_score"`
	DriftIncidentsTotal     int     `json:"drift_incidents_total"`
	VerifiedExecutionsTotal int64   `json:"verified_executions_total"`
}

// SignatureSection contains cryptographic signatures over the execution root hash.
type SignatureSection struct {
	NodeSignature     *Signature `json:"node_signature"`
	PlatformSignature *Signature `json:"platform_signature,omitempty"`
}

// Signature is a single cryptographic signature.
type Signature struct {
	Algorithm string `json:"algorithm"` // "Ed25519"
	PublicKey string `json:"public_key"` // base64-encoded
	Signature string `json:"signature"`  // base64-encoded
}

// AnchoringSection contains optional blockchain anchoring data.
type AnchoringSection struct {
	Anchored          bool   `json:"anchored"`
	AnchorChain       string `json:"anchor_chain,omitempty"`
	AnchorBlockNumber int64  `json:"anchor_block_number,omitempty"`
	AnchorTxHash      string `json:"anchor_tx_hash,omitempty"`
	AnchorMerkleRoot  string `json:"anchor_merkle_root,omitempty"`
	AnchoredAt        string `json:"anchored_at,omitempty"` // ISO-8601 UTC
}

// ReplayCertSection is populated after a successful replay verification.
type ReplayCertSection struct {
	ReplayRootHash  string `json:"replay_root_hash"`
	ReplayNodeID    string `json:"replay_node_id"`
	ReplayTimestamp string `json:"replay_timestamp"` // ISO-8601 UTC
	RootsMatch      bool   `json:"roots_match"`
}

// CertLevel controls how much data is included in the certificate.
type CertLevel string

const (
	// CertLevelLite includes root hash + node signature + minimal trust data.
	// Use for high-volume, low-cost scenarios.
	CertLevelLite CertLevel = "lite"
	// CertLevelStandard includes all component hashes + capsule hash + trust snapshot.
	// This is the default for all executions.
	CertLevelStandard CertLevel = "standard"
	// CertLevelLegal includes everything in standard plus full trace Merkle root,
	// dependency root, platform signature, and replay certification.
	// Use for enterprise compliance, audit, and legal proceedings.
	CertLevelLegal CertLevel = "legal_grade"
)

// Generate creates a signed FXCERT from a completed MEG result.
// The nodeKey is used to sign the execution_root_hash with Ed25519.
func Generate(
	meg *drecrypto.MEGResult,
	exec ExecutionSection,
	capsule CapsuleSection,
	trust TrustSection,
	level CertLevel,
	nodeKey ed25519.PrivateKey,
) (*FXCert, error) {
	if meg == nil {
		return nil, fmt.Errorf("cert: meg result is nil")
	}

	cert := &FXCert{
		FXCertVersion: "1.0",
		CertificateID: generateCertID(meg.ExecutionRootHash),
		Execution:     exec,
		Capsule:       capsule,
		Trust:         trust,
		Anchoring:     AnchoringSection{Anchored: false},
	}

	// Populate integrity section based on cert level
	cert.Integrity = IntegritySection{
		ExecutionRootHash: meg.ExecutionRootHash,
		InputHash:         meg.InputHash,
		EnvironmentHash:   meg.EnvironmentHash,
		DependencyHash:    meg.DependencyHash,
		ResourceHash:      meg.ResourceHash,
		OutputHash:        meg.OutputHash,
		MetadataHash:      meg.MetadataHash,
	}

	// Include trace hash only for standard and legal grade
	if level == CertLevelStandard || level == CertLevelLegal {
		cert.Integrity.TraceHash = meg.TraceHash
	}

	// Compute certificate hash: H("FX_CERT" || canonical_cert_without_certificate_hash)
	certHash, err := computeCertHash(cert)
	if err != nil {
		return nil, fmt.Errorf("cert: compute cert hash: %w", err)
	}
	cert.Integrity.CertificateHash = certHash

	// Sign the execution root hash with the node's Ed25519 key
	if nodeKey != nil {
		rootHashBytes := []byte(meg.ExecutionRootHash)
		sig := ed25519.Sign(nodeKey, rootHashBytes)
		pubKey := nodeKey.Public().(ed25519.PublicKey)

		cert.Signatures = SignatureSection{
			NodeSignature: &Signature{
				Algorithm: "Ed25519",
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(sig),
			},
		}
	}

	return cert, nil
}

// Verify validates all hashes and signatures in an FXCERT.
// Returns nil if the certificate is valid.
func Verify(cert *FXCert, nodePublicKey ed25519.PublicKey) error {
	if cert == nil {
		return fmt.Errorf("cert: nil certificate")
	}

	// Step 1: Recompute ExecutionRootHash from component hashes via MerkleRoot
	leafHashes := []string{
		cert.Integrity.InputHash,
		cert.Integrity.EnvironmentHash,
		cert.Integrity.DependencyHash,
		cert.Integrity.TraceHash,
		cert.Integrity.ResourceHash,
		cert.Integrity.OutputHash,
		cert.Integrity.MetadataHash,
	}

	// Convert hex strings to bytes
	leaves := make([][]byte, len(leafHashes))
	for i, h := range leafHashes {
		if h == "" {
			// Empty hash — use zero bytes
			leaves[i] = make([]byte, 32)
			continue
		}
		b, err := hexDecode(h)
		if err != nil {
			return fmt.Errorf("cert: decode leaf hash[%d]: %w", i, err)
		}
		leaves[i] = b
	}

	computedRoot := drecrypto.MerkleRoot(leaves)
	computedRootHex := hexEncode(computedRoot)

	if computedRootHex != cert.Integrity.ExecutionRootHash {
		return fmt.Errorf("cert: execution root hash mismatch: computed %s, got %s",
			computedRootHex, cert.Integrity.ExecutionRootHash)
	}

	// Step 2: Recompute CertificateHash
	savedHash := cert.Integrity.CertificateHash
	cert.Integrity.CertificateHash = ""
	computedCertHash, err := computeCertHash(cert)
	cert.Integrity.CertificateHash = savedHash
	if err != nil {
		return fmt.Errorf("cert: recompute cert hash: %w", err)
	}

	if computedCertHash != savedHash {
		return fmt.Errorf("cert: certificate hash mismatch")
	}

	// Step 3: Verify Ed25519 signature
	if nodePublicKey != nil && cert.Signatures.NodeSignature != nil {
		sig, err := base64.StdEncoding.DecodeString(cert.Signatures.NodeSignature.Signature)
		if err != nil {
			return fmt.Errorf("cert: decode node signature: %w", err)
		}
		rootHashBytes := []byte(cert.Integrity.ExecutionRootHash)
		if !ed25519.Verify(nodePublicKey, rootHashBytes, sig) {
			return fmt.Errorf("cert: node signature verification failed")
		}
	}

	return nil
}

// computeCertHash computes H("FX_CERT" || canonical_cert_without_certificate_hash).
// The cert.Integrity.CertificateHash field must be empty when calling this.
func computeCertHash(cert *FXCert) (string, error) {
	b, err := json.Marshal(cert)
	if err != nil {
		return "", fmt.Errorf("marshal cert for hash: %w", err)
	}

	canonical, err := drecrypto.Canonicalize(json.RawMessage(b))
	if err != nil {
		return "", fmt.Errorf("canonicalize cert: %w", err)
	}

	return drecrypto.HashString(drecrypto.TagCert, canonical), nil
}

// generateCertID creates a deterministic certificate ID from the execution root hash.
func generateCertID(executionRootHash string) string {
	// Use first 26 chars of the root hash as the ID suffix
	suffix := executionRootHash
	if len(suffix) > 26 {
		suffix = suffix[:26]
	}
	return "fxc_" + suffix
}

// hexDecode decodes a hex string to bytes.
func hexDecode(s string) ([]byte, error) {
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		var v byte
		for j := 0; j < 2; j++ {
			c := s[i+j]
			switch {
			case c >= '0' && c <= '9':
				v = v<<4 | (c - '0')
			case c >= 'a' && c <= 'f':
				v = v<<4 | (c - 'a' + 10)
			case c >= 'A' && c <= 'F':
				v = v<<4 | (c - 'A' + 10)
			default:
				return nil, fmt.Errorf("invalid hex char %q at position %d", c, i+j)
			}
		}
		b[i/2] = v
	}
	return b, nil
}

// hexEncode encodes bytes to a hex string.
func hexEncode(b []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(b)*2)
	for i, v := range b {
		result[i*2] = hexChars[v>>4]
		result[i*2+1] = hexChars[v&0xf]
	}
	return string(result)
}

// Now is a helper that returns the current time in ISO-8601 UTC format.
func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

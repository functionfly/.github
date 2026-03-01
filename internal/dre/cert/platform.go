// Package cert implements the FXCERT execution certificate protocol.
// This file implements platform signatures for enterprise tier certificates.
package cert

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PlatformKeyManager manages platform signing keys for enterprise certificates.
type PlatformKeyManager interface {
	// GetPlatformKey returns the current platform public key.
	GetPlatformKey() (ed25519.PublicKey, error)

	// Sign signs data with the platform key.
	Sign(data []byte) ([]byte, error)

	// Verify verifies a platform signature.
	Verify(data, signature []byte) bool

	// GetKeyID returns the current key identifier.
	GetKeyID() string

	// RotateKey rotates to a new platform key.
	RotateKey(newKey ed25519.PrivateKey) error

	// GetKeyMetadata returns metadata about the current key.
	GetKeyMetadata() (*PlatformKeyMetadata, error)
}

// PlatformKeyMetadata contains metadata about a platform signing key.
type PlatformKeyMetadata struct {
	KeyID        string    `json:"key_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Algorithm    string    `json:"algorithm"`
	IsActive     bool      `json:"is_active"`
	RotationPolicy string  `json:"rotation_policy"`
}

// PlatformSigner handles enterprise platform signatures.
type PlatformSigner struct {
	keyManager PlatformKeyManager
	auditLog   AuditLogger
	mu         sync.RWMutex
}

// AuditLogger defines the interface for logging signature operations.
type AuditLogger interface {
	// Log records an audit event.
	Log(event AuditEvent)
}

// AuditEvent represents a signature operation audit event.
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	Operation   string                 `json:"operation"` // "sign", "verify", "rotate"
	KeyID       string                 `json:"key_id"`
	CertificateID string               `json:"certificate_id,omitempty"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewPlatformSigner creates a new platform signer.
func NewPlatformSigner(keyManager PlatformKeyManager, auditLog AuditLogger) *PlatformSigner {
	return &PlatformSigner{
		keyManager: keyManager,
		auditLog:   auditLog,
	}
}

// Sign signs a certificate with the platform key.
func (s *PlatformSigner) Sign(cert *FXCert) (*Signature, error) {
	if cert == nil {
		return nil, fmt.Errorf("cert: nil certificate")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the platform key
	pubKey, err := s.keyManager.GetPlatformKey()
	if err != nil {
		s.logAudit(AuditEvent{
			Operation:      "sign",
			CertificateID: cert.CertificateID,
			Success:        false,
			Error:          err.Error(),
		})
		return nil, fmt.Errorf("cert: get platform key: %w", err)
	}

	// Prepare data to sign: certificate hash + timestamp to prevent replay
	signatureData := []byte(cert.Integrity.CertificateHash + fmt.Sprintf("%d", time.Now().Unix()))

	// Sign the data
	sig, err := s.keyManager.Sign(signatureData)
	if err != nil {
		s.logAudit(AuditEvent{
			Operation:      "sign",
			CertificateID: cert.CertificateID,
			Success:        false,
			Error:          err.Error(),
		})
		return nil, fmt.Errorf("cert: sign certificate: %w", err)
	}

	// Create signature struct
	signature := &Signature{
		Algorithm: "Ed25519",
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(sig),
	}

	s.logAudit(AuditEvent{
		Operation:      "sign",
		CertificateID:  cert.CertificateID,
		Success:        true,
		KeyID:          s.keyManager.GetKeyID(),
	})

	return signature, nil
}

// SignCertificate adds a platform signature to a certificate.
func (s *PlatformSigner) SignCertificate(cert *FXCert) error {
	sig, err := s.Sign(cert)
	if err != nil {
		return err
	}

	cert.Signatures.PlatformSignature = sig
	return nil
}

// Verify verifies a platform signature on a certificate.
func (s *PlatformSigner) Verify(cert *FXCert) (bool, error) {
	if cert == nil {
		return false, fmt.Errorf("cert: nil certificate")
	}

	if cert.Signatures.PlatformSignature == nil {
		return false, fmt.Errorf("cert: no platform signature present")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Decode the signature
	sigBytes, err := base64.StdEncoding.DecodeString(cert.Signatures.PlatformSignature.Signature)
	if err != nil {
		s.logAudit(AuditEvent{
			Operation:      "verify",
			CertificateID: cert.CertificateID,
			Success:        false,
			Error:          "invalid signature encoding",
		})
		return false, fmt.Errorf("cert: decode signature: %w", err)
	}

	// Get the platform public key
	pubKey, err := s.keyManager.GetPlatformKey()
	if err != nil {
		return false, fmt.Errorf("cert: get platform key: %w", err)
	}

	// Verify the signature
	// Note: In production, you might want to verify against multiple historical keys
	signatureValid := ed25519.Verify(pubKey, []byte(cert.Integrity.CertificateHash), sigBytes)

	s.logAudit(AuditEvent{
		Operation:      "verify",
		CertificateID: cert.CertificateID,
		Success:        signatureValid,
		KeyID:          s.keyManager.GetKeyID(),
	})

	if !signatureValid {
		return false, fmt.Errorf("cert: platform signature verification failed")
	}

	return true, nil
}

// GetPlatformKey returns the current platform public key.
func (s *PlatformSigner) GetPlatformKey() (ed25519.PublicKey, error) {
	return s.keyManager.GetPlatformKey()
}

// logAudit logs an audit event if an audit logger is configured.
func (s *PlatformSigner) logAudit(event AuditEvent) {
	if s.auditLog != nil {
		event.Timestamp = time.Now().UTC()
		s.auditLog.Log(event)
	}
}

// FilePlatformKeyManager implements PlatformKeyManager using files on disk.
// This is suitable for development and testing; production should use HSM.
type FilePlatformKeyManager struct {
	keyPath     string
	metadataPath string
	currentKey  ed25519.PrivateKey
	keyMetadata *PlatformKeyMetadata
	mu          sync.RWMutex
}

// NewFilePlatformKeyManager creates a new file-based platform key manager.
func NewFilePlatformKeyManager(keyPath, metadataPath string) (*FilePlatformKeyManager, error) {
	mgr := &FilePlatformKeyManager{
		keyPath:      keyPath,
		metadataPath: metadataPath,
	}

	// Try to load existing key
	if err := mgr.loadKey(); err != nil {
		// Generate new key if none exists
		if err := mgr.generateNewKey(); err != nil {
			return nil, fmt.Errorf("cert: generate key: %w", err)
		}
	}

	return mgr, nil
}

// loadKey loads the platform key from disk.
func (m *FilePlatformKeyManager) loadKey() error {
	// Load private key
	keyData, err := os.ReadFile(m.keyPath)
	if err != nil {
		return fmt.Errorf("cert: read key file: %w", err)
	}

	keyBytes, err := base64.StdEncoding.DecodeString(string(keyData))
	if err != nil {
		return fmt.Errorf("cert: decode key: %w", err)
	}

	if len(keyBytes) != ed25519.PrivateKeySize {
		return fmt.Errorf("cert: invalid key size")
	}

	m.currentKey = ed25519.PrivateKey(keyBytes)

	// Load metadata if it exists
	if metadataData, err := os.ReadFile(m.metadataPath); err == nil {
		m.keyMetadata = &PlatformKeyMetadata{}
		if err := json.Unmarshal(metadataData, m.keyMetadata); err != nil {
			return fmt.Errorf("cert: parse metadata: %w", err)
		}
	}

	return nil
}

// generateNewKey generates a new platform key.
func (m *FilePlatformKeyManager) generateNewKey() error {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return fmt.Errorf("cert: generate key: %w", err)
	}

	m.currentKey = privKey
	m.keyMetadata = &PlatformKeyMetadata{
		KeyID:         fmt.Sprintf("key_%d", time.Now().Unix()),
		CreatedAt:     time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(365 * 24 * time.Hour), // 1 year
		Algorithm:    "Ed25519",
		IsActive:     true,
		RotationPolicy: "annual",
	}

	// Save the key
	keyData := base64.StdEncoding.EncodeToString(m.currentKey)
	if err := os.WriteFile(m.keyPath, []byte(keyData), 0600); err != nil {
		return fmt.Errorf("cert: write key file: %w", err)
	}

	// Save metadata
	metadataData, _ := json.MarshalIndent(m.keyMetadata, "", "  ")
	if err := os.WriteFile(m.metadataPath, metadataData, 0644); err != nil {
		return fmt.Errorf("cert: write metadata file: %w", err)
	}

	_ = pubKey // Unused in this implementation
	return nil
}

// GetPlatformKey returns the current platform public key.
func (m *FilePlatformKeyManager) GetPlatformKey() (ed25519.PublicKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentKey == nil {
		return nil, fmt.Errorf("cert: no platform key loaded")
	}

	return m.currentKey.Public().(ed25519.PublicKey), nil
}

// Sign signs data with the platform key.
func (m *FilePlatformKeyManager) Sign(data []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentKey == nil {
		return nil, fmt.Errorf("cert: no platform key loaded")
	}

	return ed25519.Sign(m.currentKey, data), nil
}

// Verify verifies a platform signature.
func (m *FilePlatformKeyManager) Verify(data, signature []byte) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentKey == nil {
		return false
	}

	pubKey := m.currentKey.Public().(ed25519.PublicKey)
	return ed25519.Verify(pubKey, data, signature)
}

// GetKeyID returns the current key identifier.
func (m *FilePlatformKeyManager) GetKeyID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.keyMetadata != nil {
		return m.keyMetadata.KeyID
	}
	return ""
}

// RotateKey rotates to a new platform key.
func (m *FilePlatformKeyManager) RotateKey(newKey ed25519.PrivateKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Archive old key metadata if exists
	if m.keyMetadata != nil {
		oldMetadataPath := m.metadataPath + ".old"
		oldData, _ := json.MarshalIndent(m.keyMetadata, "", "  ")
		os.WriteFile(oldMetadataPath, oldData, 0644)
	}

	// Set new key
	m.currentKey = newKey
	m.keyMetadata = &PlatformKeyMetadata{
		KeyID:         fmt.Sprintf("key_%d", time.Now().Unix()),
		CreatedAt:     time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(365 * 24 * time.Hour),
		Algorithm:    "Ed25519",
		IsActive:     true,
		RotationPolicy: "annual",
	}

	// Save new key
	keyData := base64.StdEncoding.EncodeToString(m.currentKey)
	if err := os.WriteFile(m.keyPath, []byte(keyData), 0600); err != nil {
		return fmt.Errorf("cert: write key file: %w", err)
	}

	// Save metadata
	metadataData, _ := json.MarshalIndent(m.keyMetadata, "", "  ")
	if err := os.WriteFile(m.metadataPath, metadataData, 0644); err != nil {
		return fmt.Errorf("cert: write metadata file: %w", err)
	}

	return nil
}

// GetKeyMetadata returns metadata about the current key.
func (m *FilePlatformKeyManager) GetKeyMetadata() (*PlatformKeyMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.keyMetadata == nil {
		return nil, fmt.Errorf("cert: no key metadata available")
	}

	return m.keyMetadata, nil
}

// PlatformSignatureVerifier verifies platform signatures for external use.
type PlatformSignatureVerifier struct {
	// Known platform public keys (keyID -> publicKey)
	knownKeys map[string]ed25519.PublicKey
}

// NewPlatformSignatureVerifier creates a new platform signature verifier.
func NewPlatformSignatureVerifier() *PlatformSignatureVerifier {
	return &PlatformSignatureVerifier{
		knownKeys: make(map[string]ed25519.PublicKey),
	}
}

// RegisterKey registers a known platform public key.
func (v *PlatformSignatureVerifier) RegisterKey(keyID string, pubKey ed25519.PublicKey) {
	v.knownKeys[keyID] = pubKey
}

// Verify verifies a platform signature using known keys.
func (v *PlatformSignatureVerifier) Verify(cert *FXCert) (bool, error) {
	if cert == nil || cert.Signatures.PlatformSignature == nil {
		return false, fmt.Errorf("cert: no platform signature present")
	}

	// Get the key ID from the public key in the signature
	sigPubKey, err := base64.StdEncoding.DecodeString(cert.Signatures.PlatformSignature.PublicKey)
	if err != nil {
		return false, fmt.Errorf("cert: decode public key: %w", err)
	}

	// Try to find matching known key
	var matchingKey ed25519.PublicKey
	for _, key := range v.knownKeys {
		if string(key) == string(sigPubKey) {
			matchingKey = key
			break
		}
	}

	if matchingKey == nil {
		return false, fmt.Errorf("cert: unknown platform key")
	}

	// Decode signature
	sigBytes, err := base64.StdEncoding.DecodeString(cert.Signatures.PlatformSignature.Signature)
	if err != nil {
		return false, fmt.Errorf("cert: decode signature: %w", err)
	}

	// Verify
	return ed25519.Verify(matchingKey, []byte(cert.Integrity.CertificateHash), sigBytes), nil
}

// GetCertificateTrustLevel returns the trust level of a certificate based on its signatures.
func GetCertificateTrustLevel(cert *FXCert) string {
	if cert == nil {
		return "unknown"
	}

	hasNodeSig := cert.Signatures.NodeSignature != nil
	hasPlatformSig := cert.Signatures.PlatformSignature != nil
	anchored := cert.Anchoring.Anchored
	replayCertified := cert.ReplayCert != nil && cert.ReplayCert.RootsMatch

	// Determine trust level based on available guarantees
	if hasNodeSig && hasPlatformSig && anchored && replayCertified {
		return "enterprise" // Highest trust: all verifications present
	}

	if hasNodeSig && hasPlatformSig && anchored {
		return "verified" // High trust: platform verified and anchored
	}

	if hasNodeSig && hasPlatformSig {
		return "standard" // Standard trust: platform signed
	}

	if hasNodeSig {
		return "basic" // Basic trust: only node signed
	}

	return "lite" // Minimal trust: unsigned
}

// ValidateCertificateChain validates the signature chain of a certificate.
func ValidateCertificateChain(cert *FXCert, nodePubKey, platformPubKey ed25519.PublicKey) error {
	if cert == nil {
		return fmt.Errorf("cert: nil certificate")
	}

	// Validate node signature
	if cert.Signatures.NodeSignature != nil && nodePubKey != nil {
		nodeSig, err := base64.StdEncoding.DecodeString(cert.Signatures.NodeSignature.Signature)
		if err != nil {
			return fmt.Errorf("cert: decode node signature: %w", err)
		}

		if !ed25519.Verify(nodePubKey, []byte(cert.Integrity.ExecutionRootHash), nodeSig) {
			return fmt.Errorf("cert: node signature invalid")
		}
	}

	// Validate platform signature
	if cert.Signatures.PlatformSignature != nil && platformPubKey != nil {
		platSig, err := base64.StdEncoding.DecodeString(cert.Signatures.PlatformSignature.Signature)
		if err != nil {
			return fmt.Errorf("cert: decode platform signature: %w", err)
		}

		if !ed25519.Verify(platformPubKey, []byte(cert.Integrity.CertificateHash), platSig) {
			return fmt.Errorf("cert: platform signature invalid")
		}
	}

	return nil
}

// KeyIDFromSignature extracts the key ID from a signature's public key.
func KeyIDFromSignature(sig *Signature) (string, error) {
	if sig == nil {
		return "", fmt.Errorf("cert: nil signature")
	}

	// The key ID is typically derived from the public key
	// In a real implementation, this would look up the key in a registry
	pubKeyBytes, err := base64.StdEncoding.DecodeString(sig.PublicKey)
	if err != nil {
		return "", fmt.Errorf("cert: decode public key: %w", err)
	}

	// Create a simple key ID from the public key hash
	keyID := fmt.Sprintf("key_%x", pubKeyBytes[:8])
	return keyID, nil
}

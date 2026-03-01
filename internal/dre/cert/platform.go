// Package cert implements the FXCERT execution certificate protocol.
// This file implements platform signatures for enterprise tier certificates.
package cert

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
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

// PlatformKeysForVerificationProvider is an optional interface. When a
// PlatformKeyManager implements it, PlatformSigner.Verify will try the
// current key and each historical key in order until one succeeds.
type PlatformKeysForVerificationProvider interface {
	// GetPlatformKeysForVerification returns the current key first, then
	// any historical keys still valid for verification (e.g. after rotation).
	GetPlatformKeysForVerification() ([]ed25519.PublicKey, error)
}

// PlatformKeyMetadata contains metadata about a platform signing key.
type PlatformKeyMetadata struct {
	KeyID          string    `json:"key_id"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Algorithm      string    `json:"algorithm"`
	IsActive       bool      `json:"is_active"`
	RotationPolicy string    `json:"rotation_policy"`
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
	Timestamp     time.Time              `json:"timestamp"`
	Operation     string                 `json:"operation"` // "sign", "verify", "rotate"
	KeyID         string                 `json:"key_id"`
	CertificateID string                 `json:"certificate_id,omitempty"`
	Success       bool                   `json:"success"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
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
			Operation:     "sign",
			CertificateID: cert.CertificateID,
			Success:       false,
			Error:         err.Error(),
		})
		return nil, fmt.Errorf("cert: get platform key: %w", err)
	}

	// Prepare data to sign: certificate hash + timestamp to prevent replay
	signatureData := []byte(cert.Integrity.CertificateHash + fmt.Sprintf("%d", time.Now().Unix()))

	// Sign the data
	sig, err := s.keyManager.Sign(signatureData)
	if err != nil {
		s.logAudit(AuditEvent{
			Operation:     "sign",
			CertificateID: cert.CertificateID,
			Success:       false,
			Error:         err.Error(),
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
		Operation:     "sign",
		CertificateID: cert.CertificateID,
		Success:       true,
		KeyID:         s.keyManager.GetKeyID(),
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
			Operation:     "verify",
			CertificateID: cert.CertificateID,
			Success:       false,
			Error:         "invalid signature encoding",
		})
		return false, fmt.Errorf("cert: decode signature: %w", err)
	}

	// Build list of keys to try: current + historical if supported
	var keysToTry []ed25519.PublicKey
	if provider, ok := s.keyManager.(PlatformKeysForVerificationProvider); ok {
		keysToTry, err = provider.GetPlatformKeysForVerification()
		if err != nil {
			return false, fmt.Errorf("cert: get keys for verification: %w", err)
		}
	} else {
		pubKey, err := s.keyManager.GetPlatformKey()
		if err != nil {
			return false, fmt.Errorf("cert: get platform key: %w", err)
		}
		keysToTry = []ed25519.PublicKey{pubKey}
	}

	dataToVerify := []byte(cert.Integrity.CertificateHash)
	var signatureValid bool
	for _, pubKey := range keysToTry {
		if ed25519.Verify(pubKey, dataToVerify, sigBytes) {
			signatureValid = true
			break
		}
	}

	s.logAudit(AuditEvent{
		Operation:     "verify",
		CertificateID: cert.CertificateID,
		Success:       signatureValid,
		KeyID:         s.keyManager.GetKeyID(),
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

// maxHistoricalKeys limits how many rotated keys are kept for verification.
const maxHistoricalKeys = 32

// FilePlatformKeyManager implements PlatformKeyManager using files on disk.
// This is suitable for development and testing; production should use HSM.
type FilePlatformKeyManager struct {
	keyPath              string
	metadataPath         string
	historyPath          string
	currentKey           ed25519.PrivateKey
	keyMetadata          *PlatformKeyMetadata
	historicalPublicKeys []ed25519.PublicKey
	mu                   sync.RWMutex
}

// NewFilePlatformKeyManager creates a new file-based platform key manager.
func NewFilePlatformKeyManager(keyPath, metadataPath string) (*FilePlatformKeyManager, error) {
	mgr := &FilePlatformKeyManager{
		keyPath:      keyPath,
		metadataPath: metadataPath,
		historyPath:  metadataPath + ".history",
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

	// Load historical public keys for verification after rotation
	m.loadHistoricalKeys()

	return nil
}

// loadHistoricalKeys reads the history file and populates historicalPublicKeys.
func (m *FilePlatformKeyManager) loadHistoricalKeys() {
	data, err := os.ReadFile(m.historyPath)
	if err != nil {
		m.historicalPublicKeys = nil
		return
	}
	var b64Keys []string
	if err := json.Unmarshal(data, &b64Keys); err != nil {
		m.historicalPublicKeys = nil
		return
	}
	keys := make([]ed25519.PublicKey, 0, len(b64Keys))
	for _, b64 := range b64Keys {
		dec, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || len(dec) != ed25519.PublicKeySize {
			continue
		}
		keys = append(keys, ed25519.PublicKey(dec))
	}
	// Keep only the most recent entries
	if len(keys) > maxHistoricalKeys {
		keys = keys[len(keys)-maxHistoricalKeys:]
	}
	m.historicalPublicKeys = keys
}

// saveHistoricalKeys writes historicalPublicKeys to the history file.
func (m *FilePlatformKeyManager) saveHistoricalKeys() error {
	b64 := make([]string, len(m.historicalPublicKeys))
	for i, k := range m.historicalPublicKeys {
		b64[i] = base64.StdEncoding.EncodeToString(k)
	}
	data, err := json.Marshal(b64)
	if err != nil {
		return err
	}
	return os.WriteFile(m.historyPath, data, 0600)
}

// generateNewKey generates a new platform key.
func (m *FilePlatformKeyManager) generateNewKey() error {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return fmt.Errorf("cert: generate key: %w", err)
	}

	m.currentKey = privKey
	m.keyMetadata = &PlatformKeyMetadata{
		KeyID:          fmt.Sprintf("key_%d", time.Now().Unix()),
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(365 * 24 * time.Hour), // 1 year
		Algorithm:      "Ed25519",
		IsActive:       true,
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

	// Append current key's public part to history before replacing (for verification of old certs)
	if m.currentKey != nil {
		pub := m.currentKey.Public().(ed25519.PublicKey)
		m.historicalPublicKeys = append(m.historicalPublicKeys, pub)
		if len(m.historicalPublicKeys) > maxHistoricalKeys {
			m.historicalPublicKeys = m.historicalPublicKeys[len(m.historicalPublicKeys)-maxHistoricalKeys:]
		}
		_ = m.saveHistoricalKeys()
	}

	// Archive old key metadata if exists
	if m.keyMetadata != nil {
		oldMetadataPath := m.metadataPath + ".old"
		oldData, _ := json.MarshalIndent(m.keyMetadata, "", "  ")
		os.WriteFile(oldMetadataPath, oldData, 0644)
	}

	// Set new key
	m.currentKey = newKey
	m.keyMetadata = &PlatformKeyMetadata{
		KeyID:          fmt.Sprintf("key_%d", time.Now().Unix()),
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(365 * 24 * time.Hour),
		Algorithm:      "Ed25519",
		IsActive:       true,
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

// GetPlatformKeysForVerification implements PlatformKeysForVerificationProvider.
// It returns the current key first, then historical keys (oldest to newest).
func (m *FilePlatformKeyManager) GetPlatformKeysForVerification() ([]ed25519.PublicKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentKey == nil {
		return nil, fmt.Errorf("cert: no platform key loaded")
	}

	currentPub := m.currentKey.Public().(ed25519.PublicKey)
	out := make([]ed25519.PublicKey, 1, 1+len(m.historicalPublicKeys))
	out[0] = currentPub
	out = append(out, m.historicalPublicKeys...)
	return out, nil
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

// PlatformKeyRegistry stores platform public keys by key ID and implements
// KeyIDResolver for lookup by public key. Safe for concurrent use.
type PlatformKeyRegistry struct {
	mu   sync.RWMutex
	keys map[string]ed25519.PublicKey // keyID -> public key
}

// NewPlatformKeyRegistry creates an empty platform key registry.
func NewPlatformKeyRegistry() *PlatformKeyRegistry {
	return &PlatformKeyRegistry{
		keys: make(map[string]ed25519.PublicKey),
	}
}

// Register adds or updates a platform public key for the given key ID.
func (r *PlatformKeyRegistry) Register(keyID string, pubKey ed25519.PublicKey) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.keys == nil {
		r.keys = make(map[string]ed25519.PublicKey)
	}
	r.keys[keyID] = pubKey
}

// Get returns the public key for keyID, or nil and false if not found.
func (r *PlatformKeyRegistry) Get(keyID string) (ed25519.PublicKey, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pub, ok := r.keys[keyID]
	return pub, ok
}

// Remove removes the key for keyID. It is a no-op if the key is not present.
func (r *PlatformKeyRegistry) Remove(keyID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.keys, keyID)
}

// KeyIDs returns a copy of all registered key IDs.
func (r *PlatformKeyRegistry) KeyIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.keys))
	for id := range r.keys {
		ids = append(ids, id)
	}
	return ids
}

// KeyIDFromPublicKey implements KeyIDResolver by looking up the key ID for the
// given public key. Returns the first matching key ID and true, or "" and false.
func (r *PlatformKeyRegistry) KeyIDFromPublicKey(pubKey ed25519.PublicKey) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, p := range r.keys {
		if len(p) == len(pubKey) && string(p) == string(pubKey) {
			return id, true
		}
	}
	return "", false
}

// Verify verifies a certificate's platform signature using any key in the registry.
func (r *PlatformKeyRegistry) Verify(cert *FXCert) (bool, error) {
	if cert == nil || cert.Signatures.PlatformSignature == nil {
		return false, fmt.Errorf("cert: no platform signature present")
	}
	sigPubKey, err := base64.StdEncoding.DecodeString(cert.Signatures.PlatformSignature.PublicKey)
	if err != nil {
		return false, fmt.Errorf("cert: decode public key: %w", err)
	}
	r.mu.RLock()
	var matchingKey ed25519.PublicKey
	for _, p := range r.keys {
		if len(p) == len(sigPubKey) && string(p) == string(sigPubKey) {
			matchingKey = p
			break
		}
	}
	r.mu.RUnlock()
	if matchingKey == nil {
		return false, fmt.Errorf("cert: unknown platform key")
	}
	sigBytes, err := base64.StdEncoding.DecodeString(cert.Signatures.PlatformSignature.Signature)
	if err != nil {
		return false, fmt.Errorf("cert: decode signature: %w", err)
	}
	return ed25519.Verify(matchingKey, []byte(cert.Integrity.CertificateHash), sigBytes), nil
}

// Ensure PlatformKeyRegistry implements KeyIDResolver.
var _ KeyIDResolver = (*PlatformKeyRegistry)(nil)

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

// KeyIDResolver looks up a key ID from a registry by public key.
// Implement this when keys are managed in an HSM or registry and you need
// canonical IDs (e.g. key ARN) instead of a derived value.
type KeyIDResolver interface {
	KeyIDFromPublicKey(pubKey ed25519.PublicKey) (keyID string, ok bool)
}

// KeyIDFromSignature extracts the key ID from a signature's public key.
// It decodes the signature's PublicKey and returns a derived ID (key_%x of first 8 bytes).
func KeyIDFromSignature(sig *Signature) (string, error) {
	return KeyIDFromSignatureWithResolver(sig, nil)
}

// KeyIDFromSignatureWithResolver extracts the key ID from a signature's public key.
// If resolver is not nil and returns ok true for the signature's public key, that
// key ID is returned (e.g. from an HSM key registry). Otherwise a key ID is derived
// from the first 8 bytes of the public key.
func KeyIDFromSignatureWithResolver(sig *Signature, resolver KeyIDResolver) (string, error) {
	if sig == nil {
		return "", fmt.Errorf("cert: nil signature")
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(sig.PublicKey)
	if err != nil {
		return "", fmt.Errorf("cert: decode public key: %w", err)
	}
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return "", fmt.Errorf("cert: invalid public key size")
	}
	pubKey := ed25519.PublicKey(pubKeyBytes)

	if resolver != nil {
		if keyID, ok := resolver.KeyIDFromPublicKey(pubKey); ok {
			return keyID, nil
		}
	}

	keyID := fmt.Sprintf("key_%x", pubKeyBytes[:8])
	return keyID, nil
}

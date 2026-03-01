// Package cert: HSM-backed platform key manager for production use.
// Use this with an HSM backend (e.g. AWS KMS, GCP KMS, PKCS#11) so private
// keys never leave the HSM. Rotation is done by creating a new key in the
// HSM and adding the previous key ID to VerificationKeyIDs.
package cert

import (
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"
)

// HSMBackend performs signing and public-key retrieval using an HSM or
// cloud KMS. The backend must use Ed25519 keys and produce 64-byte raw
// Ed25519 signatures. Callers typically provide an implementation that
// wraps AWS KMS, GCP KMS, Azure Key Vault, or a PKCS#11 module.
type HSMBackend interface {
	// Sign signs data with the key identified by keyID. Returns raw Ed25519
	// signature (64 bytes).
	Sign(keyID string, data []byte) (signature []byte, err error)
	// GetPublicKey returns the Ed25519 public key for the given keyID.
	GetPublicKey(keyID string) (ed25519.PublicKey, error)
}

// HSMPlatformKeyManager implements PlatformKeyManager by delegating to an
// HSMBackend. Suitable for production; private keys never leave the HSM.
//
// Configure ActiveKeyID with the key used for signing. Optionally set
// VerificationKeyIDs to a list of key IDs to try when verifying (current
// first, then historical). If VerificationKeyIDs is empty, only ActiveKeyID
// is used for verification.
type HSMPlatformKeyManager struct {
	backend            HSMBackend
	activeKeyID        string
	verificationKeyIDs []string
	keyMetadata        *PlatformKeyMetadata
	mu                 sync.RWMutex
}

// NewHSMPlatformKeyManager creates an HSM-backed platform key manager.
// activeKeyID is the key used for signing. verificationKeyIDs is the list
// of key IDs to try when verifying signatures (current key first, then
// historical). Pass nil to use only activeKeyID for verification.
func NewHSMPlatformKeyManager(backend HSMBackend, activeKeyID string, verificationKeyIDs []string) *HSMPlatformKeyManager {
	ids := verificationKeyIDs
	if ids == nil {
		ids = []string{activeKeyID}
	} else {
		// Ensure active is first
		hasActive := false
		for _, id := range ids {
			if id == activeKeyID {
				hasActive = true
				break
			}
		}
		if !hasActive {
			ids = append([]string{activeKeyID}, ids...)
		}
	}
	return &HSMPlatformKeyManager{
		backend:            backend,
		activeKeyID:        activeKeyID,
		verificationKeyIDs: ids,
		keyMetadata: &PlatformKeyMetadata{
			KeyID:          activeKeyID,
			CreatedAt:      time.Now().UTC(),
			ExpiresAt:      time.Time{},
			Algorithm:      "Ed25519",
			IsActive:       true,
			RotationPolicy: "manual",
		},
	}
}

// GetPlatformKey returns the current platform public key from the HSM.
func (m *HSMPlatformKeyManager) GetPlatformKey() (ed25519.PublicKey, error) {
	m.mu.RLock()
	keyID := m.activeKeyID
	m.mu.RUnlock()
	return m.backend.GetPublicKey(keyID)
}

// Sign signs data with the active HSM key.
func (m *HSMPlatformKeyManager) Sign(data []byte) ([]byte, error) {
	m.mu.RLock()
	keyID := m.activeKeyID
	m.mu.RUnlock()
	return m.backend.Sign(keyID, data)
}

// Verify verifies a signature using the active key's public key.
func (m *HSMPlatformKeyManager) Verify(data, signature []byte) bool {
	pub, err := m.GetPlatformKey()
	if err != nil {
		return false
	}
	return ed25519.Verify(pub, data, signature)
}

// GetKeyID returns the current key identifier.
func (m *HSMPlatformKeyManager) GetKeyID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeKeyID
}

// RotateKey is not supported for HSM-backed keys. Keys are managed in the HSM;
// to rotate, create a new key in the HSM, then reconfigure this manager with
// the new active key ID and add the previous key ID to VerificationKeyIDs.
func (m *HSMPlatformKeyManager) RotateKey(newKey ed25519.PrivateKey) error {
	_ = newKey
	return fmt.Errorf("cert: HSM key manager does not support RotateKey; rotate by configuring a new active key ID and adding the previous key to verification key IDs")
}

// GetKeyMetadata returns metadata for the current key.
func (m *HSMPlatformKeyManager) GetKeyMetadata() (*PlatformKeyMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.keyMetadata == nil {
		return nil, fmt.Errorf("cert: no key metadata available")
	}
	meta := *m.keyMetadata
	return &meta, nil
}

// GetPlatformKeysForVerification implements PlatformKeysForVerificationProvider.
// It returns public keys for ActiveKeyID first, then each VerificationKeyID.
func (m *HSMPlatformKeyManager) GetPlatformKeysForVerification() ([]ed25519.PublicKey, error) {
	m.mu.RLock()
	ids := make([]string, len(m.verificationKeyIDs))
	copy(ids, m.verificationKeyIDs)
	m.mu.RUnlock()

	out := make([]ed25519.PublicKey, 0, len(ids))
	for _, id := range ids {
		pub, err := m.backend.GetPublicKey(id)
		if err != nil {
			continue
		}
		out = append(out, pub)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("cert: no platform keys available for verification")
	}
	return out, nil
}

// SetActiveKeyID updates the active key ID (e.g. after creating a new key in
// the HSM). Caller should also append the previous key ID to
// VerificationKeyIDs so existing certificates still verify.
func (m *HSMPlatformKeyManager) SetActiveKeyID(keyID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeKeyID = keyID
	m.keyMetadata = &PlatformKeyMetadata{
		KeyID:          keyID,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Time{},
		Algorithm:      "Ed25519",
		IsActive:       true,
		RotationPolicy: "manual",
	}
}

// SetVerificationKeyIDs sets the list of key IDs used for verification (current
// first, then historical). Used after rotation to include the previous key.
func (m *HSMPlatformKeyManager) SetVerificationKeyIDs(ids []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verificationKeyIDs = append([]string(nil), ids...)
}

// Ensure HSMPlatformKeyManager implements both interfaces at compile time.
var (
	_ PlatformKeyManager                  = (*HSMPlatformKeyManager)(nil)
	_ PlatformKeysForVerificationProvider = (*HSMPlatformKeyManager)(nil)
)

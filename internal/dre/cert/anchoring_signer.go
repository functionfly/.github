// Package cert: signer implementations for Ethereum anchoring.
// Provides SoftwareSigner for development and HSM-backed signers for production.
package cert

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signer signs Ethereum transactions for anchoring.
type Signer interface {
	// PublicKey returns the public key of the signing key.
	PublicKey() (*ecdsa.PublicKey, error)
	// SignTransaction signs an Ethereum transaction.
	SignTransaction(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
}

// SoftwareSigner signs using a raw ECDSA private key (hex).
// Suitable for development and testing; production should use HSM.
type SoftwareSigner struct {
	privateKey string
}

// NewSoftwareSigner creates a SoftwareSigner from a hex-encoded private key.
func NewSoftwareSigner(hexKey string) (*SoftwareSigner, error) {
	hexKey = strings.TrimPrefix(strings.TrimSpace(hexKey), "0x")
	if hexKey == "" {
		return nil, fmt.Errorf("cert: empty private key")
	}
	_, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("cert: invalid private key: %w", err)
	}
	return &SoftwareSigner{privateKey: hexKey}, nil
}

// PublicKey returns the public key of the signing key.
func (s *SoftwareSigner) PublicKey() (*ecdsa.PublicKey, error) {
	key, err := crypto.HexToECDSA(s.privateKey)
	if err != nil {
		return nil, err
	}
	return &key.PublicKey, nil
}

// SignTransaction signs an Ethereum transaction.
func (s *SoftwareSigner) SignTransaction(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	key, err := crypto.HexToECDSA(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("cert: invalid private key: %w", err)
	}
	signer := types.LatestSignerForChainID(chainID)
	return types.SignTx(tx, signer, key)
}

// HSMSigner signs Ethereum transactions using an HSM or cloud KMS backend.
// Private keys never leave the HSM.
type HSMSigner struct {
	backend ETHHSMBackend
	keyID   string
}

// ETHHSMBackend performs signing and public-key retrieval using an HSM or
// cloud KMS. Implementations include AWS KMS, GCP KMS, Azure Key Vault,
// or a PKCS#11 module.
type ETHHSMBackend interface {
	// Sign signs the transaction digest with the key identified by keyID.
	// Returns the raw signature bytes (65 bytes: r || s || v).
	Sign(keyID string, digest []byte) (signature []byte, err error)
	// PublicKey returns the ECDSA public key for the given keyID.
	PublicKey(keyID string) (*ecdsa.PublicKey, error)
}

// NewHSMSigner creates an HSMSigner using the given backend and keyID.
func NewHSMSigner(backend ETHHSMBackend, keyID string) *HSMSigner {
	return &HSMSigner{backend: backend, keyID: keyID}
}

// PublicKey returns the public key from the HSM.
func (s *HSMSigner) PublicKey() (*ecdsa.PublicKey, error) {
	return s.backend.PublicKey(s.keyID)
}

// SignTransaction signs an Ethereum transaction using the HSM.
// The transaction is signed by computing its digest and delegating to the HSM.
func (s *HSMSigner) SignTransaction(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	digest := tx.Hash()
	signature, err := s.backend.Sign(s.keyID, digest.Bytes())
	if err != nil {
		return nil, fmt.Errorf("cert: HSM sign failed: %w", err)
	}
	if len(signature) != 65 {
		return nil, fmt.Errorf("cert: HSM signature must be 65 bytes, got %d", len(signature))
	}
	return tx.WithSignature(types.LatestSignerForChainID(chainID), signature)
}

// Ensure interfaces are satisfied at compile time.
var (
	_ Signer = (*SoftwareSigner)(nil)
	_ Signer = (*HSMSigner)(nil)
)

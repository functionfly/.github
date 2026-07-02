package trustapi

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

// SignatureAlgorithm identifies the cryptographic algorithm used for signing.
type SignatureAlgorithm string

const (
	AlgEd25519 SignatureAlgorithm = "Ed25519"
	AlgECDSA   SignatureAlgorithm = "ECDSA-P256"
	AlgRSA     SignatureAlgorithm = "RSA-PSS-SHA256"
)

// Signer is the interface that all attestation signing backends implement.
// Backends: software (Ed25519/ECDSA), PKCS#11 HSM, AWS KMS, GCP Cloud KMS.
type Signer interface {
	// Sign produces a hex-encoded signature over data.
	Sign(data []byte) (string, error)

	// Verify checks a hex-encoded signature against the signer's public key.
	Verify(data []byte, sigHex string) (bool, error)

	// PublicKeyHex returns the hex-encoded public key for external verification.
	PublicKeyHex() string

	// KeyID returns a stable identifier for the signing key.
	KeyID() string

	// Algorithm returns the signature algorithm in use.
	Algorithm() SignatureAlgorithm

	// SignAttestation computes proof hash, signs it, and populates Signature + PublicKeyID
	// on the attestation. This is the high-level convenience method.
	SignAttestation(att *TrustAttestation) error

	// VerifyAttestationSignature verifies the signature on an attestation
	// using this signer's public key.
	VerifyAttestationSignature(att *TrustAttestation) (bool, error)
}

// CloseableSigner is an optional interface for signers that need cleanup
type CloseableSigner interface {
	Signer
	io.Closer
}

// ecdsaSig is the ASN.1 structure for ECDSA signatures.
type ecdsaSig struct {
	R, S *big.Int
}

// VerifyAttestationSignatureWithKey verifies an attestation signature using
// a raw public key (hex) and algorithm identifier. Works regardless of which
// backend produced the signature.
func VerifyAttestationSignatureWithKey(att *TrustAttestation, pubKeyHex string, alg SignatureAlgorithm) (bool, error) {
	if att.Signature == "" {
		return false, nil
	}
	if !att.VerifyIntegrity() {
		return false, nil
	}

	sigBytes, err := hex.DecodeString(att.Signature)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}
	pubBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return false, fmt.Errorf("decode public key: %w", err)
	}

	switch alg {
	case AlgEd25519:
		if len(pubBytes) != ed25519.PublicKeySize {
			return false, fmt.Errorf("ed25519 public key must be %d bytes, got %d", ed25519.PublicKeySize, len(pubBytes))
		}
		return ed25519.Verify(ed25519.PublicKey(pubBytes), []byte(att.ProofHash), sigBytes), nil

	case AlgECDSA:
		x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
		if x == nil {
			return false, fmt.Errorf("invalid ECDSA public key")
		}
		pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
		var ecdsaSig ecdsaSig
		if _, err := asn1.Unmarshal(sigBytes, &ecdsaSig); err != nil {
			return false, fmt.Errorf("unmarshal ECDSA signature: %w", err)
		}
		hash := sha256.Sum256([]byte(att.ProofHash))
		return ecdsa.Verify(pub, hash[:], ecdsaSig.R, ecdsaSig.S), nil

	default:
		return false, fmt.Errorf("unsupported algorithm: %s", alg)
	}
}

// GenerateSourceDataHash produces a SHA-256 hex hash of arbitrary source data.
func GenerateSourceDataHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

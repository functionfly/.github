package trustapi

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
)

const (
	signingKeyDir  = "/etc/functionfly/keys"
	signingKeyFile = "attestation_signing_key.pem"
	keyIDPrefix    = "ff_att_v"
)

var (
	globalSigner   Signer
	signerOnce     sync.Once
	signerBackend  string
)

// SignerConfig controls which signing backend to use.
// Read from environment variables at startup.
type SignerConfig struct {
	// Backend: "software" (default), "pkcs11", "awskms"
	Backend string

	// Software backend fields
	Algorithm  SignatureAlgorithm // "Ed25519" (default), "ECDSA-P256"
	KeyFile    string             // PEM key file path
	KeyDir     string             // Directory for key storage

	// PKCS#11 backend fields
	PKCS11LibPath string // Path to PKCS#11 shared library (.so/.dylib)
	PKCS11SlotID  int    // Slot ID (0 = first available)
	PKCS11Label   string // Key label in the HSM
	PKCS11Pin     string // PIN for the HSM token

	// AWS KMS backend fields
	AWSCMKID    string // KMS Customer Master Key ID (arn or alias)
	AWSRegion   string // AWS region
	AWSEndpoint string // Custom endpoint (for LocalStack/testing)
}

// configFromEnv builds a SignerConfig from environment variables.
func configFromEnv() SignerConfig {
	cfg := SignerConfig{
		Backend:      envOrDefault("ATTESTATION_SIGNER_BACKEND", "software"),
		Algorithm:    SignatureAlgorithm(envOrDefault("ATTESTATION_SIGNER_ALGORITHM", string(AlgEd25519))),
		KeyDir:       envOrDefault("ATTESTATION_KEY_DIR", signingKeyDir),
		KeyFile:      envOrDefault("ATTESTATION_KEY_FILE", signingKeyFile),
		PKCS11LibPath: os.Getenv("PKCS11_LIBRARY_PATH"),
		PKCS11Label:   envOrDefault("PKCS11_KEY_LABEL", "ff-attestation"),
		PKCS11Pin:     os.Getenv("PKCS11_PIN"),
		AWSCMKID:     os.Getenv("AWS_KMS_CMK_ID"),
		AWSRegion:    envOrDefault("AWS_KMS_REGION", "us-east-1"),
		AWSEndpoint:  os.Getenv("AWS_KMS_ENDPOINT"),
	}
	if v := os.Getenv("PKCS11_SLOT_ID"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.PKCS11SlotID)
	}
	return cfg
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// GetSigner returns the global attestation signer singleton.
// On first call it reads env vars to determine which backend to use,
// then initializes the appropriate signer.
func GetSigner() Signer {
	signerOnce.Do(func() {
		cfg := configFromEnv()
		s, err := newSignerFromConfig(cfg)
		if err != nil {
			// Log warning and fall back to ephemeral software signer
			fmt.Fprintf(os.Stderr, "attestation: signer init failed (%v), using ephemeral Ed25519\n", err)
			s = newEphemeralSoftwareSigner(AlgEd25519)
		}
		globalSigner = s
		signerBackend = cfg.Backend
		if signerBackend == "" {
			signerBackend = "software"
		}
	})
	return globalSigner
}

// SetSigner overrides the global signer (for testing or programmatic setup).
func SetSigner(s Signer) {
	globalSigner = s
	signerOnce = sync.Once{}
}

// ResetSigner resets the singleton so the next GetSigner() call re-initializes
// from environment variables. Used during key rotation.
func ResetSigner() {
	CloseSigner()
	globalSigner = nil
	signerOnce = sync.Once{}
	signerBackend = ""
}

// GetSignerBackend returns the backend name (software, pkcs11, awskms).
func GetSignerBackend() string {
	if signerBackend == "" {
		GetSigner() // ensure initialized
	}
	return signerBackend
}

// CloseSigner closes the global signer if it implements CloseableSigner
func CloseSigner() {
	if globalSigner == nil {
		return
	}
	if cs, ok := globalSigner.(CloseableSigner); ok {
		if err := cs.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "attestation: signer close error: %v\n", err)
		}
	}
}

// newSignerFromConfig is the factory that dispatches to the correct backend.
func newSignerFromConfig(cfg SignerConfig) (Signer, error) {
	switch cfg.Backend {
	case "software", "":
		return newSoftwareSigner(cfg)
	case "pkcs11":
		return newPKCS11Signer(cfg)
	case "awskms":
		return newAWSSigner(cfg)
	default:
		return nil, fmt.Errorf("unknown signer backend: %q", cfg.Backend)
	}
}

// ============================================
// Software Signer (Ed25519 or ECDSA-P256)
// ============================================

// SoftwareSigner implements Signer using in-process keys.
// Ed25519 keys are persisted to disk as PEM. ECDSA keys likewise.
type SoftwareSigner struct {
	algorithm SignatureAlgorithm
	keyID     string

	// Ed25519 fields (when algorithm == AlgEd25519)
	edPriv ed25519.PrivateKey
	edPub  ed25519.PublicKey

	// ECDSA fields (when algorithm == AlgECDSA)
	ecPriv *ecdsa.PrivateKey
	ecPub  *ecdsa.PublicKey
}

// newSoftwareSigner loads or generates a persistent key pair on disk.
func newSoftwareSigner(cfg SignerConfig) (*SoftwareSigner, error) {
	if err := os.MkdirAll(cfg.KeyDir, 0700); err != nil {
		return nil, fmt.Errorf("create key dir: %w", err)
	}
	keyPath := cfg.KeyDir + "/" + cfg.KeyFile

	// Try loading existing key
	if data, err := os.ReadFile(keyPath); err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err == nil {
				return signerFromPrivateKey(key, cfg.Algorithm)
			}
		}
	}

	// Generate new key pair
	return generateAndPersist(keyPath, cfg.Algorithm)
}

// signerFromPrivateKey wraps a parsed PKCS#8 key into a SoftwareSigner.
func signerFromPrivateKey(key interface{}, alg SignatureAlgorithm) (*SoftwareSigner, error) {
	switch k := key.(type) {
	case ed25519.PrivateKey:
		pub := k.Public().(ed25519.PublicKey)
		h := sha256.Sum256(pub)
		return &SoftwareSigner{
			algorithm: AlgEd25519,
			keyID:     keyIDPrefix + hex.EncodeToString(h[:8]),
			edPriv:    k,
			edPub:     pub,
		}, nil
	case *ecdsa.PrivateKey:
		pubBytes := elliptic.Marshal(k.Curve, k.PublicKey.X, k.PublicKey.Y)
		h := sha256.Sum256(pubBytes)
		return &SoftwareSigner{
			algorithm: AlgECDSA,
			keyID:     keyIDPrefix + hex.EncodeToString(h[:8]),
			ecPriv:    k,
			ecPub:     &k.PublicKey,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %T", key)
	}
}

// generateAndPersist creates a new key and writes it to disk as PEM.
func generateAndPersist(path string, alg SignatureAlgorithm) (*SoftwareSigner, error) {
	var (
		privKey interface{}
		pubKey  []byte
		err     error
	)

	switch alg {
	case AlgECDSA:
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate ECDSA key: %w", err)
		}
		privKey = priv
		pubKey = elliptic.Marshal(priv.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	default: // AlgEd25519
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate Ed25519 key: %w", err)
		}
		privKey = priv
		pubKey = pub
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	pemBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}
	if err := os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0600); err != nil {
		return nil, fmt.Errorf("write key file: %w", err)
	}

	h := sha256.Sum256(pubKey)
	signer := &SoftwareSigner{
		algorithm: alg,
		keyID:     keyIDPrefix + hex.EncodeToString(h[:8]),
	}
	switch alg {
	case AlgECDSA:
		signer.ecPriv = privKey.(*ecdsa.PrivateKey)
		signer.ecPub = &signer.ecPriv.PublicKey
	default:
		signer.edPriv = privKey.(ed25519.PrivateKey)
		signer.edPub = signer.edPriv.Public().(ed25519.PublicKey)
	}
	return signer, nil
}

// newEphemeralSoftwareSigner creates a signer with a non-persisted key.
func newEphemeralSoftwareSigner(alg SignatureAlgorithm) *SoftwareSigner {
	switch alg {
	case AlgECDSA:
		priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		pubBytes := elliptic.Marshal(priv.Curve, priv.PublicKey.X, priv.PublicKey.Y)
		h := sha256.Sum256(pubBytes)
		return &SoftwareSigner{
			algorithm: AlgECDSA,
			keyID:     keyIDPrefix + hex.EncodeToString(h[:8]),
			ecPriv:    priv,
			ecPub:     &priv.PublicKey,
		}
	default:
		pub, priv, _ := ed25519.GenerateKey(rand.Reader)
		h := sha256.Sum256(pub)
		return &SoftwareSigner{
			algorithm: AlgEd25519,
			keyID:     keyIDPrefix + hex.EncodeToString(h[:8]),
			edPriv:    priv,
			edPub:     pub,
		}
	}
}

func (s *SoftwareSigner) Algorithm() SignatureAlgorithm { return s.algorithm }
func (s *SoftwareSigner) KeyID() string                 { return s.keyID }

func (s *SoftwareSigner) PublicKeyHex() string {
	switch s.algorithm {
	case AlgECDSA:
		return hex.EncodeToString(elliptic.Marshal(s.ecPub.Curve, s.ecPub.X, s.ecPub.Y))
	default:
		return hex.EncodeToString(s.edPub)
	}
}

func (s *SoftwareSigner) Sign(data []byte) (string, error) {
	switch s.algorithm {
	case AlgECDSA:
		hash := sha256.Sum256(data)
		r, sv, err := ecdsa.Sign(rand.Reader, s.ecPriv, hash[:])
		if err != nil {
			return "", err
		}
		sig, err := asn1.Marshal(ecdsaSig{R: r, S: sv})
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(sig), nil
	default:
		sig := ed25519.Sign(s.edPriv, data)
		return hex.EncodeToString(sig), nil
	}
}

func (s *SoftwareSigner) Verify(data []byte, sigHex string) (bool, error) {
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}
	switch s.algorithm {
	case AlgECDSA:
		hash := sha256.Sum256(data)
		var ecdsaSig ecdsaSig
		if _, err := asn1.Unmarshal(sig, &ecdsaSig); err != nil {
			return false, fmt.Errorf("unmarshal ECDSA signature: %w", err)
		}
		return ecdsa.Verify(s.ecPub, hash[:], ecdsaSig.R, ecdsaSig.S), nil
	default:
		return ed25519.Verify(s.edPub, data, sig), nil
	}
}

func (s *SoftwareSigner) SignAttestation(att *TrustAttestation) error {
	att.ProofHash = att.CalculateProofHash()
	sig, err := s.Sign([]byte(att.ProofHash))
	if err != nil {
		return fmt.Errorf("sign attestation: %w", err)
	}
	att.Signature = sig
	att.PublicKeyID = s.keyID
	return nil
}

func (s *SoftwareSigner) VerifyAttestationSignature(att *TrustAttestation) (bool, error) {
	if att.Signature == "" {
		return false, nil
	}
	if !att.VerifyIntegrity() {
		return false, nil
	}
	return s.Verify([]byte(att.ProofHash), att.Signature)
}

// Ensure interface compliance at compile time.
var _ Signer = (*SoftwareSigner)(nil)

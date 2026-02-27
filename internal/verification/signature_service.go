package verification

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// SignatureService handles digital signatures for function content verification
type SignatureService struct {
	repo *registry.RegistryRepository
}

// NewSignatureService creates a new signature service
func NewSignatureService(repo *registry.RegistryRepository) *SignatureService {
	return &SignatureService{repo: repo}
}

// SignFunction signs a function version's content
func (s *SignatureService) SignFunction(functionVersionID uuid.UUID, signerID string, privateKeyPEM string, algorithm string) (*registry.RegistryFunctionSignature, error) {
	// Get the function version to sign
	version, err := s.repo.GetFunctionVersion(functionVersionID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get function version: %w", err)
	}

	// Calculate content hash to sign
	contentHash, err := s.calculateContentHash(version)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate content hash: %w", err)
	}

	// Parse private key
	privateKey, err := s.parsePrivateKey(privateKeyPEM, algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Sign the content hash
	signature, keyID, err := s.signContent(contentHash, privateKey, algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to sign content: %w", err)
	}

	// Create signature record
	sig := &registry.RegistryFunctionSignature{
		FunctionVersionID: functionVersionID,
		Algorithm:         algorithm,
		KeyID:             keyID,
		Signature:         base64.StdEncoding.EncodeToString(signature),
		SignedBy:          signerID,
		SignedAt:          time.Now(),
		IsValid:           true, // Initially assume valid
	}

	// Store signature
	if err := s.repo.CreateFunctionSignature(sig); err != nil {
		return nil, fmt.Errorf("failed to store signature: %w", err)
	}

	return sig, nil
}

// VerifySignature verifies a function signature
func (s *SignatureService) VerifySignature(signatureID uuid.UUID, publicKeyPEM string) error {
	// Get signature
	sig, err := s.getSignatureByID(signatureID)
	if err != nil {
		return fmt.Errorf("failed to get signature: %w", err)
	}

	// Get function version
	version, err := s.repo.GetFunctionVersion(sig.FunctionVersionID, "")
	if err != nil {
		return fmt.Errorf("failed to get function version: %w", err)
	}

	// Calculate current content hash
	contentHash, err := s.calculateContentHash(version)
	if err != nil {
		return fmt.Errorf("failed to calculate content hash: %w", err)
	}

	// Parse public key
	publicKey, err := s.parsePublicKey(publicKeyPEM, sig.Algorithm)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Decode signature
	signatureBytes, err := base64.StdEncoding.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature
	valid, err := s.verifyContent(contentHash, signatureBytes, publicKey, sig.Algorithm)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// Update signature status
	updates := map[string]interface{}{
		"verified_at":       time.Now(),
		"is_valid":          valid,
		"verification_error": nil,
	}

	if !valid {
		updates["verification_error"] = "signature verification failed"
	}

	if err := s.updateSignature(signatureID, updates); err != nil {
		return fmt.Errorf("failed to update signature status: %w", err)
	}

	return nil
}

// calculateContentHash calculates a hash of the function content for signing
func (s *SignatureService) calculateContentHash(version *registry.RegistryFunctionVersion) ([]byte, error) {
	hasher := sha256.New()

	// Hash manifest
	if version.Manifest != nil {
		hasher.Write(version.Manifest)
	}

	// Hash WASM binary
	if version.WasmBinary != nil {
		hasher.Write(version.WasmBinary)
	}

	// Hash source hash if available
	if version.SourceHash.Valid {
		hasher.Write([]byte(version.SourceHash.String))
	}

	return hasher.Sum(nil), nil
}

// parsePrivateKey parses a private key from PEM format
func (s *SignatureService) parsePrivateKey(pemData, algorithm string) (interface{}, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch algorithm {
	case "rsa-sha256", "rsa-sha384", "rsa-sha512":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "ecdsa-p256-sha256", "ecdsa-p384-sha384", "ecdsa-p521-sha512":
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// parsePublicKey parses a public key from PEM format
func (s *SignatureService) parsePublicKey(pemData, algorithm string) (interface{}, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch algorithm {
	case "rsa-sha256", "rsa-sha384", "rsa-sha512":
		if rsaKey, ok := pubKey.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
	case "ecdsa-p256-sha256", "ecdsa-p384-sha384", "ecdsa-p521-sha512":
		if ecdsaKey, ok := pubKey.(*ecdsa.PublicKey); ok {
			return ecdsaKey, nil
		}
	}

	return nil, fmt.Errorf("public key type does not match algorithm")
}

// signContent signs content using the specified algorithm
func (s *SignatureService) signContent(content []byte, privateKey interface{}, algorithm string) ([]byte, string, error) {
	var hash crypto.Hash
	var keyID string

	switch algorithm {
	case "rsa-sha256":
		hash = crypto.SHA256
		if rsaKey, ok := privateKey.(*rsa.PrivateKey); ok {
			keyID = fmt.Sprintf("rsa:%x", rsaKey.PublicKey.N)
		}
	case "rsa-sha384":
		hash = crypto.SHA384
		if rsaKey, ok := privateKey.(*rsa.PrivateKey); ok {
			keyID = fmt.Sprintf("rsa:%x", rsaKey.PublicKey.N)
		}
	case "rsa-sha512":
		hash = crypto.SHA512
		if rsaKey, ok := privateKey.(*rsa.PrivateKey); ok {
			keyID = fmt.Sprintf("rsa:%x", rsaKey.PublicKey.N)
		}
	case "ecdsa-p256-sha256":
		hash = crypto.SHA256
		if ecdsaKey, ok := privateKey.(*ecdsa.PrivateKey); ok {
			keyID = fmt.Sprintf("ecdsa-p256:%x", elliptic.Marshal(ecdsaKey.Curve, ecdsaKey.PublicKey.X, ecdsaKey.PublicKey.Y))
		}
	case "ecdsa-p384-sha384":
		hash = crypto.SHA384
		if ecdsaKey, ok := privateKey.(*ecdsa.PrivateKey); ok {
			keyID = fmt.Sprintf("ecdsa-p384:%x", elliptic.Marshal(ecdsaKey.Curve, ecdsaKey.PublicKey.X, ecdsaKey.PublicKey.Y))
		}
	case "ecdsa-p521-sha512":
		hash = crypto.SHA512
		if ecdsaKey, ok := privateKey.(*ecdsa.PrivateKey); ok {
			keyID = fmt.Sprintf("ecdsa-p521:%x", elliptic.Marshal(ecdsaKey.Curve, ecdsaKey.PublicKey.X, ecdsaKey.PublicKey.Y))
		}
	default:
		return nil, "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	hashed := hash.New().Sum(content)

	switch algorithm {
	case "rsa-sha256", "rsa-sha384", "rsa-sha512":
		if rsaKey, ok := privateKey.(*rsa.PrivateKey); ok {
			signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, hash, hashed)
			return signature, keyID, err
		}
	case "ecdsa-p256-sha256", "ecdsa-p384-sha384", "ecdsa-p521-sha512":
		if ecdsaKey, ok := privateKey.(*ecdsa.PrivateKey); ok {
			signature, err := ecdsa.SignASN1(rand.Reader, ecdsaKey, hashed)
			return signature, keyID, err
		}
	}

	return nil, "", fmt.Errorf("invalid key type for algorithm")
}

// verifyContent verifies content signature using the specified algorithm
func (s *SignatureService) verifyContent(content, signature []byte, publicKey interface{}, algorithm string) (bool, error) {
	var hash crypto.Hash

	switch algorithm {
	case "rsa-sha256":
		hash = crypto.SHA256
	case "rsa-sha384":
		hash = crypto.SHA384
	case "rsa-sha512":
		hash = crypto.SHA512
	case "ecdsa-p256-sha256":
		hash = crypto.SHA256
	case "ecdsa-p384-sha384":
		hash = crypto.SHA384
	case "ecdsa-p521-sha512":
		hash = crypto.SHA512
	default:
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	hashed := hash.New().Sum(content)

	switch algorithm {
	case "rsa-sha256", "rsa-sha384", "rsa-sha512":
		if rsaKey, ok := publicKey.(*rsa.PublicKey); ok {
			return rsa.VerifyPKCS1v15(rsaKey, hash, hashed, signature) == nil, nil
		}
	case "ecdsa-p256-sha256", "ecdsa-p384-sha384", "ecdsa-p521-sha512":
		if ecdsaKey, ok := publicKey.(*ecdsa.PublicKey); ok {
			return ecdsa.VerifyASN1(ecdsaKey, hashed, signature), nil
		}
	}

	return false, fmt.Errorf("invalid key type for algorithm")
}

// getSignatureByID retrieves a signature by ID
func (s *SignatureService) getSignatureByID(signatureID uuid.UUID) (*registry.RegistryFunctionSignature, error) {
	return s.repo.GetSignatureByID(signatureID)
}

// updateSignature updates a signature record
func (s *SignatureService) updateSignature(signatureID uuid.UUID, updates map[string]interface{}) error {
	return s.repo.UpdateSignature(signatureID, updates)
}

// GenerateKeyPair generates a new key pair for signing
func (s *SignatureService) GenerateKeyPair(algorithm string) (privateKeyPEM, publicKeyPEM string, keyID string, err error) {
	switch algorithm {
	case "rsa-sha256":
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to generate RSA key: %w", err)
		}

		privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
		privatePEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateDER,
		})

		publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to marshal public key: %w", err)
		}

		publicPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicDER,
		})

		keyID = fmt.Sprintf("rsa:%x", privateKey.PublicKey.N)

		return string(privatePEM), string(publicPEM), keyID, nil

	case "ecdsa-p256-sha256":
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to generate ECDSA key: %w", err)
		}

		privateDER, err := x509.MarshalECPrivateKey(privateKey)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to marshal private key: %w", err)
		}

		privatePEM := pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: privateDER,
		})

		publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to marshal public key: %w", err)
		}

		publicPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicDER,
		})

		keyID = fmt.Sprintf("ecdsa-p256:%x", elliptic.Marshal(privateKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y))

		return string(privatePEM), string(publicPEM), keyID, nil

	default:
		return "", "", "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}
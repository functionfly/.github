//go:build !cgo

package trustapi

import "fmt"

// PKCS11Signer is a stub type that exists so the rest of the codebase
// can compile when CGO is disabled. newPKCS11Signer always returns an
// error in this build configuration, so a value of *PKCS11Signer can
// never be produced at runtime.
type PKCS11Signer struct{}

func newPKCS11Signer(cfg SignerConfig) (Signer, error) {
	return nil, fmt.Errorf("pkcs11 signer backend requires CGO (rebuild with CGO_ENABLED=1)")
}

func (s *PKCS11Signer) Algorithm() SignatureAlgorithm                       { return "" }
func (s *PKCS11Signer) KeyID() string                                       { return "" }
func (s *PKCS11Signer) PublicKeyHex() string                                { return "" }
func (s *PKCS11Signer) Sign([]byte) (string, error)                         { return "", fmt.Errorf("pkcs11 signer disabled (no CGO)") }
func (s *PKCS11Signer) Verify([]byte, string) (bool, error)                 { return false, fmt.Errorf("pkcs11 signer disabled (no CGO)") }
func (s *PKCS11Signer) SignAttestation(*TrustAttestation) error             { return fmt.Errorf("pkcs11 signer disabled (no CGO)") }
func (s *PKCS11Signer) VerifyAttestationSignature(*TrustAttestation) (bool, error) {
	return false, fmt.Errorf("pkcs11 signer disabled (no CGO)")
}
func (s *PKCS11Signer) Close() error { return nil }
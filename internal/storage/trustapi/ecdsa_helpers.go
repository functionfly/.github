package trustapi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"
)

// rawECDSAToDER converts PKCS#11 raw (r||s) to ASN.1 DER.
func rawECDSAToDER(raw []byte) ([]byte, error) {
	if len(raw)%2 != 0 {
		return nil, fmt.Errorf("invalid raw ECDSA signature length")
	}
	half := len(raw) / 2
	r := new(big.Int).SetBytes(raw[:half])
	s := new(big.Int).SetBytes(raw[half:])
	return asn1.Marshal(struct{ R, S *big.Int }{R: r, S: s})
}

// verifyECDSADER verifies a DER-encoded ECDSA signature using Go's crypto/ecdsa.
func verifyECDSADER(data, derSig []byte, x, y *big.Int) (bool, error) {
	var sig struct{ R, S *big.Int }
	if _, err := asn1.Unmarshal(derSig, &sig); err != nil {
		return false, fmt.Errorf("unmarshal DER signature: %w", err)
	}
	hash := sha256.Sum256(data)
	pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	return ecdsa.Verify(pub, hash[:], sig.R, sig.S), nil
}
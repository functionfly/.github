package trustapi

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sync"

	"github.com/miekg/pkcs11"
)

// PKCS11Signer implements Signer using a PKCS#11-compatible HSM
// (YubiHSM, SoftHSM, AWS CloudHSM, Thales Luna, etc.).
type PKCS11Signer struct {
	ctx      *pkcs11.Ctx
	session  pkcs11.SessionHandle
	keyLabel string
	keyID    string
	slotID   uint
	alg      SignatureAlgorithm
	mu       sync.Mutex

	// Cached public key for Verify (avoids repeated HSM round-trips)
	edPub  ed25519.PublicKey
	ecPubX *big.Int
	ecPubY *big.Int
}

// newPKCS11Signer initializes a PKCS#11 session, locates or generates
// a signing key, and caches the public key.
func newPKCS11Signer(cfg SignerConfig) (*PKCS11Signer, error) {
	if cfg.PKCS11LibPath == "" {
		return nil, fmt.Errorf("PKCS11_LIBRARY_PATH is required for pkcs11 backend")
	}

	ctx := pkcs11.New(cfg.PKCS11LibPath)
	if ctx == nil {
		return nil, fmt.Errorf("failed to load PKCS#11 library: %s", cfg.PKCS11LibPath)
	}

	if err := ctx.Initialize(); err != nil {
		ctx.Destroy()
		return nil, fmt.Errorf("PKCS#11 initialize: %w", err)
	}

	slotID := uint(cfg.PKCS11SlotID)
	if cfg.PKCS11SlotID == 0 {
		slots, err := ctx.GetSlotList(true)
		if err != nil || len(slots) == 0 {
			ctx.Destroy()
			return nil, fmt.Errorf("no PKCS#11 slots available: %v", err)
		}
		slotID = slots[0]
	}

	session, err := ctx.OpenSession(slotID, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		ctx.Destroy()
		return nil, fmt.Errorf("open PKCS#11 session: %w", err)
	}

	if cfg.PKCS11Pin != "" {
		if err := ctx.Login(session, pkcs11.CKU_USER, cfg.PKCS11Pin); err != nil {
			ctx.CloseSession(session)
			ctx.Destroy()
			return nil, fmt.Errorf("PKCS#11 login: %w", err)
		}
	}

	s := &PKCS11Signer{
		ctx:      ctx,
		session:  session,
		keyLabel: cfg.PKCS11Label,
		slotID:   slotID,
		alg:      AlgEd25519,
	}

	if err := s.loadOrCreateKey(); err != nil {
		s.Close()
		return nil, err
	}

	return s, nil
}

// loadOrCreateKey finds an existing key or generates a new one.
func (s *PKCS11Signer) loadOrCreateKey() error {
	if err := s.loadPublicKey(); err == nil {
		return nil
	}
	// Key not found — generate
	if err := s.generateKeyPair(); err != nil {
		return fmt.Errorf("generate key pair: %w", err)
	}
	return s.loadPublicKey()
}

// loadPublicKey reads the public key from the HSM and caches it.
func (s *PKCS11Signer) loadPublicKey() error {
	template := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PUBLIC_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, []byte(s.keyLabel)),
	}

	if err := s.ctx.FindObjectsInit(s.session, template); err != nil {
		return err
	}
	objs, _, err := s.ctx.FindObjects(s.session, 1)
	if err != nil {
		s.ctx.FindObjectsFinal(s.session)
		return err
	}
	if err := s.ctx.FindObjectsFinal(s.session); err != nil {
		return err
	}
	if len(objs) == 0 {
		return fmt.Errorf("no public key with label %q", s.keyLabel)
	}

	attrs, err := s.ctx.GetAttributeValue(s.session, objs[0], []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, nil),
		pkcs11.NewAttribute(pkcs11.CKA_EC_POINT, nil),
	})
	if err != nil {
		return fmt.Errorf("get public key: %w", err)
	}

	keyType := bigEndianUint(attrs[0].Value)

	switch keyType {
	case pkcs11.CKK_EC:
		s.alg = AlgECDSA
		x, y, err := parseECPoint(attrs[1].Value)
		if err != nil {
			return fmt.Errorf("parse EC point: %w", err)
		}
		s.ecPubX = x
		s.ecPubY = y
	default:
		s.alg = AlgEd25519
		raw := attrs[1].Value
		if len(raw) < ed25519.PublicKeySize {
			return fmt.Errorf("Ed25519 public key too short: %d bytes", len(raw))
		}
		// Some HSMs wrap in DER OCTET STRING — strip if needed
		if len(raw) > ed25519.PublicKeySize && raw[0] == 0x04 {
			raw = raw[len(raw)-ed25519.PublicKeySize:]
		}
		s.edPub = ed25519.PublicKey(raw[:ed25519.PublicKeySize])
	}

	h := sha256.Sum256(attrs[1].Value)
	s.keyID = keyIDPrefix + hex.EncodeToString(h[:8])
	return nil
}

// generateKeyPair creates a new key pair in the HSM.
func (s *PKCS11Signer) generateKeyPair() error {
	oidP256 := []byte{0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x03, 0x01, 0x07}

	pubTmpl := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PUBLIC_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, []byte(s.keyLabel)),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_VERIFY, true),
	}
	privTmpl := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, []byte(s.keyLabel)),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
		pkcs11.NewAttribute(pkcs11.CKA_SENSITIVE, true),
		pkcs11.NewAttribute(pkcs11.CKA_EXTRACTABLE, false),
	}

	// Try Ed25519 (CKM_EDDSA = 0x00000086 in PKCS#11 3.0)
	edMech := []*pkcs11.Mechanism{pkcs11.NewMechanism(0x00000086, nil)}
	_, _, err := s.ctx.GenerateKeyPair(s.session, edMech, pubTmpl, privTmpl)
	if err == nil {
		s.alg = AlgEd25519
		fmt.Fprintln(os.Stderr, "pkcs11: generated Ed25519 key pair")
		return nil
	}

	// Fall back to ECDSA-P256
	pubTmpl = append(pubTmpl,
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_EC),
		pkcs11.NewAttribute(pkcs11.CKA_EC_PARAMS, oidP256),
	)
	privTmpl = append(privTmpl,
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_EC),
	)
	ecMech := []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_EC_KEY_PAIR_GEN, nil)}
	_, _, err = s.ctx.GenerateKeyPair(s.session, ecMech, pubTmpl, privTmpl)
	if err != nil {
		return fmt.Errorf("generate key pair (Ed25519 + ECDSA both failed): %w", err)
	}
	s.alg = AlgECDSA
	fmt.Fprintln(os.Stderr, "pkcs11: generated ECDSA-P256 key pair")
	return nil
}

func (s *PKCS11Signer) Algorithm() SignatureAlgorithm { return s.alg }
func (s *PKCS11Signer) KeyID() string                 { return s.keyID }

func (s *PKCS11Signer) PublicKeyHex() string {
	switch s.alg {
	case AlgECDSA:
		if s.ecPubX == nil || s.ecPubY == nil {
			return ""
		}
		return hex.EncodeToString(elliptic.Marshal(elliptic.P256(), s.ecPubX, s.ecPubY))
	default:
		return hex.EncodeToString(s.edPub)
	}
}

func (s *PKCS11Signer) Sign(data []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	privKey, err := s.findPrivateKey()
	if err != nil {
		return "", fmt.Errorf("find private key: %w", err)
	}

	var mechType uint
	switch s.alg {
	case AlgECDSA:
		mechType = pkcs11.CKM_ECDSA_SHA256
	default:
		mechType = 0x00000087 // CKM_EDDSA single-part
	}

	mech := []*pkcs11.Mechanism{pkcs11.NewMechanism(mechType, nil)}
	if err := s.ctx.SignInit(s.session, mech, privKey); err != nil {
		return "", fmt.Errorf("sign init: %w", err)
	}

	rawSig, err := s.ctx.Sign(s.session, data)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	if s.alg == AlgECDSA {
		// PKCS#11 returns raw (r||s), convert to ASN.1 DER
		der, err := rawECDSAToDER(rawSig)
		if err != nil {
			return "", fmt.Errorf("marshal ECDSA sig: %w", err)
		}
		return hex.EncodeToString(der), nil
	}

	return hex.EncodeToString(rawSig), nil
}

func (s *PKCS11Signer) Verify(data []byte, sigHex string) (bool, error) {
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}

	switch s.alg {
	case AlgECDSA:
		if s.ecPubX == nil || s.ecPubY == nil {
			return false, fmt.Errorf("no ECDSA public key cached")
		}
		return verifyECDSADER(data, sig, s.ecPubX, s.ecPubY)
	default:
		if len(s.edPub) == 0 {
			return false, fmt.Errorf("no Ed25519 public key cached")
		}
		return ed25519.Verify(s.edPub, data, sig), nil
	}
}

func (s *PKCS11Signer) SignAttestation(att *TrustAttestation) error {
	att.ProofHash = att.CalculateProofHash()
	sig, err := s.Sign([]byte(att.ProofHash))
	if err != nil {
		return fmt.Errorf("sign attestation: %w", err)
	}
	att.Signature = sig
	att.PublicKeyID = s.keyID
	return nil
}

func (s *PKCS11Signer) VerifyAttestationSignature(att *TrustAttestation) (bool, error) {
	if att.Signature == "" {
		return false, nil
	}
	if !att.VerifyIntegrity() {
		return false, nil
	}
	return s.Verify([]byte(att.ProofHash), att.Signature)
}

// Close releases the PKCS#11 session and context.
func (s *PKCS11Signer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ctx == nil {
		return nil
	}
	_ = s.ctx.Logout(s.session)
	_ = s.ctx.CloseSession(s.session)
	s.ctx.Destroy()
	s.ctx = nil
	return nil
}

// findPrivateKey locates the private key handle by label.
func (s *PKCS11Signer) findPrivateKey() (pkcs11.ObjectHandle, error) {
	tmpl := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, []byte(s.keyLabel)),
	}
	if err := s.ctx.FindObjectsInit(s.session, tmpl); err != nil {
		return 0, err
	}
	objs, _, err := s.ctx.FindObjects(s.session, 1)
	if err != nil {
		s.ctx.FindObjectsFinal(s.session)
		return 0, err
	}
	if err := s.ctx.FindObjectsFinal(s.session); err != nil {
		return 0, err
	}
	if len(objs) == 0 {
		return 0, fmt.Errorf("no private key with label %q", s.keyLabel)
	}
	return objs[0], nil
}

// bigEndianUint reads a uint from a big-endian byte slice.
func bigEndianUint(b []byte) uint {
	var v uint
	for _, by := range b {
		v = v<<8 | uint(by)
	}
	return v
}

// parseECPoint extracts X,Y from a DER-encoded OCTET STRING or raw uncompressed point.
func parseECPoint(der []byte) (*big.Int, *big.Int, error) {
	raw := der
	// Strip DER OCTET STRING wrapper if present
	if len(raw) > 2 && raw[0] == 0x04 && raw[1] == byte(len(raw)-2) {
		raw = raw[2:]
	}
	if len(raw) == 65 && raw[0] == 0x04 {
		return new(big.Int).SetBytes(raw[1:33]), new(big.Int).SetBytes(raw[33:65]), nil
	}
	return nil, nil, fmt.Errorf("unsupported EC point format (%d bytes)", len(raw))
}

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

// Compile-time interface check.
var _ Signer = (*PKCS11Signer)(nil)

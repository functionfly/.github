package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Client-wrap envelope constants (2026-06-16 design).
const (
	// ClientWrapIVLength is the AES-GCM IV size (12 bytes / 96 bits).
	ClientWrapIVLength = 12
	// ClientWrapTagLength is the AES-GCM auth tag size (16 bytes).
	ClientWrapTagLength = 16
	// ClientWrapKeyLength is the AES-256 key size (32 bytes).
	ClientWrapKeyLength = 32
	// ClientWrapKeyVersion is the KDF envelope version (Argon2id).
	ClientWrapKeyVersion = 3
	// ClientWrapAADPrefix is the prefix that binds ciphertext to a
	// (tenant, target) row. Prevents a ciphertext minted for target A
	// from being replayed as a ciphertext for target B.
	ClientWrapAADPrefix = "client-wrap:"
)

// ClientWrapAAD builds the Additional Authenticated Data string used
// for both encrypt and decrypt: "client-wrap:‹tenantID›:‹targetID›".
// The values are bound to the caller's identity so a ciphertext
// cannot be moved across rows.
func ClientWrapAAD(tenantID, targetID uuid.UUID) []byte {
	return []byte(fmt.Sprintf("%s%s:%s", ClientWrapAADPrefix, tenantID.String(), targetID.String()))
}

// ValidateClientWrapEnvelope checks the structural shape of a
// client-wrap envelope. It does NOT verify the AEAD tag (the server
// does not hold the key); it only catches obviously malformed input.
func ValidateClientWrapEnvelope(ciphertext, iv, tag []byte) error {
	if len(iv) != ClientWrapIVLength {
		return fmt.Errorf("client-wrap: iv must be %d bytes, got %d", ClientWrapIVLength, len(iv))
	}
	if len(tag) != ClientWrapTagLength {
		return fmt.Errorf("client-wrap: auth tag must be %d bytes, got %d", ClientWrapTagLength, len(tag))
	}
	if len(ciphertext) == 0 {
		return errors.New("client-wrap: ciphertext is empty")
	}
	return nil
}

// GenerateClientDEK returns 32 random bytes to be used as the
// per-tenant DEK. The DEK is generated client-side; this helper is
// provided so the agent binary and tests have a single source of
// randomness that matches the on-the-wire format.
func GenerateClientDEK() ([]byte, error) {
	dek := make([]byte, ClientWrapKeyLength)
	if _, err := rand.Read(dek); err != nil {
		return nil, fmt.Errorf("generate client DEK: %w", err)
	}
	return dek, nil
}

// ClientWrapEncrypt encrypts plaintext under a DEK with the standard
// AAD. Exposed for tests + the agent binary; the server's own code
// path never produces a client-wrap ciphertext.
func ClientWrapEncrypt(plaintext []byte, dek []byte, tenantID, targetID uuid.UUID) (ciphertext, iv, tag []byte, err error) {
	if len(dek) != ClientWrapKeyLength {
		return nil, nil, nil, fmt.Errorf("client-wrap: DEK must be %d bytes, got %d", ClientWrapKeyLength, len(dek))
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("client-wrap: cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("client-wrap: gcm: %w", err)
	}
	iv = make([]byte, ClientWrapIVLength)
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, fmt.Errorf("client-wrap: iv: %w", err)
	}
	sealed := gcm.Seal(nil, iv, plaintext, ClientWrapAAD(tenantID, targetID))
	tagLen := ClientWrapTagLength
	tag = sealed[len(sealed)-tagLen:]
	ciphertext = sealed[:len(sealed)-tagLen]
	return ciphertext, iv, tag, nil
}

// ClientWrapDecrypt decrypts a client-wrap envelope. Exposed for the
// agent binary + tests; the server must never call this with a key it
// does not hold.
func ClientWrapDecrypt(ciphertext, iv, tag []byte, dek []byte, tenantID, targetID uuid.UUID) ([]byte, error) {
	if len(dek) != ClientWrapKeyLength {
		return nil, fmt.Errorf("client-wrap: DEK must be %d bytes, got %d", ClientWrapKeyLength, len(dek))
	}
	if err := ValidateClientWrapEnvelope(ciphertext, iv, tag); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("client-wrap: cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("client-wrap: gcm: %w", err)
	}
	combined := make([]byte, 0, len(ciphertext)+len(tag))
	combined = append(combined, ciphertext...)
	combined = append(combined, tag...)
	return gcm.Open(nil, iv, combined, ClientWrapAAD(tenantID, targetID))
}

// Zeroize overwrites the contents of a byte slice with zeros. Callers
// should use this to clear any plaintext password material after
// use. Best-effort: the Go GC may have already copied the bytes
// elsewhere; this is a defense-in-depth measure.
func Zeroize(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// ZeroizeString is a convenience wrapper for *string. Strings in Go
// are immutable, so this only works when callers keep a *string or
// pass a []byte and convert.
func ZeroizeString(s *string) {
	if s == nil {
		return
	}
	v := []byte(*s)
	Zeroize(v)
	*s = ""
}

package vault

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/crypto"
	"github.com/google/uuid"
)

// decryptAdminPassword decrypts the target's encrypted admin password
// using the platform's server-side envelope. The stored layout matches
// what internal/crypto.ServerEncrypt produces: ciphertext bytes with
// the 16-byte GCM auth tag appended, and a separate 12-byte nonce.
func decryptAdminPassword(ctx context.Context, t *DynamicSecretTarget) (string, error) {
	if t == nil {
		return "", fmt.Errorf("nil target")
	}
	ct := t.EncryptedAdminPassword
	iv := t.PasswordNonce
	tagLen := 16
	if len(ct) < tagLen {
		return "", fmt.Errorf("encrypted admin password too short")
	}
	tag := ct[len(ct)-tagLen:]
	body := ct[:len(ct)-tagLen]
	plain, err := crypto.ServerDecrypt(body, iv, []byte{}, tag, t.TenantID)
	if err != nil {
		return "", fmt.Errorf("decrypt admin password: %w", err)
	}
	return string(plain), nil
}

// encryptAdminPasswordForTarget wraps crypto.ServerEncrypt and re-lays
// the bytes in the layout decryptAdminPassword expects.
func encryptAdminPasswordForTarget(plaintext []byte, tenantID uuid.UUID) (ct, nonce []byte, keyVersion int, err error) {
	body, iv, _, tag, err := crypto.ServerEncrypt(plaintext, tenantID)
	if err != nil {
		return nil, nil, 0, err
	}
	combined := make([]byte, 0, len(body)+len(tag))
	combined = append(combined, body...)
	combined = append(combined, tag...)
	return combined, iv, 1, nil
}

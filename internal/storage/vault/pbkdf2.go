package vault

import (
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

// PBKDF2Deriver is the legacy deriver used for v1 secrets. It is retained
// so existing ciphertexts remain readable; new secrets should use Argon2id.
type PBKDF2Deriver struct{}

func (PBKDF2Deriver) DeriveKey(passphrase string, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		return nil, errEmptySalt
	}
	return pbkdf2.Key([]byte(passphrase), salt, PBKDF2Iterations, PBKDF2KeyBytes, sha256.New), nil
}

func (PBKDF2Deriver) DefaultParams() KDFParams {
	return KDFParams{
		Method:     KDFMethodPBKDF2SHA256,
		Iterations: PBKDF2Iterations,
		KeyBytes:   PBKDF2KeyBytes,
		SaltBytes:  PBKDF2SaltBytes,
	}
}

func (PBKDF2Deriver) Method() KDFMethod { return KDFMethodPBKDF2SHA256 }

var errEmptySalt = &kdfError{msg: "pbkdf2: salt must not be empty"}

type kdfError struct{ msg string }

func (e *kdfError) Error() string { return e.msg }

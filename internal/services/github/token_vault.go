package github

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type TokenVault struct {
	key []byte
}

func NewTokenVault(encryptionKey string) (*TokenVault, error) {
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes, got %d", len(encryptionKey))
	}
	return &TokenVault{key: []byte(encryptionKey)}, nil
}

func (v *TokenVault) Encrypt(plaintext string) (ciphertext, iv, tag string, err error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", "", "", fmt.Errorf("create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", "", fmt.Errorf("generate nonce: %w", err)
	}

	sealed := aesGCM.Seal(nil, nonce, []byte(plaintext), nil)
	tagSize := aesGCM.Overhead()

	encryptedData := sealed[:len(sealed)-tagSize]
	authTag := sealed[len(sealed)-tagSize:]

	return base64.StdEncoding.EncodeToString(encryptedData),
		base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(authTag),
		nil
}

func (v *TokenVault) Decrypt(ciphertextB64, ivB64, tagB64 string) (string, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return "", fmt.Errorf("decode iv: %w", err)
	}

	authTag, err := base64.StdEncoding.DecodeString(tagB64)
	if err != nil {
		return "", fmt.Errorf("decode tag: %w", err)
	}

	block, err := aes.NewCipher(v.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	if len(nonce) != aesGCM.NonceSize() {
		return "", fmt.Errorf("invalid nonce size: got %d, want %d", len(nonce), aesGCM.NonceSize())
	}

	sealed := append(encryptedData, authTag...)

	plaintext, err := aesGCM.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

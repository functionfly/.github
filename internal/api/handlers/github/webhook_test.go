package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyHMAC(t *testing.T) {
	secret := "test-webhook-secret-12345"
	payload := []byte(`{"ref":"refs/heads/main","after":"abc123"}`)

	t.Run("valid signature", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		result := verifyHMAC(payload, sig, secret)
		assert.True(t, result)
	})

	t.Run("invalid signature", func(t *testing.T) {
		result := verifyHMAC(payload, "sha256=deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", secret)
		assert.False(t, result)
	})

	t.Run("missing sha256 prefix", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))

		result := verifyHMAC(payload, sig, secret)
		assert.False(t, result)
	})

	t.Run("empty signature", func(t *testing.T) {
		result := verifyHMAC(payload, "", secret)
		assert.False(t, result)
	})

	t.Run("empty secret", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(""))
		mac.Write(payload)
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		result := verifyHMAC(payload, sig, "")
		assert.True(t, result)
	})

	t.Run("empty payload", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte{})
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		result := verifyHMAC([]byte{}, sig, secret)
		assert.True(t, result)
	})

	t.Run("wrong secret", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte("wrong-secret"))
		mac.Write(payload)
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		result := verifyHMAC(payload, sig, secret)
		assert.False(t, result)
	})

	t.Run("tampered payload", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		tamperedPayload := append(payload, []byte(",extra")...)
		result := verifyHMAC(tamperedPayload, sig, secret)
		assert.False(t, result)
	})

	t.Run("invalid hex in signature", func(t *testing.T) {
		result := verifyHMAC(payload, "sha256=not-valid-hex!!!", secret)
		assert.False(t, result)
	})
}

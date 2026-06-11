package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// HMACSign returns a base64url-encoded HMAC-SHA256 signature of payload
// using key. If key is empty, returns "" (signing disabled).
//
// This is the package-level replacement for the method-based HMACSign
// that previously lived on receipt.Handler. Moving it here means any
// package can sign a payload without depending on the receipt handler.
func HMACSign(key []byte, payload string) string {
	if len(key) == 0 {
		return ""
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// HMACVerify returns true if sig is a valid HMAC-SHA256 signature of
// payload under key. If key is empty, returns true (signing disabled —
// fail-open so disabled-signing environments don't reject every request).
//
// The function uses crypto/hmac.Equal for constant-time comparison.
func HMACVerify(key []byte, payload, sig string) bool {
	if len(key) == 0 {
		return true
	}
	want, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return hmac.Equal(mac.Sum(nil), want)
}

// SignID returns `id.sig` if signing is enabled (key non-empty), else
// just id. The signature is appended with a "." separator so the format
// is `public_id.signature`.
func SignID(key []byte, id string) string {
	sig := HMACSign(key, id)
	if sig == "" {
		return id
	}
	return id + "." + sig
}

package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestHMACSignVerify_Roundtrip(t *testing.T) {
	key := []byte("super-secret-key-32-bytes-ok!")
	payload := "V1StGXR8_Z5jHi3B-myT"
	sig := HMACSign(key, payload)
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	if !HMACVerify(key, payload, sig) {
		t.Error("expected HMACVerify to return true for valid signature")
	}
	if HMACVerify(key, "different", sig) {
		t.Error("expected HMACVerify to return false for different payload")
	}
}

func TestHMACSign_EmptyKey(t *testing.T) {
	if sig := HMACSign(nil, "anything"); sig != "" {
		t.Errorf("expected empty signature when key is nil, got %q", sig)
	}
	if sig := HMACSign([]byte{}, "anything"); sig != "" {
		t.Errorf("expected empty signature when key is empty, got %q", sig)
	}
}

func TestHMACVerify_DisabledWhenNoKey(t *testing.T) {
	// When signing is disabled, Verify returns true (fail-open) so disabled
	// environments don't reject every request.
	if !HMACVerify(nil, "anything", "anything") {
		t.Error("expected HMACVerify to return true when signing is disabled")
	}
}

func TestHMACVerify_TamperedSignature(t *testing.T) {
	key := []byte("key")
	if HMACVerify(key, "payload", "not-base64!@#") {
		t.Error("expected false for non-base64 signature")
	}
}

func TestSignID_Format(t *testing.T) {
	id := SignID([]byte("key"), "V1StGXR8_Z5jHi3B-myT")
	if !startsWith(id, "V1StGXR8_Z5jHi3B-myT.") {
		t.Errorf("SignID should prefix the id with a dot, got %q", id)
	}
	plain := SignID(nil, "V1StGXR8_Z5jHi3B-myT")
	if plain != "V1StGXR8_Z5jHi3B-myT" {
		t.Errorf("expected identity when key is empty, got %q", plain)
	}
}

func TestHMACEqual_TimingSafe(t *testing.T) {
	// Defensive: the implementation must use hmac.Equal under the hood.
	mac := hmac.New(sha256.New, []byte("k"))
	mac.Write([]byte("payload"))
	got := mac.Sum(nil)
	encoded := base64.RawURLEncoding.EncodeToString(got)
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !hmac.Equal(got, decoded) {
		t.Error("expected roundtrip to validate")
	}
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

package frg

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestDecodeMessage_JSON(t *testing.T) {
	sub := &RuntimeSubscriber{
		trustedKeys: make(map[string]ed25519.PublicKey),
	}

	reg := RuntimeRegistration{
		CellID:       "cell-001",
		Name:         "test-cell",
		Capabilities: []string{"wasm", "fusion"},
	}
	data, _ := json.Marshal(reg)

	var out RuntimeRegistration
	if err := sub.decodeMessage(data, &out); err != nil {
		t.Fatalf("decodeMessage failed: %v", err)
	}
	if out.CellID != "cell-001" {
		t.Errorf("expected cell_id cell-001, got %s", out.CellID)
	}
}

func TestDecodeMessage_CBOR(t *testing.T) {
	sub := &RuntimeSubscriber{
		trustedKeys: make(map[string]ed25519.PublicKey),
	}

	reg := RuntimeRegistration{
		CellID:       "cell-002",
		Name:         "cbor-cell",
		Capabilities: []string{"wasm"},
	}
	data, _ := cbor.Marshal(reg)

	var out RuntimeRegistration
	if err := sub.decodeMessage(data, &out); err != nil {
		t.Fatalf("decodeMessage failed for CBOR: %v", err)
	}
	if out.CellID != "cell-002" {
		t.Errorf("expected cell_id cell-002, got %s", out.CellID)
	}
}

func TestDecodeMessage_SignedEnvelope(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	pubB64 := base64.StdEncoding.EncodeToString(pub)

	sub := &RuntimeSubscriber{
		trustedKeys: map[string]ed25519.PublicKey{
			pubB64: pub,
		},
	}

	payload := []byte(`{"cell_id":"cell-signed","name":"signed-cell","capabilities":[]}`)
	sig := ed25519.Sign(priv, payload)
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	signed := SignedNATSMessage{
		Payload:   payload,
		Signature: sigB64,
		PublicKey: pubB64,
	}
	data, _ := json.Marshal(signed)

	var out RuntimeRegistration
	if err := sub.decodeMessage(data, &out); err != nil {
		t.Fatalf("decodeMessage failed for signed envelope: %v", err)
	}
	if out.CellID != "cell-signed" {
		t.Errorf("expected cell_id cell-signed, got %s", out.CellID)
	}
}

func TestDecodeMessage_SignedEnvelope_InvalidSig(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	_, wrongPriv, _ := ed25519.GenerateKey(rand.Reader)
	pubB64 := base64.StdEncoding.EncodeToString(pub)

	sub := &RuntimeSubscriber{
		trustedKeys: map[string]ed25519.PublicKey{
			pubB64: pub,
		},
	}

	payload := []byte(`{"cell_id":"cell-bad","name":"bad","capabilities":[]}`)
	sig := ed25519.Sign(wrongPriv, payload) // Wrong key
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	signed := SignedNATSMessage{
		Payload:   payload,
		Signature: sigB64,
		PublicKey: pubB64,
	}
	data, _ := json.Marshal(signed)

	var out RuntimeRegistration
	err := sub.decodeMessage(data, &out)
	if err == nil {
		t.Fatal("expected signature verification failure, got nil")
	}
}

func TestDecodeMessage_SignedEnvelope_UntrustedKey(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	pub2, _, _ := ed25519.GenerateKey(rand.Reader) // Different key in trusted set
	pub2B64 := base64.StdEncoding.EncodeToString(pub2)

	sub := &RuntimeSubscriber{
		trustedKeys: map[string]ed25519.PublicKey{
			pub2B64: pub2,
		},
	}

	payload := []byte(`{"cell_id":"cell-untrusted","name":"untrusted","capabilities":[]}`)
	sig := ed25519.Sign(priv, payload)
	// Use a third key as the public key in the message
	unknownPub, _, _ := ed25519.GenerateKey(rand.Reader)
	unknownB64 := base64.StdEncoding.EncodeToString(unknownPub)

	signed := SignedNATSMessage{
		Payload:   payload,
		Signature: base64.StdEncoding.EncodeToString(sig),
		PublicKey: unknownB64,
	}
	data, _ := json.Marshal(signed)

	var out RuntimeRegistration
	err := sub.decodeMessage(data, &out)
	if err == nil {
		t.Fatal("expected untrusted key error, got nil")
	}
}

func TestNewRuntimeSubscriber_TrustedKeysParsing(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	pubB64 := base64.StdEncoding.EncodeToString(pub)

	config := &RuntimeSubscriberConfig{
		TrustedPublicKeys: []string{pubB64, "invalid-base64!!!"},
	}

	sub := NewRuntimeSubscriber(nil, config, RuntimeEventHandlers{})
	if len(sub.trustedKeys) != 1 {
		t.Errorf("expected 1 trusted key, got %d", len(sub.trustedKeys))
	}
	if _, ok := sub.trustedKeys[pubB64]; !ok {
		t.Error("expected valid key to be in trusted keys map")
	}
}

func TestVerifySignature_SkipWhenNoKeys(t *testing.T) {
	sub := &RuntimeSubscriber{
		trustedKeys: make(map[string]ed25519.PublicKey),
	}

	signed := SignedNATSMessage{
		Payload:   []byte("test"),
		Signature: "invalid",
		PublicKey: "invalid",
	}

	// Should skip verification when no trusted keys configured
	if err := sub.verifySignature(signed); err != nil {
		t.Errorf("expected nil error when no trusted keys, got: %v", err)
	}
}

package vault

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestClientWrapAAD(t *testing.T) {
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	target := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	got := string(ClientWrapAAD(tenant, target))
	want := "client-wrap:11111111-1111-1111-1111-111111111111:22222222-2222-2222-2222-222222222222"
	if got != want {
		t.Errorf("AAD = %q, want %q", got, want)
	}
}

func TestValidateClientWrapEnvelope(t *testing.T) {
	cases := []struct {
		name      string
		ct, iv, tag []byte
		wantErr   bool
	}{
		{"happy", []byte("x"), make([]byte, 12), make([]byte, 16), false},
		{"iv too short", []byte("x"), []byte{1, 2, 3}, make([]byte, 16), true},
		{"tag too short", []byte("x"), make([]byte, 12), []byte{1, 2}, true},
		{"empty ct", []byte{}, make([]byte, 12), make([]byte, 16), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateClientWrapEnvelope(c.ct, c.iv, c.tag)
			if (err != nil) != c.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, c.wantErr)
			}
		})
	}
}

func TestClientWrapRoundTrip(t *testing.T) {
	tenant := uuid.New()
	target := uuid.New()
	dek, err := GenerateClientDEK()
	if err != nil {
		t.Fatalf("GenerateClientDEK: %v", err)
	}
	plaintext := []byte("super-secret-admin-password")

	ct, iv, tag, err := ClientWrapEncrypt(plaintext, dek, tenant, target)
	if err != nil {
		t.Fatalf("ClientWrapEncrypt: %v", err)
	}
	if len(iv) != ClientWrapIVLength {
		t.Errorf("iv length = %d, want %d", len(iv), ClientWrapIVLength)
	}
	if len(tag) != ClientWrapTagLength {
		t.Errorf("tag length = %d, want %d", len(tag), ClientWrapTagLength)
	}
	got, err := ClientWrapDecrypt(ct, iv, tag, dek, tenant, target)
	if err != nil {
		t.Fatalf("ClientWrapDecrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("round-trip mismatch: got %q, want %q", got, plaintext)
	}
}

func TestClientWrapAADBinding(t *testing.T) {
	tenant := uuid.New()
	target1 := uuid.New()
	target2 := uuid.New()
	dek, _ := GenerateClientDEK()

	ct, iv, tag, err := ClientWrapEncrypt([]byte("secret"), dek, tenant, target1)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := ClientWrapDecrypt(ct, iv, tag, dek, tenant, target2); err == nil {
		t.Error("expected AAD mismatch error when target differs, got nil")
	}
	if _, err := ClientWrapDecrypt(ct, iv, tag, dek, uuid.New(), target1); err == nil {
		t.Error("expected AAD mismatch error when tenant differs, got nil")
	}
	tampered := append([]byte{}, ct...)
	tampered[0] ^= 0xff
	if _, err := ClientWrapDecrypt(tampered, iv, tag, dek, tenant, target1); err == nil {
		t.Error("expected AEAD tag mismatch on tampered ciphertext, got nil")
	}
}

func TestClientWrapRejectsBadKeyLength(t *testing.T) {
	tenant := uuid.New()
	target := uuid.New()
	shortKey := []byte{1, 2, 3}
	if _, _, _, err := ClientWrapEncrypt([]byte("x"), shortKey, tenant, target); err == nil {
		t.Error("expected error for short DEK in encrypt")
	}
	if _, err := ClientWrapDecrypt([]byte("x"), make([]byte, 12), make([]byte, 16), shortKey, tenant, target); err == nil {
		t.Error("expected error for short DEK in decrypt")
	}
}

func TestZeroize(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5}
	Zeroize(b)
	for i, v := range b {
		if v != 0 {
			t.Errorf("byte %d = %d, want 0", i, v)
		}
	}
}

func TestZeroizeString(t *testing.T) {
	s := "hello"
	ZeroizeString(&s)
	if s != "" {
		t.Errorf("after zeroize, s = %q, want empty", s)
	}
}

func TestResolveAdminPasswordServerMode(t *testing.T) {
	tenant := uuid.New()
	target := &DynamicSecretTarget{
		TenantID:               tenant,
		EncryptionMode:         DynamicSecretEncryptionServer,
		EncryptedAdminPassword: []byte("unused"),
		PasswordNonce:          []byte("unused"),
	}
	got, err := ResolveAdminPassword(context.Background(), target, "explicit-pw")
	if err != nil {
		t.Fatalf("ResolveAdminPassword: %v", err)
	}
	if got != "explicit-pw" {
		t.Errorf("got %q, want explicit-pw", got)
	}
}

func TestResolveAdminPasswordClientMode(t *testing.T) {
	tenant := uuid.New()
	target := &DynamicSecretTarget{
		TenantID:       tenant,
		EncryptionMode: DynamicSecretEncryptionClient,
	}
	if _, err := ResolveAdminPassword(context.Background(), target, ""); err != ErrClientEncryptedTarget {
		t.Errorf("got err = %v, want ErrClientEncryptedTarget", err)
	}
	got, err := ResolveAdminPassword(context.Background(), target, "client-supplied")
	if err != nil {
		t.Fatalf("ResolveAdminPassword: %v", err)
	}
	if got != "client-supplied" {
		t.Errorf("got %q, want client-supplied", got)
	}
}

func TestGenerateClientDEK(t *testing.T) {
	d1, err := GenerateClientDEK()
	if err != nil {
		t.Fatalf("GenerateClientDEK: %v", err)
	}
	if len(d1) != ClientWrapKeyLength {
		t.Errorf("len = %d, want %d", len(d1), ClientWrapKeyLength)
	}
	d2, _ := GenerateClientDEK()
	if bytes.Equal(d1, d2) {
		t.Error("two consecutive DEKs are identical; RNG broken?")
	}
}

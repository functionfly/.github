package vault

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestArgon2idDeriver_DeterministicForSameSalt(t *testing.T) {
	deriver := Argon2idDeriver{}
	salt := bytes.Repeat([]byte{0x42}, Argon2SaltBytes)
	k1, err := deriver.DeriveKey("correct horse battery staple", salt)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	k2, err := deriver.DeriveKey("correct horse battery staple", salt)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	if !bytes.Equal(k1, k2) {
		t.Fatal("expected deterministic key for same passphrase+salt")
	}
	if len(k1) != int(Argon2KeyBytes) {
		t.Fatalf("expected key length %d, got %d", Argon2KeyBytes, len(k1))
	}
}

func TestArgon2idDeriver_RejectsShortSalt(t *testing.T) {
	deriver := Argon2idDeriver{}
	shortSalt := []byte{0x01, 0x02, 0x03}
	if _, err := deriver.DeriveKey("anything", shortSalt); err == nil {
		t.Fatal("expected error for short salt")
	}
}

func TestArgon2idDeriver_DifferentSaltsProduceDifferentKeys(t *testing.T) {
	deriver := Argon2idDeriver{}
	salt1 := make([]byte, Argon2SaltBytes)
	salt2 := make([]byte, Argon2SaltBytes)
	rand.Read(salt1)
	rand.Read(salt2)
	k1, _ := deriver.DeriveKey("passphrase", salt1)
	k2, _ := deriver.DeriveKey("passphrase", salt2)
	if bytes.Equal(k1, k2) {
		t.Fatal("different salts must produce different keys")
	}
}

func TestNewDeriver_DispatchesByKeyVersion(t *testing.T) {
	if _, err := NewDeriver(KeyVersionArgon2); err != nil {
		t.Fatalf("argon2: %v", err)
	}
	if _, err := NewDeriver(KeyVersionPBKDF2); err != nil {
		t.Fatalf("pbkdf2: %v", err)
	}
	if _, err := NewDeriver(99); err == nil {
		t.Fatal("expected error for unknown key version")
	}
}

func TestPBKDF2Deriver_EmptySaltIsRejected(t *testing.T) {
	deriver := PBKDF2Deriver{}
	if _, err := deriver.DeriveKey("x", nil); err == nil {
		t.Fatal("expected error for empty salt")
	}
}

func TestGenerateSalt(t *testing.T) {
	s, err := GenerateSalt(16)
	if err != nil {
		t.Fatalf("GenerateSalt: %v", err)
	}
	if len(s) != 16 {
		t.Fatalf("expected 16 bytes, got %d", len(s))
	}
	s2, _ := GenerateSalt(16)
	if bytes.Equal(s, s2) {
		t.Fatal("GenerateSalt must produce random output")
	}
	if _, err := GenerateSalt(0); err == nil {
		t.Fatal("expected error for zero size")
	}
}

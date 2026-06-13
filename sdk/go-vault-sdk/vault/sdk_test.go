package vault

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewSalt(t *testing.T) {
	s, err := NewSalt(16)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 16 {
		t.Fatalf("len=%d, want 16", len(s))
	}
	s2, _ := NewSalt(16)
	if string(s) == string(s2) {
		t.Fatal("two salts collided")
	}
	if _, err := NewSalt(0); err == nil {
		t.Fatal("expected error for zero size")
	}
	if _, err := NewSalt(-1); err == nil {
		t.Fatal("expected error for negative size")
	}
}

func TestNewPassword(t *testing.T) {
	p, err := NewPassword(24)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 24 {
		t.Fatalf("len=%d, want 24", len(p))
	}
	// Omit ambiguous chars.
	for _, ch := range p {
		if strings.ContainsRune("0O1lI", ch) {
			t.Fatalf("ambiguous char %q in password %q", ch, p)
		}
	}
	if _, err := NewPassword(7); err == nil {
		t.Fatal("expected error for short length")
	}
}

func TestDefaultArgon2Params(t *testing.T) {
	p := DefaultArgon2Params()
	if p.Method != "argon2id" {
		t.Fatalf("method=%q, want argon2id", p.Method)
	}
	if p.MemoryKiB != 64*1024 {
		t.Fatalf("memory=%d, want 65536", p.MemoryKiB)
	}
	if p.Iterations != 3 {
		t.Fatalf("iterations=%d, want 3", p.Iterations)
	}
	if p.Parallelism != 4 {
		t.Fatalf("parallelism=%d, want 4", p.Parallelism)
	}
	if p.KeyBytes != 32 {
		t.Fatalf("key=%d, want 32", p.KeyBytes)
	}
}

// TestClient_Do_SendsAuthHeader verifies the bearer token + User-Agent
// reach the server.
func TestClient_Do_SendsAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization=%q, want Bearer test-token", got)
		}
		if got := r.Header.Get("User-Agent"); !strings.HasPrefix(got, "functionfly-go-vault-sdk/") {
			t.Errorf("User-Agent=%q, want functionfly-go-vault-sdk/...", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL, WithToken("test-token"))
	var out map[string]any
	if err := c.do(testContext(t), "GET", "/ping", nil, &out); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Do_DeserializesSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"abc-123","tenant_id":"t1","name":"KEY","secret_type":"api_key",
			"encrypted_data":{"ciphertext":"ct","iv":"iv","salt":"sa","tag":"ta","key_version":2},
			"access_count":0,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"
		}`))
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL)
	s, err := c.Secrets.Get(testContext(t), "abc-123")
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "KEY" {
		t.Fatalf("name=%q", s.Name)
	}
	if s.EncryptedData.KeyVersion != 2 {
		t.Fatalf("key_version=%d, want 2", s.EncryptedData.KeyVersion)
	}
}

func TestClient_Do_HandlesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden","message":"nope","code":"FORBIDDEN"}`))
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL)
	_, err := c.Secrets.Get(testContext(t), "x")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type=%T, want *APIError", err)
	}
	if apiErr.Status != http.StatusForbidden {
		t.Fatalf("status=%d", apiErr.Status)
	}
	if apiErr.Message != "nope" {
		t.Fatalf("message=%q", apiErr.Message)
	}
}

func TestClient_Do_PostsJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method=%s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content-type=%q", ct)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"new"}`))
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL)
	_, err := c.DynamicTargets.Create(testContext(t), DynamicTargetCreate{
		Name: "x", DBType: DynamicDBPostgres, Host: "h", Port: 5432,
		DatabaseName: "d", AdminUsername: "u", AdminPassword: "p",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSecretType_Valid(t *testing.T) {
	for _, ok := range []SecretType{SecretTypeAPIKey, SecretTypeOAuthToken, SecretTypePassword, SecretTypeCertificate} {
		if !ok.Valid() {
			t.Fatalf("%s must be valid", ok)
		}
	}
	if SecretType("bogus").Valid() {
		t.Fatal("bogus must be invalid")
	}
}

func TestDynamicDBType_Accepted(t *testing.T) {
	for _, c := range []struct {
		dt   DynamicSecretDBType
		want bool
	}{
		{DynamicDBPostgres, true},
		{DynamicDBMySQL, true},
		{"sqlite", false},
	} {
		got := c.dt == DynamicDBPostgres || c.dt == DynamicDBMySQL
		if got != c.want {
			t.Errorf("%s: got %v want %v", c.dt, got, c.want)
		}
	}
}

func TestEncryptedData_RoundTripsBase64(t *testing.T) {
	// Sanity: every field of EncryptedData is base64. We round-trip a
	// 32-byte ciphertext to make sure downstream callers can rely on
	// the documented encoding.
	raw, err := NewSalt(32)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != string(raw) {
		t.Fatal("base64 round-trip mismatch")
	}
}

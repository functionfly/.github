package vault

import (
	"strings"
	"testing"
)

func TestGenerateUsername_PostgresHasCorrectPrefix(t *testing.T) {
	u := generateUsername(DynamicSecretDBPostgres)
	if !strings.HasPrefix(u, "vault_p_") {
		t.Fatalf("postgres username must start with vault_p_, got %q", u)
	}
}

func TestGenerateUsername_MySQLHasCorrectPrefix(t *testing.T) {
	u := generateUsername(DynamicSecretDBMySQL)
	if !strings.HasPrefix(u, "vault_m_") {
		t.Fatalf("mysql username must start with vault_m_, got %q", u)
	}
}

func TestGenerateUsername_Unique(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		u := generateUsername(DynamicSecretDBPostgres)
		if _, ok := seen[u]; ok {
			t.Fatalf("collision: %q", u)
		}
		seen[u] = struct{}{}
	}
}

func TestGeneratePassword_LengthAndAlphabet(t *testing.T) {
	p, err := generatePassword(24)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 24 {
		t.Fatalf("len=%d, want 24", len(p))
	}
	// The alphabet intentionally omits 0/O/1/l/I to avoid confusion.
	for _, ch := range p {
		if strings.ContainsRune("0O1lI", ch) {
			t.Fatalf("ambiguous char %q in password %q", ch, p)
		}
	}
}

func TestGeneratePassword_Unique(t *testing.T) {
	a, _ := generatePassword(24)
	b, _ := generatePassword(24)
	if a == b {
		t.Fatal("two generated passwords collided")
	}
}

func TestGenerateLeaseID_Shape(t *testing.T) {
	id := generateLeaseID()
	if !strings.HasPrefix(id, "lease_") {
		t.Fatalf("lease id must start with lease_, got %q", id)
	}
	if len(id) < 10 {
		t.Fatalf("lease id too short: %q", id)
	}
}

func TestQuoteIdent(t *testing.T) {
	if got := quoteIdent("simple"); got != `"simple"` {
		t.Fatalf("simple: %q", got)
	}
	if got := quoteIdent(`odd"name`); got != `"odd""name"` {
		t.Fatalf("escape: %q", got)
	}
}

func TestQuoteLiteral(t *testing.T) {
	if got := quoteLiteral("abc"); got != `'abc'` {
		t.Fatalf("simple: %q", got)
	}
	if got := quoteLiteral(`o'brien`); got != `'o''brien'` {
		t.Fatalf("escape: %q", got)
	}
}

func TestDynamicSecretDBType_Valid(t *testing.T) {
	if !DynamicSecretDBPostgres.Valid() {
		t.Fatal("postgres must be valid")
	}
	if !DynamicSecretDBMySQL.Valid() {
		t.Fatal("mysql must be valid")
	}
	if DynamicSecretDBType("sqlite").Valid() {
		t.Fatal("sqlite must be invalid")
	}
}

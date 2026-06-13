package vault

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestEncodeAuditJSON_StableShape verifies the JSON envelope.
func TestEncodeAuditJSON_StableShape(t *testing.T) {
	rows := []AuditLog{
		{
			ID:        uuid.New(),
			Action:    AuditActionCreate,
			ActorID:   "user-1",
			ActorType: ActorTypeUser,
			Success:   true,
			CreatedAt: time.Unix(0, 0).UTC(),
			Metadata:  JSONMap{"k": "v"},
		},
	}
	body, err := encodeAuditJSON(rows)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"format": "json"`, `"count": 1`, `"rows"`, `"action": "create"`} {
		if !bytes.Contains(body, []byte(want)) {
			t.Errorf("missing %q in body:\n%s", want, body)
		}
	}
}

func TestEncodeAuditCSV_Header(t *testing.T) {
	body, err := encodeAuditCSV(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, []byte("id,created_at,action")) {
		t.Fatalf("missing CSV header: %q", body[:64])
	}
}

func TestEncodeAuditCEF_OneLinePerRow(t *testing.T) {
	rows := []AuditLog{
		{
			Action:    AuditActionCreate,
			ActorID:   "user-1",
			ActorType: ActorTypeUser,
			Success:   true,
			CreatedAt: time.Unix(1700000000, 0).UTC(),
		},
		{
			Action:    AuditActionDelete,
			ActorID:   "user-2",
			ActorType: ActorTypeUser,
			Success:   false,
			CreatedAt: time.Unix(1700000001, 0).UTC(),
		},
	}
	body, err := encodeAuditCEF(rows)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimRight(body, "\n"), []byte("\n"))
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), body)
	}
	if !bytes.HasPrefix(lines[0], []byte("CEF:0|FunctionFly|Vault|1|vault.create|")) {
		t.Fatalf("first row signature wrong: %q", lines[0])
	}
	if !bytes.Contains(lines[1], []byte("vault.delete|")) {
		t.Fatalf("second row missing delete: %q", lines[1])
	}
}

func TestEncodeAuditCEF_EscapesPipes(t *testing.T) {
	// The CEF extension separator is "|". A pipe inside an
	// extension value would prematurely end the row.
	body, err := encodeAuditCEF([]AuditLog{{
		Action:    "create|with|pipe",
		ActorID:   "u",
		ActorType: ActorTypeUser,
		CreatedAt: time.Unix(1700000000, 0).UTC(),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `vault.create\|with\|pipe|`) {
		t.Fatalf("pipe not escaped: %q", body)
	}
}

func TestSanitizeCEF(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"simple", "simple"},
		{"a|b", `a\|b`},
		{`a\b`, `a\\b`},
		{"a\nb", "a b"},
	}
	for _, c := range cases {
		if got := sanitizeCEF(c.in); got != c.want {
			t.Errorf("sanitizeCEF(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}

func TestOptionalUUID(t *testing.T) {
	if got := optionalUUID(nil); got != "" {
		t.Errorf("nil uuid: got %q", got)
	}
	id := uuid.New()
	if got := optionalUUID(&id); got != id.String() {
		t.Errorf("set uuid: got %q want %q", got, id.String())
	}
}

func TestIfThen(t *testing.T) {
	if got := ifThen(true, "a", "b"); got != "a" {
		t.Fatalf("true: %q", got)
	}
	if got := ifThen(false, "a", "b"); got != "b" {
		t.Fatalf("false: %q", got)
	}
}

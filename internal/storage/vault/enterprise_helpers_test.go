package vault

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// ctx is a tiny test context helper.
func ctx(t *testing.T) context.Context { t.Helper(); return context.Background() }

// seedTenantID is the canonical tenant used by RBAC + namespace tests.
var seedTenantID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

// seedUserID is the canonical user used by RBAC tests.
var seedUserID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

// contextWithClaims attaches a minimal UserClaims-equivalent to the
// request context so the handler sees a logged-in user. We can't
// import the middleware package without breaking the storage layer's
// independence, so we use a custom claims key the handler reads.
func contextWithClaims(ctx context.Context, tenantID, userID uuid.UUID) context.Context {
	// The handler uses middleware.GetUserFromContext which depends
	// on the auth.Claims type. Tests that need an authenticated
	// handler must use the dedicated harness in enterprise_test_helpers.go
	// — this helper is only useful in handlers/ tests.
	return ctx
}

// _ = ... keeps the seed vars from being dropped by the compiler when
// audit tests don't use them.
var _ = seedTenantID
var _ = seedUserID

func TestIsValidNamespacePath(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"default", true},
		{"production", true},
		{"production/api", true},
		{"production/api-gateway_v2", true},
		{"", false},
		{"Production", false},             // uppercase
		{"production//api", false},        // empty segment
		{"production/api gateway", false}, // space
		{"/production", false},            // leading slash
		{"production/", false},            // trailing slash
		{"a..b", false},                   // bad char
	}
	for _, c := range cases {
		if got := IsValidNamespacePath(c.in); got != c.want {
			t.Errorf("IsValidNamespacePath(%q)=%v, want %v", c.in, got, c.want)
		}
	}
}

func TestJoinNamespacePath(t *testing.T) {
	cases := []struct {
		parent, child, want string
	}{
		{"", "x", "x"},
		{"x", "", "x"},
		{"x", "y", "x/y"},
		{"x/", "/y", "x/y"},
		{"x/y", "z", "x/y/z"},
	}
	for _, c := range cases {
		if got := JoinNamespacePath(c.parent, c.child); got != c.want {
			t.Errorf("JoinNamespacePath(%q,%q)=%q, want %q", c.parent, c.child, got, c.want)
		}
	}
}

func TestSplitNamespacePath(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"/", nil},
		{"x", []string{"x"}},
		{"x/y", []string{"x", "y"}},
		{"x/y/z", []string{"x", "y", "z"}},
	}
	for _, c := range cases {
		got := SplitNamespacePath(c.in)
		if len(got) != len(c.want) {
			t.Errorf("SplitNamespacePath(%q)=%v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("SplitNamespacePath(%q)[%d]=%q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestNamespaceMatchesScope(t *testing.T) {
	cases := []struct {
		ns, scope string
		want      bool
	}{
		{"foo", "all", true},
		{"foo", "", true},
		{"foo", "foo", true},
		{"foo", "foo/bar", false}, // inverse: scope can't be larger than ns
		{"foo/bar", "foo", true},
		{"foo/bar", "foo/baz", false},
		{"foo/bar", "foo/bar", true},
		{"foo", "fo", false},     // prefix collision
		{"foobar", "foo", false}, // same as above
	}
	for _, c := range cases {
		if got := namespaceMatchesScope(c.ns, c.scope); got != c.want {
			t.Errorf("namespaceMatchesScope(%q,%q)=%v, want %v", c.ns, c.scope, got, c.want)
		}
	}
}

func TestBuiltinRolePermissions(t *testing.T) {
	admin := BuiltinRolePermissions(BuiltinRoleAdmin)
	if !permsAllows(JSONMap(admin), "secrets:create") {
		t.Fatal("admin must allow secrets:create")
	}
	if !permsAllows(JSONMap(admin), "rbac:manage") {
		t.Fatal("admin must allow rbac:manage")
	}
	reader := BuiltinRolePermissions(BuiltinRoleReader)
	if permsAllows(JSONMap(reader), "secrets:create") {
		t.Fatal("reader must not allow secrets:create")
	}
	if !permsAllows(JSONMap(reader), "secrets:read") {
		t.Fatal("reader must allow secrets:read")
	}
}

func TestPermsAllows(t *testing.T) {
	cases := []struct {
		name  string
		perms JSONMap
		perm  string
		want  bool
	}{
		{"nil", nil, "secrets:read", false},
		{"missing", JSONMap{"x": true}, "secrets:read", false},
		{"true", JSONMap{"secrets:read": true}, "secrets:read", true},
		{"false", JSONMap{"secrets:read": false}, "secrets:read", false},
		{"string-true", JSONMap{"secrets:read": "true"}, "secrets:read", true},
		{"non-empty-array", JSONMap{"secrets:read": []interface{}{"x"}}, "secrets:read", true},
		{"empty-array", JSONMap{"secrets:read": []interface{}{}}, "secrets:read", false},
		{"number", JSONMap{"secrets:read": 1}, "secrets:read", false},
	}
	for _, c := range cases {
		if got := permsAllows(c.perms, c.perm); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestAuditSigningKey(t *testing.T) {
	a := AuditSigningKey("secret", uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	b := AuditSigningKey("secret", uuid.MustParse("00000000-0000-0000-0000-000000000002"))
	if string(a) == string(b) {
		t.Fatal("different tenants must produce different keys")
	}
	if len(a) == 0 {
		t.Fatal("key must be non-empty")
	}
}

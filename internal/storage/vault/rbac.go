package vault

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BuiltinRoleNames are the names of the role seeds the engine creates
// lazily on first use.
const (
	BuiltinRoleAdmin    = "admin"
	BuiltinRoleOperator = "operator"
	BuiltinRoleReader   = "reader"
)

// BuiltinRolePermissions returns the default permission set for a
// built-in role.
func BuiltinRolePermissions(name string) map[string]interface{} {
	switch name {
	case BuiltinRoleAdmin:
		return map[string]interface{}{
			"secrets:create":               true,
			"secrets:read":                 true,
			"secrets:update":               true,
			"secrets:delete":               true,
			"secrets:rotate":               true,
			"tokens:create":                true,
			"tokens:revoke":                true,
			"dynamic_credentials:create":   true,
			"dynamic_credentials:generate": true,
			"dynamic_credentials:revoke":   true,
			"audit:read":                   true,
			"audit:export":                 true,
			"rbac:manage":                  true,
			"namespaces:manage":            true,
			"shares:manage":                true,
			"sso:manage":                   true,
			"break_glass:request":          true,
			"break_glass:approve":          true,
		}
	case BuiltinRoleOperator:
		return map[string]interface{}{
			"secrets:create":               true,
			"secrets:read":                 true,
			"secrets:update":               true,
			"secrets:delete":               true,
			"secrets:rotate":               true,
			"tokens:create":                true,
			"tokens:revoke":                true,
			"dynamic_credentials:create":   true,
			"dynamic_credentials:generate": true,
			"dynamic_credentials:revoke":   true,
			"audit:read":                   true,
		}
	case BuiltinRoleReader:
		return map[string]interface{}{
			"secrets:read":                 true,
			"audit:read":                   true,
			"dynamic_credentials:generate": true,
		}
	}
	return nil
}

// RBACEngine answers permission checks for a user, taking the user's
// role assignments, the target namespace, and the resource's
// namespace into account.
type RBACEngine struct {
	repo *Repository
}

// NewRBACEngine constructs an RBAC engine.
func NewRBACEngine(repo *Repository) *RBACEngine {
	return &RBACEngine{repo: repo}
}

// PermissionDecision is the result of a permission check.
type PermissionDecision struct {
	Allowed bool
	Role    string
	Reason  string
	Source  string // "builtin:<name>" or "role:<name>"
}

// EnsureBuiltinRoles lazily creates the admin / operator / reader
// roles for a tenant on first use.
func (e *RBACEngine) EnsureBuiltinRoles(ctx context.Context, tenantID uuid.UUID) error {
	for _, name := range []string{BuiltinRoleAdmin, BuiltinRoleOperator, BuiltinRoleReader} {
		existing, err := e.repo.ListRoles(ctx, tenantID)
		if err != nil {
			return err
		}
		found := false
		for _, r := range existing {
			if r.Name == name {
				found = true
				break
			}
		}
		if found {
			continue
		}
		perms := BuiltinRolePermissions(name)
		if perms == nil {
			continue
		}
		role := &VaultRole{
			TenantID:    tenantID,
			Name:        name,
			Description: fmt.Sprintf("Built-in %s role", name),
			Permissions: JSONMap(perms),
			IsBuiltin:   true,
		}
		if err := e.repo.CreateRole(ctx, role); err != nil {
			return err
		}
	}
	return nil
}

// Check evaluates whether the user has `permission` for a resource
// in `namespacePath`. Returns the first match (admin wins over others).
func (e *RBACEngine) Check(ctx context.Context, tenantID, userID uuid.UUID, permission, namespacePath string) (PermissionDecision, error) {
	assignments, err := e.repo.ListAssignmentsForUser(ctx, tenantID, userID)
	if err != nil {
		return PermissionDecision{}, err
	}
	// Resolve every assignment to its role.
	type boundRole struct {
		role  VaultRole
		scope string
	}
	var bound []boundRole
	for _, a := range assignments {
		role, err := e.repo.GetRole(ctx, a.RoleID, tenantID)
		if err != nil {
			return PermissionDecision{}, err
		}
		if role == nil {
			continue
		}
		bound = append(bound, boundRole{role: *role, scope: a.EffectiveScope()})
	}
	// Admin always wins.
	for _, b := range bound {
		if !b.role.IsBuiltin || b.role.Name != BuiltinRoleAdmin {
			continue
		}
		if permsAllows(b.role.Permissions, permission) {
			return PermissionDecision{
				Allowed: true,
				Role:    b.role.Name,
				Source:  "builtin:admin",
			}, nil
		}
	}
	for _, b := range bound {
		if !namespaceMatchesScope(namespacePath, b.scope) {
			continue
		}
		if permsAllows(b.role.Permissions, permission) {
			return PermissionDecision{
				Allowed: true,
				Role:    b.role.Name,
				Source:  fmt.Sprintf("role:%s", b.role.Name),
			}, nil
		}
	}
	return PermissionDecision{
		Allowed: false,
		Reason:  "no role grants this permission",
	}, nil
}

// CheckAny returns true if any of the permissions in the list is
// allowed. This is the helper used by handlers that need to gate on
// "either create or update".
func (e *RBACEngine) CheckAny(ctx context.Context, tenantID, userID uuid.UUID, namespacePath string, permissions ...string) (bool, error) {
	for _, p := range permissions {
		dec, err := e.Check(ctx, tenantID, userID, p, namespacePath)
		if err != nil {
			return false, err
		}
		if dec.Allowed {
			return true, nil
		}
	}
	return false, nil
}

// permsAllows reports whether the JSONB permissions map grants the
// given permission. The map may contain either `true` (allowed) or a
// string array of scopes (allowed when the namespace matches).
func permsAllows(perms JSONMap, permission string) bool {
	if perms == nil {
		return false
	}
	v, ok := perms[permission]
	if !ok {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true"
	case []interface{}:
		return len(t) > 0
	}
	return false
}

// namespaceMatchesScope returns true if `namespacePath` matches the
// given scope. A scope of "all" matches everything; a scope of
// "/foo" matches /foo exactly; a scope of "/foo" matches /foo/bar
// but not /foobar.
func namespaceMatchesScope(namespacePath, scope string) bool {
	if scope == "" || scope == "all" {
		return true
	}
	if namespacePath == scope {
		return true
	}
	return strings.HasPrefix(namespacePath, scope+"/")
}

// AuditSigningKey derives a stable per-tenant HMAC key for audit
// exports. It is intentionally deterministic so a tenant can rotate
// it by changing the system secret and the tenant ID.
func AuditSigningKey(systemSecret string, tenantID uuid.UUID) []byte {
	h := make([]byte, 0, len(systemSecret)+len(tenantID))
	h = append(h, []byte(systemSecret)...)
	h = append(h, []byte(tenantID.String())...)
	return h
}

// Suppress unused-import warnings if time isn't otherwise used.
var _ = time.Now

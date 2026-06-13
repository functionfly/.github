package vault

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DynamicCredentialMaterial is the in-memory representation of an
// issued dynamic credential. It is what callers hand back to
// applications. The storage-layer "DynamicCredential" is a different
// thing: a named template. We avoid the name collision by giving the
// in-memory type a distinct name.
type DynamicCredentialMaterial struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Database  string    `json:"database"`
	ExpiresAt time.Time `json:"expires_at"`
	LeaseID   string    `json:"lease_id"`
}

// DynamicSecretManager is the contract every supported backend must
// implement. Each method is responsible for the DB-specific CREATE /
// GRANT / DROP USER statements.
type DynamicSecretManager interface {
	// CreateUser mints a new temporary user with the given TTL and
	// grants it the supplied role(s). Returns the username + a freshly
	// generated password.
	CreateUser(ctx context.Context, opts CreateUserOptions) (username, password string, err error)

	// RenewUser extends an existing user's VALID UNTIL expiry.
	RenewUser(ctx context.Context, opts RenewUserOptions) error

	// DropUser removes a temporary user. The implementation must be
	// idempotent (return nil if the user doesn't exist).
	DropUser(ctx context.Context, opts DropUserOptions) error

	// DBType returns the manager's DB type identifier.
	DBType() DynamicSecretDBType
}

// CreateUserOptions is the cross-backend parameter pack for CreateUser.
type CreateUserOptions struct {
	Target       *DynamicSecretTarget
	Username     string
	Password     string
	TTL          time.Duration
	RoleTemplate string
	AllowedRoles []string
}

// RenewUserOptions is the cross-backend parameter pack for RenewUser.
type RenewUserOptions struct {
	Target   *DynamicSecretTarget
	Username string
	NewTTL   time.Duration
}

// DropUserOptions is the cross-backend parameter pack for DropUser.
type DropUserOptions struct {
	Target   *DynamicSecretTarget
	Username string
}

// DynamicSecretService ties the cross-cutting pieces together: it
// owns the per-DB-type manager registry, writes lease rows, and calls
// the right backend when a credential is generated.
type DynamicSecretService struct {
	repo            *Repository
	postgresManager DynamicSecretManager
	mysqlManager    DynamicSecretManager
}

// NewDynamicSecretService constructs a service with the standard set of
// backend managers. Future backends (Redis, Mongo) plug in here.
func NewDynamicSecretService(repo *Repository) *DynamicSecretService {
	return &DynamicSecretService{
		repo:            repo,
		postgresManager: &PostgresDynamicManager{},
		mysqlManager:    &MySQLDynamicManager{},
	}
}

// managerFor returns the registered manager for a target's DB type.
func (s *DynamicSecretService) managerFor(t *DynamicSecretTarget) (DynamicSecretManager, error) {
	switch t.DBType {
	case DynamicSecretDBPostgres:
		return s.postgresManager, nil
	case DynamicSecretDBMySQL:
		return s.mysqlManager, nil
	default:
		return nil, fmt.Errorf("unsupported db type %q", t.DBType)
	}
}

// Issue creates a new lease against a target, mints a temp user, and
// returns the lease + the credential material.
func (s *DynamicSecretService) Issue(
	ctx context.Context,
	cred *DynamicCredential,
	target *DynamicSecretTarget,
	ttl time.Duration,
	issuedBy *uuid.UUID,
	issuedIP string,
) (*DynamicCredentialLease, *DynamicCredentialMaterial, error) {
	if target.Status != "active" {
		return nil, nil, fmt.Errorf("target is disabled")
	}
	if ttl <= 0 {
		ttl = time.Duration(target.DefaultTTLSeconds) * time.Second
	}
	if ttl > time.Duration(target.MaxTTLSeconds)*time.Second {
		ttl = time.Duration(target.MaxTTLSeconds) * time.Second
	}

	manager, err := s.managerFor(target)
	if err != nil {
		return nil, nil, err
	}

	username := generateUsername(target.DBType)
	password, err := generatePassword(24)
	if err != nil {
		return nil, nil, err
	}

	role := cred.RoleTemplate
	if role == "" {
		// Default: pick the first allowed role on the target, if any.
		if len(target.AllowedRoles) > 0 {
			role = string(target.AllowedRoles[0])
		}
	}

	opts := CreateUserOptions{
		Target:       target,
		Username:     username,
		Password:     password,
		TTL:          ttl,
		RoleTemplate: role,
		AllowedRoles: []string(target.AllowedRoles),
	}
	if _, _, err := manager.CreateUser(ctx, opts); err != nil {
		s.repo.MarkTargetError(ctx, target.ID, err.Error())
		return nil, nil, fmt.Errorf("create user: %w", err)
	}

	leaseID := generateLeaseID()
	lease := &DynamicCredentialLease{
		LeaseID:      leaseID,
		CredentialID: cred.ID,
		TargetID:     target.ID,
		TenantID:     target.TenantID,
		DBUsername:   username,
		ExpiresAt:    time.Now().Add(ttl),
		IssuedTo:     issuedBy,
		IssuedIP:     issuedIP,
	}
	if err := s.repo.CreateLease(ctx, lease); err != nil {
		// Best-effort cleanup so we don't leak a DB user with no lease.
		_ = manager.DropUser(ctx, DropUserOptions{Target: target, Username: username})
		return nil, nil, fmt.Errorf("create lease: %w", err)
	}

	s.repo.MarkTargetUsed(ctx, target.ID)

	dc := &DynamicCredentialMaterial{
		Username:  username,
		Password:  password,
		Host:      target.Host,
		Port:      target.Port,
		Database:  target.DatabaseName,
		ExpiresAt: lease.ExpiresAt,
		LeaseID:   leaseID,
	}
	return lease, dc, nil
}

// Renew extends an active lease, returning the new expiry.
func (s *DynamicSecretService) Renew(
	ctx context.Context,
	lease *DynamicCredentialLease,
	target *DynamicSecretTarget,
	ttl time.Duration,
) (time.Time, error) {
	if !lease.IsActive(time.Now()) {
		return time.Time{}, fmt.Errorf("lease is no longer active")
	}
	if ttl > time.Duration(target.MaxTTLSeconds)*time.Second {
		ttl = time.Duration(target.MaxTTLSeconds) * time.Second
	}
	manager, err := s.managerFor(target)
	if err != nil {
		return time.Time{}, err
	}
	newExpires := time.Now().Add(ttl)
	if err := manager.RenewUser(ctx, RenewUserOptions{
		Target:   target,
		Username: lease.DBUsername,
		NewTTL:   ttl,
	}); err != nil {
		return time.Time{}, fmt.Errorf("renew user: %w", err)
	}
	if err := s.repo.RenewLease(ctx, lease.LeaseID, newExpires); err != nil {
		return time.Time{}, fmt.Errorf("renew lease: %w", err)
	}
	return newExpires, nil
}

// Revoke drops the DB user and marks the lease as revoked.
func (s *DynamicSecretService) Revoke(
	ctx context.Context,
	lease *DynamicCredentialLease,
	target *DynamicSecretTarget,
	reason string,
) error {
	manager, err := s.managerFor(target)
	if err != nil {
		return err
	}
	if err := manager.DropUser(ctx, DropUserOptions{
		Target:   target,
		Username: lease.DBUsername,
	}); err != nil {
		return fmt.Errorf("drop user: %w", err)
	}
	return s.repo.RevokeLease(ctx, lease.LeaseID, reason)
}

// generateUsername produces a unique, DB-safe username with a per-DB
// prefix. Each backend's identifier length limit differs so we stay
// conservative.
func generateUsername(dbType DynamicSecretDBType) string {
	prefix := "vault_"
	switch dbType {
	case DynamicSecretDBPostgres:
		prefix = "vault_p_"
	case DynamicSecretDBMySQL:
		prefix = "vault_m_"
	}
	suffix, err := randomLowerAlphaNum(16)
	if err != nil {
		// Fall back to timestamp-based suffix; collision risk is tiny
		// in a single-process sweep window.
		suffix = fmt.Sprintf("u%d", time.Now().UnixNano())
	}
	return prefix + suffix
}

func generatePassword(n int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[idx.Int64()]
	}
	return string(b), nil
}

func generateLeaseID() string {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		// best-effort fallback
		return fmt.Sprintf("lease-%d", time.Now().UnixNano())
	}
	return "lease_" + base64.RawURLEncoding.EncodeToString(b)
}

func randomLowerAlphaNum(n int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[idx.Int64()]
	}
	return string(b), nil
}

// quoteIdent quotes a SQL identifier (username / role) so it can be
// embedded in a DDL statement. Both PostgreSQL and MySQL accept
// double-quoted identifiers with the same escaping rule.
func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// quoteLiteral returns a single-quoted SQL string literal with
// embedded single quotes doubled.
func quoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}

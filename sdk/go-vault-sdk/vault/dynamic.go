package vault

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// DynamicSecretDBType identifies the underlying database engine.
type DynamicSecretDBType string

const (
	DynamicDBPostgres DynamicSecretDBType = "postgres"
	DynamicDBMySQL    DynamicSecretDBType = "mysql"
)

// DynamicTargetCreate is the input to DynamicTargets.Create.
type DynamicTargetCreate struct {
	Name              string              `json:"name"`
	Description       string              `json:"description,omitempty"`
	DBType            DynamicSecretDBType `json:"db_type"`
	Host              string              `json:"host"`
	Port              int                 `json:"port"`
	DatabaseName      string              `json:"database_name"`
	AdminUsername     string              `json:"admin_username"`
	AdminPassword     string              `json:"admin_password"`
	SSLMode           string              `json:"ssl_mode,omitempty"`
	AllowedRoles      []string            `json:"allowed_roles,omitempty"`
	DefaultTTLSeconds int                 `json:"default_ttl_seconds,omitempty"`
	MaxTTLSeconds     int                 `json:"max_ttl_seconds,omitempty"`
}

// DynamicTarget is the API projection (never includes the password).
type DynamicTarget struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	Description       string              `json:"description,omitempty"`
	DBType            DynamicSecretDBType `json:"db_type"`
	Host              string              `json:"host"`
	Port              int                 `json:"port"`
	DatabaseName      string              `json:"database_name"`
	AdminUsername     string              `json:"admin_username"`
	SSLMode           string              `json:"ssl_mode"`
	AllowedRoles      []string            `json:"allowed_roles,omitempty"`
	DefaultTTLSeconds int                 `json:"default_ttl_seconds"`
	MaxTTLSeconds     int                 `json:"max_ttl_seconds"`
	Status            string              `json:"status"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

// DynamicTargetsService manages database target configurations.
type DynamicTargetsService struct{ client *Client }

// Create registers a new DB target. AdminPassword is encrypted
// server-side and never returned.
func (s *DynamicTargetsService) Create(ctx context.Context, in DynamicTargetCreate) (*DynamicTarget, error) {
	if in.AdminPassword == "" {
		return nil, fmt.Errorf("admin_password is required")
	}
	switch in.DBType {
	case DynamicDBPostgres, DynamicDBMySQL:
		// ok
	default:
		return nil, fmt.Errorf("invalid db_type %q (expected postgres or mysql)", in.DBType)
	}
	var out DynamicTarget
	if err := s.client.do(ctx, "POST", "/v1/vault/dynamic-secret-targets", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all configured targets.
func (s *DynamicTargetsService) List(ctx context.Context) ([]DynamicTarget, error) {
	var out struct {
		Targets []DynamicTarget `json:"targets"`
		Total   int             `json:"total"`
	}
	if err := s.client.do(ctx, "GET", "/v1/vault/dynamic-secret-targets", nil, &out); err != nil {
		return nil, err
	}
	return out.Targets, nil
}

// Delete disables a target.
func (s *DynamicTargetsService) Delete(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/v1/vault/dynamic-secret-targets/"+url.PathEscape(id), nil, nil)
}

// Test performs a connection smoke-test (issues + revokes a 60s
// credential against the target).
func (s *DynamicTargetsService) Test(ctx context.Context, id string) error {
	return s.client.do(ctx, "POST", "/v1/vault/dynamic-secret-targets/"+url.PathEscape(id)+"/test", nil, nil)
}

// DynamicCredentialCreate is the input to DynamicCredentials.Create.
type DynamicCredentialCreate struct {
	TargetID      string `json:"target_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	RoleTemplate  string `json:"role_template,omitempty"`
	TTLSeconds    int    `json:"ttl_seconds,omitempty"`
	MaxTTLSeconds int    `json:"max_ttl_seconds,omitempty"`
}

// DynamicCredential is the API projection of a credential template.
type DynamicCredential struct {
	ID            string    `json:"id"`
	TargetID      string    `json:"target_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	RoleTemplate  string    `json:"role_template,omitempty"`
	TTLSeconds    int       `json:"ttl_seconds"`
	MaxTTLSeconds int       `json:"max_ttl_seconds"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GeneratedCredential is the response from DynamicCredentials.Generate.
type GeneratedCredential struct {
	LeaseID    string            `json:"lease_id"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	Database   string            `json:"database"`
	ExpiresAt  time.Time         `json:"expires_at"`
	Credential DynamicCredential `json:"credential"`
	Target     DynamicTarget     `json:"target"`
}

// DynamicCredentialsService manages named credential templates and
// issues one-off credentials against them.
type DynamicCredentialsService struct{ client *Client }

// Create registers a new credential template.
func (s *DynamicCredentialsService) Create(ctx context.Context, in DynamicCredentialCreate) (*DynamicCredential, error) {
	if in.TargetID == "" {
		return nil, fmt.Errorf("target_id is required")
	}
	var out DynamicCredential
	if err := s.client.do(ctx, "POST", "/v1/vault/dynamic-credentials", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Generate mints a fresh credential against the template.
type GenerateOptions struct {
	TTLSeconds int `json:"ttl_seconds,omitempty"`
}

func (s *DynamicCredentialsService) Generate(ctx context.Context, credentialID string, opts GenerateOptions) (*GeneratedCredential, error) {
	var body interface{}
	if opts.TTLSeconds > 0 {
		body = opts
	}
	var out GeneratedCredential
	path := "/v1/vault/dynamic-credentials/" + url.PathEscape(credentialID) + "/generate"
	if err := s.client.do(ctx, "POST", path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokeAll revokes all active leases for a credential template.
func (s *DynamicCredentialsService) RevokeAll(ctx context.Context, credentialID string) error {
	path := "/v1/vault/dynamic-credentials/" + url.PathEscape(credentialID) + "/revoke"
	return s.client.do(ctx, "POST", path, nil, nil)
}

// LeasesService manages the lease lifecycle.
type LeasesService struct{ client *Client }

// RenewOptions carries the new TTL.
type RenewOptions struct {
	TTLSeconds int `json:"ttl_seconds,omitempty"`
}

// Renew extends an active lease's expiry.
func (l *LeasesService) Renew(ctx context.Context, credentialID, leaseID string, opts RenewOptions) (time.Time, error) {
	var out struct {
		LeaseID   string    `json:"lease_id"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	path := fmt.Sprintf("/v1/vault/dynamic-credentials/%s/leases/%s/renew",
		url.PathEscape(credentialID), url.PathEscape(leaseID))
	if err := l.client.do(ctx, "POST", path, opts, &out); err != nil {
		return time.Time{}, err
	}
	return out.ExpiresAt, nil
}

// Revoke drops the underlying DB user and marks the lease revoked.
func (l *LeasesService) Revoke(ctx context.Context, credentialID, leaseID string) error {
	path := fmt.Sprintf("/v1/vault/dynamic-credentials/%s/leases/%s/revoke",
		url.PathEscape(credentialID), url.PathEscape(leaseID))
	return l.client.do(ctx, "POST", path, nil, nil)
}

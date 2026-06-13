package vault

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// SecretType is the kind of secret being stored.
type SecretType string

const (
	SecretTypeAPIKey      SecretType = "api_key"
	SecretTypeOAuthToken  SecretType = "oauth_token"
	SecretTypePassword    SecretType = "password"
	SecretTypeCertificate SecretType = "certificate"
)

// Valid reports whether the SecretType is recognised.
func (s SecretType) Valid() bool {
	switch s {
	case SecretTypeAPIKey, SecretTypeOAuthToken, SecretTypePassword, SecretTypeCertificate:
		return true
	}
	return false
}

// EncryptedData is the zero-knowledge ciphertext payload. All fields
// are base64-encoded bytes produced by the caller's encryption layer.
type EncryptedData struct {
	Ciphertext string `json:"ciphertext"`
	IV         string `json:"iv"`
	Salt       string `json:"salt"`
	Tag        string `json:"tag"`
	KeyVersion int    `json:"key_version"`
}

// Secret is the public projection of a stored secret.
type Secret struct {
	ID            string        `json:"id"`
	TenantID      string        `json:"tenant_id"`
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	SecretType    SecretType    `json:"secret_type"`
	EncryptedData EncryptedData `json:"encrypted_data"`
	AccessCount   int           `json:"access_count"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	ExpiresAt     *time.Time    `json:"expires_at,omitempty"`
}

// SecretCreate is the input to Secrets.Create.
type SecretCreate struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	SecretType    SecretType             `json:"secret_type"`
	EncryptedData EncryptedData          `json:"encrypted_data"`
	Scopes        []string               `json:"scopes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SecretListOptions controls Secrets.List.
type SecretListOptions struct {
	Limit      int
	Offset     int
	SecretType SecretType
}

// SecretList is the response from Secrets.List.
type SecretList struct {
	Secrets []Secret `json:"secrets"`
	Total   int64    `json:"total"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
}

// SecretsService exposes the secret-management API.
type SecretsService struct{ client *Client }

// Create persists a new secret.
func (s *SecretsService) Create(ctx context.Context, in SecretCreate) (*Secret, error) {
	if !in.SecretType.Valid() {
		return nil, fmt.Errorf("invalid secret_type %q", in.SecretType)
	}
	var out Secret
	if err := s.client.do(ctx, "POST", "/v1/vault/secrets", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get fetches a single secret by ID.
func (s *SecretsService) Get(ctx context.Context, id string) (*Secret, error) {
	var out Secret
	if err := s.client.do(ctx, "GET", "/v1/vault/secrets/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies the metadata (name, description, scopes) of a secret.
type SecretUpdate struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Scopes      *[]string `json:"scopes,omitempty"`
}

func (s *SecretsService) Update(ctx context.Context, id string, in SecretUpdate) (*Secret, error) {
	var out Secret
	if err := s.client.do(ctx, "PATCH", "/v1/vault/secrets/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Rotate replaces the encrypted value (creates a new version).
type SecretRotate struct {
	EncryptedData EncryptedData `json:"encrypted_data"`
	Reason        string        `json:"reason,omitempty"`
}

func (s *SecretsService) Rotate(ctx context.Context, id string, in SecretRotate) (*Secret, error) {
	var out Secret
	if err := s.client.do(ctx, "PATCH", "/v1/vault/secrets/"+url.PathEscape(id)+"/rotate", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete soft-deletes a secret and revokes its tokens.
func (s *SecretsService) Delete(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/v1/vault/secrets/"+url.PathEscape(id), nil, nil)
}

// List returns a paginated list of secrets.
func (s *SecretsService) List(ctx context.Context, opts SecretListOptions) (*SecretList, error) {
	v := url.Values{}
	if opts.Limit > 0 {
		v.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		v.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.SecretType != "" {
		v.Set("secret_type", string(opts.SecretType))
	}
	path := "/v1/vault/secrets"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out SecretList
	if err := s.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

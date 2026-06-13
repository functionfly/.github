package vault

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// TokenCreate is the input to Tokens.Create.
type TokenCreate struct {
	SecretID       string   `json:"secret_id"`
	ExpiresInHours int      `json:"expires_in_hours"`
	Scopes         []string `json:"scopes,omitempty"`
	Name           string   `json:"name,omitempty"`
}

// Token is the response from Tokens.Create.
type Token struct {
	TokenID   string    `json:"token_id"`
	Token     string    `json:"token"` // plaintext, shown only once
	SecretID  string    `json:"secret_id"`
	Name      string    `json:"name,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenInfo is a non-secret projection of a token (for listing).
type TokenInfo struct {
	ID            string     `json:"id"`
	SecretID      string     `json:"secret_id"`
	Name          string     `json:"name,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	IsRevoked     bool       `json:"is_revoked"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	RevokedReason string     `json:"revoked_reason,omitempty"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	UseCount      int        `json:"use_count"`
	CreatedAt     time.Time  `json:"created_at"`
}

// TokenList is the response from Tokens.List.
type TokenList struct {
	Tokens []TokenInfo `json:"tokens"`
	Total  int64       `json:"total"`
}

// TokensService manages runtime access tokens.
type TokensService struct{ client *Client }

// Create mints a new access token. The plaintext token is returned
// exactly once in the response — store it immediately.
func (t *TokensService) Create(ctx context.Context, in TokenCreate) (*Token, error) {
	if in.SecretID == "" {
		return nil, fmt.Errorf("secret_id is required")
	}
	if in.ExpiresInHours <= 0 {
		in.ExpiresInHours = 24
	}
	if in.ExpiresInHours > 8760 {
		return nil, fmt.Errorf("expires_in_hours cannot exceed 8760 (1 year)")
	}
	var out Token
	path := "/v1/vault/secrets/" + url.PathEscape(in.SecretID) + "/tokens"
	if err := t.client.do(ctx, "POST", path, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all tokens for a given secret.
func (t *TokensService) List(ctx context.Context, secretID string) (*TokenList, error) {
	if secretID == "" {
		return nil, fmt.Errorf("secret_id is required")
	}
	var out TokenList
	path := "/v1/vault/secrets/" + url.PathEscape(secretID) + "/tokens"
	if err := t.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Revoke revokes a token by its ID.
func (t *TokensService) Revoke(ctx context.Context, tokenID string) error {
	return t.client.do(ctx, "DELETE", "/v1/vault/tokens/"+url.PathEscape(tokenID), nil, nil)
}

// Package vault is the official Go SDK for the FunctionFly zero-knowledge
// secrets vault.
//
// Quick start:
//
//	client, err := vault.NewClient("https://api.functionfly.com",
//	    vault.WithToken("fnly_xxx"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Client-side encryption happens in the dashboard / per-passphrase
//	// helpers. The SDK accepts a pre-encrypted payload.
//	secret, err := client.Secrets.Create(ctx, vault.SecretCreate{
//	    Name:       "STRIPE_API_KEY",
//	    SecretType: vault.SecretTypeAPIKey,
//	    EncryptedData: vault.EncryptedData{
//	        Ciphertext: ct,
//	        IV:         iv,
//	        Salt:       salt,
//	        Tag:        tag,
//	        KeyVersion: 2, // 1=PBKDF2, 2=Argon2id
//	    },
//	})
//
// See the README for full examples and the dynamic-credentials API.
package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultBaseURL is the public FunctionFly API endpoint.
const DefaultBaseURL = "https://api.functionfly.com"

// Version of the SDK. Bumped per release.
const Version = "0.1.0"

// Client is a thin, dependency-light HTTP wrapper around the
// FunctionFly vault REST API. The SDK never sees plaintext secret
// values — all encryption is performed by the caller.
type Client struct {
	BaseURL   string
	Token     string
	HTTP      *http.Client
	UserAgent string

	// Sub-services.
	Secrets            *SecretsService
	Tokens             *TokensService
	DynamicCredentials *DynamicCredentialsService
	DynamicTargets     *DynamicTargetsService
	Leases             *LeasesService
	Audit              *AuditService
}

// Option configures a Client.
type Option func(*Client)

// WithToken sets the bearer token used to authenticate.
func WithToken(token string) Option {
	return func(c *Client) { c.Token = token }
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.BaseURL = url }
}

// WithHTTPClient overrides the underlying HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.HTTP = h }
}

// NewClient constructs a Client with sensible defaults.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	c := &Client{
		BaseURL:   baseURL,
		Token:     "",
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		UserAgent: "functionfly-go-vault-sdk/" + Version,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Secrets = &SecretsService{client: c}
	c.Tokens = &TokensService{client: c}
	c.DynamicCredentials = &DynamicCredentialsService{client: c}
	c.DynamicTargets = &DynamicTargetsService{client: c}
	c.Leases = &LeasesService{client: c}
	c.Audit = &AuditService{client: c}
	return c, nil
}

// ============================================================================
// Errors
// ============================================================================

// APIError is the structured error returned by the vault API.
type APIError struct {
	Status    int    `json:"-"`
	Code      string `json:"code,omitempty"`
	ErrString string `json:"error,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = e.ErrString
	}
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", e.Status)
	}
	return msg
}

// ============================================================================
// Internal request helper
// ============================================================================

// do is the shared JSON request/response helper.
func (c *Client) do(ctx context.Context, method, path string, in, out interface{}) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(data)
	}
	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("network: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		apiErr := &APIError{Status: resp.StatusCode}
		_ = json.Unmarshal(respBody, apiErr)
		if apiErr.Message == "" && apiErr.ErrString == "" {
			apiErr.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return apiErr
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

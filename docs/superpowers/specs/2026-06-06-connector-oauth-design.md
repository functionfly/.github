# Connector OAuth Implementation - Design Spec

## Status
- Date: 2026-06-06
- Implemented: No

## Overview

Implement full OAuth callback flow for connector integrations (GitHub, Slack, Notion, Gmail, Linear) with zero-knowledge encryption and server-side token refresh capability.

## Architecture

### OAuth Flow (Sequence)

```
1. User clicks "Connect" in dashboard
   → GET /v1/connectors/oauth-url?slug=github
   ← { oauth_url, state }

2. Popup opens to OAuth provider authorization page
   → User authorizes
   → Provider redirects to https://app.functionfly.com/connectors/callback?code=xxx&state=yyy

3. ConnectorCallbackPage (React) receives redirect
   → Validates state with backend: POST /v1/connectors/callback { code, state }
   ← { tokens, encrypted_blob }

4. Frontend encrypts tokens with vault key (client-side)
   → POST /v1/connectors/link { connector_slug, encrypted_credentials, display_name }

5. Backend stores encrypted credentials
   → User connector ready for sync
```

### Token Storage Strategy (Hybrid Zero-Knowledge)

Two encrypted blobs stored in `user_connectors.encrypted_credentials`:

1. **`user_vault`**: Client-side encrypted with user's vault passphrase (zero-knowledge)
   - User can decrypt with their passphrase
   - Server cannot decrypt

2. **`server_sync`**: Server-side encrypted with a per-tenant master key
   - Server can decrypt for background sync/token refresh
   - User cannot decrypt (intentional - server automation only)

```json
{
  "user_vault": { "ciphertext": "...", "iv": "...", "tag": "...", "salt": "...", "key_version": 2 },
  "server_sync": { "ciphertext": "...", "iv": "...", "tag": "...", "salt": "...", "key_version": 1 },
  "provider": "github",
  "linked_at": "2026-06-06T22:00:00Z"
}
```

### Database Schema

```sql
-- OAuth states for CSRF protection
CREATE TABLE IF NOT EXISTS connector_oauth_states (
    state VARCHAR(64) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    connector_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '10 minutes',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for cleanup
CREATE INDEX IF NOT EXISTS idx_connector_oauth_states_expires ON connector_oauth_states(expires_at);
```

## Components

### 1. Backend: `internal/storage/connector_repository.go`

New methods:

```go
// GetOAuthState retrieves and validates an OAuth state (does NOT consume)
func (r *ConnectorRepository) GetOAuthState(ctx context.Context, state string) (*OAuthState, error)

// ConsumeOAuthState retrieves and deletes an OAuth state (one-time use)
func (r *ConnectorRepository) ConsumeOAuthState(ctx context.Context, state string) (*OAuthState, error)

// CleanupExpiredOAuthStates removes states older than 10 minutes
func (r *ConnectorRepository) CleanupExpiredOAuthStates(ctx context.Context) error

type OAuthState struct {
    State       string
    TenantID    uuid.UUID
    ConnectorID uuid.UUID
    ExpiresAt   time.Time
}
```

### 2. Backend: `internal/api/handlers/connectors/handler.go`

Changes to `HandleOAuthCallback`:

```go
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
    // 1. Auth check
    // 2. Parse { code, state }
    // 3. ConsumeOAuthState (validates state, deletes after use)
    // 4. Exchange code for tokens with OAuth provider
    // 5. Encrypt tokens with server-side key (for sync)
    // 6. Return { tokens, server_encrypted_blob, connector_slug }
}
```

New method `exchangeOAuthCode`:

```go
func (h *Handler) exchangeOAuthCode(ctx context.Context, slug, code, redirectURI string) (*TokenResponse, error)

// TokenResponse contains all OAuth tokens
type TokenResponse struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int
    TokenType    string
    // Provider-specific fields (e.g., bot_token for Slack)
    Raw map[string]interface{}
}
```

### 3. Backend: Server-side encryption

`internal/crypto/server_encryption.go` (new file):

```go
// ServerEncrypt encrypts data with a tenant-specific key
// Uses AES-256-GCM with key derived from SERVER_MASTER_KEY + tenant_id
func ServerEncrypt(plaintext []byte, tenantID uuid.UUID) (ciphertext, iv, salt, tag []byte, err error)

// ServerDecrypt decrypts data encrypted with ServerEncrypt
func ServerDecrypt(ciphertext, iv, salt, tag []byte, tenantID uuid.UUID) ([]byte, error)
```

### 4. Frontend: Callback Page

`web/dashboard/src/pages/ConnectorsCallbackPage/index.tsx`:

- Route: `/connectors/callback`
- Receives OAuth redirect with ?code=xxx&state=yyy&slug=github
- Validates state server-side via POST /v1/connectors/callback
- Receives encrypted server blob and raw tokens
- Encrypts tokens with vault key (user passphrase)
- Calls linkConnector with both encrypted blobs
- Displays success/error UI
- Sends postMessage to opener window with result

### 5. Frontend: OAuth Popup Flow Update

Update `IntegrationsSettingsTab.tsx` to handle the new callback page:

- Instead of opening OAuth URL directly in popup
- Open `/connectors/callback?slug=xxx&oauth_url=yyy`
- This page handles the redirect and postMessage

## OAuth Provider Implementation

### Token Exchange Per Provider

```go
// GitHub
POST https://github.com/login/oauth/access_token
Client-ID, Client-Secret, code, redirect_uri

// Slack
POST https://slack.com/api/oauth.v2.access
client_id, client_secret, code, redirect_uri

// Notion
POST https://api.notion.com/v1/oauth/token
Authorization: Basic base64(client_id:client_secret)
code, redirect_uri, grant_type=authorization_code

// Google (Gmail)
POST https://oauth2.googleapis.com/token
client_id, client_secret, code, redirect_uri, grant_type=authorization_code

// Linear
POST https://api.linear.app/oauth/token
client_id, client_secret, code, redirect_uri
```

### Provider-Specific Responses

| Provider | Access Token Location | Refresh Token |
|----------|---------------------|---------------|
| GitHub | `access_token` | `refresh_token` (if scope includes) |
| Slack | `access_token` + `bot_token` | `refresh_token` |
| Notion | `access_token` | `refresh_token` |
| Gmail | `access_token` | `refresh_token` |
| Linear | `access_token` | `refresh_token` |

## Security Considerations

1. **State validation**: OAuth state must be validated and consumed (one-time use)
2. **State expiration**: States expire after 10 minutes
3. **Redirect URI validation**: Each provider must have their redirect URI configured
4. **Token transport**: Tokens never stored in plaintext server-side
5. **PKCE support**: For providers that support PKCE, use it
6. **Cleanup job**: Background job to clean expired OAuth states

## Error Handling

| Error | User Message | Action |
|-------|-------------|--------|
| State expired/invalid | "OAuth session expired. Please try again." | Re-initiate OAuth |
| Code exchange failed | "Failed to authorize with {provider}. Please try again." | Re-initiate OAuth |
| Network error | "Network error. Please check your connection." | Retry with backoff |
| Provider error | "{Provider} returned an error: {details}" | Show provider message |

## File Changes Summary

### New Files
- `web/dashboard/src/pages/ConnectorsCallbackPage/index.tsx` - React callback page
- `internal/crypto/server_encryption.go` - Server-side encryption utilities

### Modified Files
- `internal/storage/connector_repository.go` - Add OAuth state methods
- `internal/api/handlers/connectors/handler.go` - Full OAuth callback implementation
- `web/dashboard/src/api/connectors.ts` - Add callback API
- `web/dashboard/src/pages/SettingsPage/components/IntegrationsSettingsTab.tsx` - Update popup flow

### New Files for Migration
- `migrations/YYYYMMDDHHMMSS_connector_oauth_states.sql` - Create table

## Testing

1. Test OAuth flow end-to-end with each provider (manual)
2. Verify state validation (expired, used, invalid)
3. Verify encryption/decryption roundtrip
4. Verify token refresh works for sync
5. Test error scenarios (network, provider errors)
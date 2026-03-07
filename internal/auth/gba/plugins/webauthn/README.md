# GoBetterAuth WebAuthn Plugin (Phase 3)

WebAuthn/Passkeys plugin for GoBetterAuth providing passwordless authentication support.

## Features

- **Passkey Registration**: Register platform authenticators (Touch ID, Windows Hello) and security keys
- **Discoverable Credentials**: Support for passkeys that don't require username entry
- **Multiple Passkeys**: Users can register multiple passkeys per account
- **Secure Storage**: Credentials stored with sign count tracking to prevent replay attacks
- **Session Management**: Temporary sessions for registration and authentication ceremonies

## Installation

The WebAuthn plugin is automatically included with GoBetterAuth. No additional installation required.

## Configuration

Add the following to your `.env` file:

```bash
# WebAuthn Master Switch
GBA_WEBAUTHN_ENABLED=true

# Relying Party Configuration
WEBAUTHN_RP_DISPLAY_NAME=FunctionFly    # Display name shown to users
WEBAUTHN_RP_ID=functionfly.com          # Domain without scheme
WEBAUTHN_RP_ORIGIN=https://app.functionfly.com  # Full origin with scheme

# Session Configuration
WEBAUTHN_SESSION_TIMEOUT=5m             # Timeout for ceremonies
```

### Relying Party Configuration

- **RPDisplayName**: The human-readable name displayed during passkey registration
- **RPID**: The domain identifier (e.g., `functionfly.com`). Must match the actual domain.
- **RPOrigin**: The full origin including scheme (e.g., `https://app.functionfly.com`)

## Usage

### Initialize the Plugin

```go
import (
    "github.com/functionfly/internal/auth/gba/plugins/webauthn"
)

// Create WebAuthn plugin
webauthnConfig := &webauthn.WebAuthnConfig{
    RPDisplayName: "FunctionFly",
    RPID:          "functionfly.com",
    RPOrigin:      "https://app.functionfly.com",
    SessionTimeout: 5 * time.Minute,
}

webauthnPlugin, err := webauthn.New(db, webauthnConfig, logger)
if err != nil {
    log.Fatal(err)
}
```

### Setup HTTP Handlers

```go
// Create handler
webauthnHandler := webauthn.NewHandler(webauthnPlugin, logger)

// Register routes (typically under /v1/auth/webauthn)
mux := http.NewServeMux()
webauthnHandler.SetupRoutes(mux, "/v1/auth/webauthn")
```

## API Endpoints

### Register a Passkey

#### Begin Registration
```http
POST /v1/auth/webauthn/register/begin
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "MacBook Pro",
  "authenticator_type": "platform",  // or "cross-platform"
  "resident_key": true               // for discoverable credentials
}

Response:
{
  "sessionId": "550e8400-e29b-41d4-a716-446655440000",
  "options": {
    "rp": { "name": "FunctionFly", "id": "functionfly.com" },
    "user": { ... },
    "challenge": "...",
    "pubKeyCredParams": [...],
    ...
  }
}
```

#### Complete Registration
```http
POST /v1/auth/webauthn/register/complete
Authorization: Bearer <token>
Content-Type: application/json

{
  "sessionId": "550e8400-e29b-41d4-a716-446655440000",
  "response": {
    "id": "...",
    "rawId": "...",
    "type": "public-key",
    "response": { ... }
  }
}

Response:
{
  "credentialId": "550e8400-e29b-41d4-a716-446655440001",
  "name": "Passkey",
  "createdAt": "2024-03-07T12:00:00Z",
  "message": "Passkey registered successfully"
}
```

### Authenticate with Passkey

#### Begin Authentication
```http
POST /v1/auth/webauthn/authenticate/begin
Authorization: Bearer <token>

Response:
{
  "sessionId": "550e8400-e29b-41d4-a716-446655440002",
  "options": {
    "challenge": "...",
    "rpId": "functionfly.com",
    "allowCredentials": [...],
    ...
  }
}
```

#### Complete Authentication
```http
POST /v1/auth/webauthn/authenticate/complete
Content-Type: application/json

{
  "sessionId": "550e8400-e29b-41d4-a716-446655440002",
  "response": {
    "id": "...",
    "rawId": "...",
    "type": "public-key",
    "response": { ... }
  }
}

Response:
{
  "userId": "550e8400-e29b-41d4-a716-446655440003",
  "email": "user@example.com",
  "message": "Authentication successful"
}
```

### Discoverable Authentication (Passkey without username)

#### Begin Discoverable Authentication
```http
POST /v1/auth/webauthn/authenticate/discoverable/begin

Response:
{
  "sessionId": "550e8400-e29b-41d4-a716-446655440004",
  "options": { ... }
}
```

#### Complete Discoverable Authentication
```http
POST /v1/auth/webauthn/authenticate/discoverable/complete
Content-Type: application/json

{
  "sessionId": "550e8400-e29b-41d4-a716-446655440004",
  "response": { ... }
}

Response:
{
  "userId": "550e8400-e29b-41d4-a716-446655440003",
  "email": "user@example.com",
  "message": "Authentication successful"
}
```

### Manage Passkeys

#### List Credentials
```http
GET /v1/auth/webauthn/credentials
Authorization: Bearer <token>

Response:
{
  "credentials": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "MacBook Pro",
      "createdAt": "2024-03-07T12:00:00Z",
      "lastUsedAt": "2024-03-07T15:30:00Z",
      "signCount": 42
    }
  ]
}
```

#### Rename Credential
```http
PATCH /v1/auth/webauthn/credentials/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "MacBook Pro (Work)"
}

Response:
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "name": "MacBook Pro (Work)",
  "message": "Credential updated successfully"
}
```

#### Delete Credential
```http
DELETE /v1/auth/webauthn/credentials/{id}
Authorization: Bearer <token>

Response:
{
  "deleted": true,
  "message": "Credential deleted successfully"
}
```

### Get Status

```http
GET /v1/auth/webauthn/status
Authorization: Bearer <token>

Response:
{
  "enabled": true,
  "credentialCount": 2
}
```

## Frontend Integration

### Registering a Passkey

```javascript
// 1. Begin registration
const beginResp = await fetch('/v1/auth/webauthn/register/begin', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: 'My Device',
    authenticator_type: 'platform',
    resident_key: true
  })
});

const { sessionId, options } = await beginResp.json();

// 2. Create credential with WebAuthn API
const credential = await navigator.credentials.create({
  publicKey: options
});

// 3. Complete registration
const completeResp = await fetch('/v1/auth/webauthn/register/complete', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    sessionId,
    response: credential
  })
});
```

### Authenticating with a Passkey

```javascript
// 1. Begin authentication
const beginResp = await fetch('/v1/auth/webauthn/authenticate/begin', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const { sessionId, options } = await beginResp.json();

// 2. Get credential with WebAuthn API
const assertion = await navigator.credentials.get({
  publicKey: options
});

// 3. Complete authentication
const completeResp = await fetch('/v1/auth/webauthn/authenticate/complete', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    sessionId,
    response: assertion
  })
});

const { userId, email } = await completeResp.json();
```

## Security Considerations

1. **Sign Count Tracking**: The plugin tracks signature counters to prevent replay attacks. Each credential has a sign count that is updated after each authentication.

2. **Session Expiration**: Registration and authentication sessions expire after the configured timeout (default: 5 minutes).

3. **Credential Uniqueness**: Credential IDs are unique and enforced at the database level.

4. **Attestation**: The plugin supports attestation verification for enterprise deployments requiring hardware authenticator verification.

5. **Transport Security**: All WebAuthn endpoints should be served over HTTPS in production.

## Database Schema

### gba_webauthn_credentials
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | Foreign key to gba_users |
| credential_id | BYTEA | Unique WebAuthn credential identifier |
| public_key | BYTEA | Public key for verification |
| attestation_type | VARCHAR(50) | Attestation type |
| transport | TEXT[] | Supported transport methods |
| flags | INT | Credential flags |
| authenticator | JSONB | Authenticator metadata |
| sign_count | INT | Signature counter |
| name | VARCHAR(255) | User-friendly name |
| created_at | TIMESTAMP | Creation time |
| last_used_at | TIMESTAMP | Last authentication time |

### gba_webauthn_sessions
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | Foreign key to gba_users |
| challenge | VARCHAR(255) | Base64-encoded session data |
| operation | VARCHAR(20) | "registration" or "authentication" |
| expires_at | TIMESTAMP | Session expiration time |
| created_at | TIMESTAMP | Creation time |

## Migration

Run the migration to create WebAuthn tables:

```bash
# Using golang-migrate
migrate -path migrations -database "postgres://user:pass@localhost/dbname" up

# Or apply manually
psql -d functionfly -f migrations/20260307000004_add_gba_webauthn_tables.up.sql
```

## Testing

Example test flow:

```go
func TestWebAuthnFlow(t *testing.T) {
    ctx := context.Background()
    userID := uuid.MustParse("...")

    // 1. Begin registration
    regResp, err := webauthnPlugin.BeginRegistration(ctx, userID, BeginRegistrationRequest{
        Name: "Test Device",
    })
    require.NoError(t, err)
    require.NotEmpty(t, regResp.SessionID)
    require.NotNil(t, regResp.Options)

    // 2. Simulate credential creation with test client
    // (Would use a WebAuthn test client or mock)

    // 3. Begin authentication
    authResp, err := webauthnPlugin.BeginAuthentication(ctx, userID)
    require.NoError(t, err)
    require.NotEmpty(t, authResp.SessionID)

    // 4. Simulate credential assertion
    // (Would use a WebAuthn test client or mock)
}
```

## Browser Support

- Chrome/Edge 67+
- Firefox 60+
- Safari 13+
- iOS Safari 13.3+
- Chrome Android 70+

## License

MIT License - Same as GoBetterAuth
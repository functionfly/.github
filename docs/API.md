# FunctionFly API Reference

Complete reference for the FunctionFly serverless platform API.

---

## Base URL

| Environment | URL |
|-------------|-----|
| Production | `https://api.functionfly.com` |
| Staging | `https://api.staging.functionfly.com` |
| Development | `http://localhost:8080` |

---

## API Versions

| Version | Status | Endpoint |
|---------|--------|----------|
| v1 | Active | `/v1/*` |
| v2 | Beta | `/v2/*` |

---

## Authentication

### OAuth Authentication

```bash
# Get OAuth providers
GET /v1/auth/oauth/providers

# Get OAuth URL for specific provider
GET /v1/auth/oauth/url?provider=github&redirect_uri=...

# OAuth callback (handled by client)
GET /v1/auth/oauth/{provider}/callback?code=...&state=...
```

### API Key Authentication

```bash
# Authenticate with API key
POST /v1/auth/api-key
Authorization: Bearer <api_key>
```

### Session Authentication

```bash
# Login with email/password
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}

# Signup
POST /v1/auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password",
  "username": "username"
}

# Refresh token
POST /v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "..."
}

# Validate token
GET /v1/auth/validate

# Logout
POST /v1/auth/logout

# Password reset request
POST /v1/auth/password-reset
Content-Type: application/json

{
  "email": "user@example.com"
}

# Password reset confirmation
POST /v1/auth/password-reset/confirm
Content-Type: application/json

{
  "token": "...",
  "new_password": "..."
}
```

---

## Users

### User Profile

```bash
# Get current user
GET /v1/users/me

# Update current user
PATCH /v1/users/me
Content-Type: application/json

{
  "display_name": "...",
  "bio": "...",
  "avatar_url": "..."
}

# Get public profile
GET /v1/users/{username}

# Get user analytics
GET /v1/users/{username}/analytics

# Get user achievements
GET /v1/users/{username}/achievements

# Get user activity
GET /v1/users/{username}/activity

# Get user skills
GET /v1/users/{username}/skills

# Add user skill
POST /v1/users/me/skills
Content-Type: application/json

{
  "skill": "Go"
}

# Remove user skill
DELETE /v1/users/me/skills/{id}
```

### User Settings

```bash
# Get user settings
GET /v1/users/me/settings
GET /v1/users/{username}/settings

# Update profile settings
PATCH /v1/users/me/settings/profile
PATCH /v1/users/{username}/settings/profile

# Update notification settings
PATCH /v1/users/me/settings/notifications
PATCH /v1/users/{username}/settings/notifications

# Update privacy settings
PATCH /v1/users/me/settings/privacy
PATCH /v1/users/{username}/settings/privacy

# Update visibility settings
PATCH /v1/users/me/settings/visibility
PATCH /v1/users/{username}/settings/visibility
```

### User Sessions

```bash
# List sessions
GET /v1/users/me/sessions

# Revoke other sessions
POST /v1/users/me/sessions/revoke-others

# Revoke specific session
DELETE /v1/users/me/sessions/{id}

# Create user activity
POST /v1/users/me/activity
Content-Type: application/json

{
  "type": "deployment",
  "metadata": {}
}
```

---

## Functions

### Function Management

```bash
# List functions
GET /v1/functions

# Get function
GET /v1/functions/{id}

# Create function
POST /v1/functions
Content-Type: application/json

{
  "name": "my-function",
  "runtime": "python",
  "code": "..."
}

# Update function
PATCH /v1/functions/{id}
Content-Type: application/json

{
  "name": "...",
  "runtime": "...",
  "code": "..."
}

# Delete function
DELETE /v1/functions/{id}

# Deploy function
POST /v1/functions/{id}/deploy

# Get function stats
GET /v1/functions/{id}/stats
```

---

## API Keys

### API Key Management

```bash
# List API keys
GET /v1/api-keys

# Create API key
POST /v1/api-keys
Content-Type: application/json

{
  "name": "Production Key",
  "permissions": ["read", "execute"],
  "environments": ["production"]
}

# Get API key
GET /v1/api-keys/{id}

# Update API key
PATCH /v1/api-keys/{id}
Content-Type: application/json

{
  "name": "Updated Name"
}

# Delete API key
DELETE /v1/api-keys/{id}

# Rotate API key
POST /v1/api-keys/{id}/rotate

# List permissions
GET /v1/api-keys/{id}/permissions

# Add permission
POST /v1/api-keys/{id}/permissions
Content-Type: application/json

{
  "permission": "execute"
}

# Remove permission
DELETE /v1/api-keys/{id}/permissions/{perm_id}

# List environments
GET /v1/api-keys/{id}/environments

# Add environment
POST /v1/api-keys/{id}/environments
Content-Type: application/json

{
  "environment": "staging"
}

# Remove environment
DELETE /v1/api-keys/{id}/environments/{env_id}
```

---

## Registry

### Function Registry

```bash
# Search registry
GET /v1/registry/search?q=...

# Get function from registry
GET /v1/registry/{id}

# Publish function
POST /v1/registry/publish
Content-Type: application/json

{
  "name": "my-function",
  "description": "...",
  "code": "...",
  "runtime": "python",
  "version": "1.0.0"
}

# Get versions
GET /v1/registry/{id}/versions

# Get latest version
GET /v1/registry/{id}/latest
```

---

## Execution

### Execute Function

```bash
# Execute function (public)
POST /v1/execute/{functionId}
Content-Type: application/json

{
  "data": {}
}

# Execute function (authenticated)
POST /v1/run/{functionId}
Authorization: Bearer <token>
Content-Type: application/json

{
  "input": {}
}

# Get execution result
GET /v1/executions/{id}
```

---

## Follow System

### Follow Users

```bash
# Follow user
POST /v1/follow/users/{username}/follow

# Unfollow user
DELETE /v1/follow/users/{username}/follow

# Get user followers
GET /v1/follow/users/{username}/followers

# Get user following
GET /v1/follow/users/{username}/following

# Check following status
GET /v1/follow/users/{username}/status
```

### Follow Functions

```bash
# Follow function
POST /v1/follow/functions/{functionID}/follow

# Unfollow function
DELETE /v1/follow/functions/{functionID}/follow

# Get function followers
GET /v1/follow/functions/{functionID}/followers

# Check function following status
GET /v1/follow/functions/{functionID}/status
```

### My Follows

```bash
# Get my followed functions
GET /v1/follow/me/functions

# Get my follow stats
GET /v1/follow/me/stats
```

---

## Analytics

### Analytics Endpoints

```bash
# Get analytics overview
GET /v1/analytics/overview

# Get function analytics
GET /v1/analytics/functions/{id}

# Get user analytics
GET /v1/analytics/users/{id}
```

---

## Monitoring

### Health and Metrics

```bash
# Health check
GET /healthz

# Readiness check
GET /readyz

# Metrics (Prometheus format)
GET /metrics

# Custom metrics
GET /v1/monitoring/metrics
```

---

## WebSocket

### Real-time Updates

```bash
# Notifications WebSocket
WS /v1/ws/notifications

# Execution updates
WS /v1/ws/executions

# Status updates
WS /v1/ws/status
```

---

## Rate Limits

| Endpoint Category | Limit |
|-------------------|-------|
| Authentication | 5 requests/minute |
| Function Execution | 100 requests/minute |
| Registry Operations | 30 requests/minute |
| General API | 60 requests/minute |

---

## Error Responses

### Error Format

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Human readable message",
    "details": {}
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| RATE_LIMITED | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Server error |

---

## SDK Examples

### Python SDK

```python
from functionfly import FunctionClient

client = FunctionClient(api_key="your-api-key")

# Execute function
result = client.invoke("function-name", {"key": "value"})
```

### JavaScript SDK

```javascript
const { FunctionClient } = require('@functionfly/flypy');

const client = new FunctionClient({ apiKey: 'your-api-key' });
const result = await client.invoke('function-name', { key: 'value' });
```

---

## Postman Collection

Import the Postman collection for testing:

```bash
# Download collection
curl -o functionfly-api.json https://raw.githubusercontent.com/functionfly/functionfly/main/docs/postman-collection.json
```

---

## Support

- Discord: https://discord.gg/functionfly
- Email: support@functionfly.com
- GitHub Issues: https://github.com/functionfly/functionfly/issues
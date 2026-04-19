# FunctionFly (MVP1) API Specification

This is an MVP1-oriented REST API and routing surface.

## Conventions

- JSON over HTTPS.
- `Authorization: Bearer <jwt>` for dashboard.
- `X-App-Key: <key>` for app-scoped programmatic access.
- `X-Request-Id` propagated end-to-end.

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Human-readable description",
    "details": { ... },
    "requestId": "req_..."
  }
}
```

Common HTTP status codes:

| Status | Code | Description |
|--------|------|-------------|
| 400 | `invalid_request` | Missing or invalid parameters |
| 401 | `unauthorized` | Missing or invalid authentication |
| 403 | `forbidden` | Insufficient permissions |
| 404 | `not_found` | Resource does not exist |
| 409 | `conflict` | Resource already exists or state conflict |
| 422 | `validation_error` | Request body validation failed |
| 429 | `rate_limited` | Too many requests |
| 500 | `internal_error` | Server error (retry with backoff) |
| 503 | `service_unavailable` | Service temporarily unavailable |

## Pagination

List endpoints support standard pagination:

| Parameter | Type | Default | Max | Description |
|-----------|------|---------|-----|-------------|
| `limit` | integer | 20 | 100 | Number of items per page |
| `offset` | integer | 0 | - | Number of items to skip |

Paginated responses include:

```json
{
  "items": [ ... ],
  "pagination": {
    "total": 1000,
    "limit": 20,
    "offset": 0,
    "hasMore": true,
    "nextOffset": 20
  }
}
```

## Auth

### POST /v1/auth/login

MVP1 option: email+password.

**Request:**

```json
{
  "email": "user@example.com",
  "password": "securePassword123!",
  "rememberMe": true
}
```

**Success Response (200):**

```json
{
  "token": "<jwt>",
  "expiresIn": 3600,
  "user": {
    "id": "usr_...",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

**Error Responses:**
- `400 invalid_request` - Missing email or password
- `401 unauthorized` - Invalid credentials
- `403 forbidden` - Account suspended or email not verified
- `422 validation_error` - Password too weak (doesn't meet policy)
- `429 rate_limited` - Too many login attempts (exponential backoff required)

## Apps

### POST /v1/apps

Request:

```json
{ "name": "my-app", "slug": "my-app" }
```

Response:

```json
{ "id": "app_...", "name": "my-app", "slug": "my-app" }
```

### GET /v1/apps/{appId}

Returns app configuration including backend list.

## Backends

### POST /v1/apps/{appId}/backends

Registers a BYO backend endpoint.

Request:

```json
{
  "provider": "workers",
  "region": "iad",
  "url": "https://example.edge-target.dev",
  "sharedSecret": "<hmac secret>",
  "enabled": true
}
```

### GET /v1/apps/{appId}/backends

Lists configured backends and their current health summary.

**Query Parameters:**
- `limit` (integer, optional): Items per page (default: 20, max: 100)
- `offset` (integer, optional): Pagination offset (default: 0)

**Response:**

```json
{
  "items": [
    {
      "id": "b_...",
      "provider": "workers",
      "region": "iad",
      "url": "https://...",
      "health": { "status": "healthy", "latencyMs": 45 }
    }
  ],
  "pagination": {
    "total": 50,
    "limit": 20,
    "offset": 0,
    "hasMore": true,
    "nextOffset": 20
  }
}
```

**Error Responses:**
- `401 unauthorized` - Invalid or missing JWT
- `403 forbidden` - Not authorized for this app
- `404 not_found` - App does not exist
- `429 rate_limited` - Too many requests

## Routing

### GET /v1/apps/{appId}/route

Debug endpoint: returns the current best route decision for a given request context.

Query params (optional):

- `clientRegion`
- `method`

Response:

```json
{
  "primary": { "backendId": "b_...", "url": "https://..." },
  "failovers": [ { "backendId": "b_...", "url": "https://..." } ],
  "reason": "lowest_score",
  "requestId": "..."
}
```

## Data-plane routing surface

### /{appSlug}/*

Public routing endpoint served via Caddy.

Behavior:

- Caddy forwards to orchestrator for selection.
- Request proxied to selected backend.
- Idempotent retry to failover on timeout or 5xx.

## Health ingest (optional)

### POST /v1/ingest/health

For edge targets that can report their own local signals.


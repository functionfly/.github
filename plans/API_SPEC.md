# FunctionFly (MVP1) API Specification

This is an MVP1-oriented REST API and routing surface.

## Conventions

- JSON over HTTPS.
- `Authorization: Bearer <jwt>` for dashboard.
- `X-App-Key: <key>` for app-scoped programmatic access.
- `X-Request-Id` propagated end-to-end.

## Auth

### POST /v1/auth/login

MVP1 option: email+password.

Response:

```json
{ "token": "<jwt>", "expiresIn": 3600 }
```

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


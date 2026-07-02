---
title: AI Models API
description: Full API reference for model catalog, BYOK, and preferences
sidebar:
  order: 5
---


---

## Model Catalog

### List Models (Agent)

```
GET /v1/ai/models
```

Returns the curated model list for agents, with BYOK annotations.

### Get Full Catalog

```
GET /v1/ai/models/catalog
```

Query parameters:

| Param | Type | Description |
|-------|------|-------------|
| `provider` | string | Filter by provider |
| `tier` | string | Filter by tier (frontier, fast, reasoning, code, embedding, local) |

### Check Model Availability

```
POST /v1/ai/models/check
```

```json
{
  "model": "claude-sonnet-4-6",
  "provider": "anthropic"
}
```

### Refresh Catalog Cache

```
POST /v1/ai/models/catalog/refresh
```

---

## Preferences

### Get Preferences

```
GET /v1/ai/models/preferences
```

Returns the current tenant's AI preferences (profile, feature overrides,
allowed providers/models, user override setting).

### Update Preferences

```
PUT /v1/ai/models/preferences
```

```json
{
  "profile": "balanced",
  "feature_overrides": {
    "composer": "gpt-5-codex",
    "agent": "claude-sonnet-4-6"
  },
  "allowed_providers": ["openai", "anthropic"],
  "allowed_models": ["gpt-5.5", "claude-sonnet-4.6"],
  "allow_user_overrides": true
}
```

---

## BYOK Keys

### Connect Key

```
POST /v1/ai-keys/connect
```

```json
{
  "provider": "openai",
  "api_key": "sk-proj-..."
}
```

The key is validated against the provider before storage.

### List Connected Keys

```
GET /v1/ai-keys
```

```json
{
  "keys": [
    {
      "provider": "openai",
      "key_source": "byok",
      "connected_at": "2026-06-01T00:00:00Z",
      "last_used_at": "2026-06-30T12:00:00Z",
      "is_valid": true
    }
  ]
}
```

### List Supported Providers

```
GET /v1/ai-keys/providers
```

### Test Key

```
POST /v1/ai-keys/{provider}/test
```

### Rotate Key

```
POST /v1/ai-keys/{provider}/rotate
```

```json
{
  "api_key": "sk-proj-new-key..."
}
```

### Disconnect Key

```
DELETE /v1/ai-keys/{provider}
```

---

## AI Composer

### Generate Function

```
POST /v1/ai/composer/generate
```

```json
{
  "prompt": "Create a function that summarizes PDFs",
  "model": "claude-sonnet-4-6",
  "runtime": "python"
}
```

### Stream Generation

```
GET /v1/ai/composer/generate/stream
```

Server-Sent Events stream.

### Refine Function

```
POST /v1/ai/composer/refine
```

```json
{
  "code": "existing function code...",
  "instructions": "Add error handling for timeout",
  "model": "gpt-5-codex"
}
```

### Stream Refinement

```
GET /v1/ai/composer/refine/stream
```

---

## Health

### AI Service Health

```
GET /api/ai/health
```

### AI Feature Status

```
GET /api/ai/status
```

---

## Error Codes

| Status | Code | Meaning |
|--------|------|---------|
| 400 | `INVALID_PROVIDER` | Unknown provider |
| 400 | `INVALID_KEY_FORMAT` | Key format doesn't match provider |
| 400 | `KEY_VALIDATION_FAILED` | Key rejected by provider API |
| 404 | `KEY_NOT_FOUND` | No BYOK key for this provider |
| 429 | `AI_CALL_LIMIT_REACHED` | Plan AI call limit exceeded |
| 502 | `PROVIDER_ERROR` | Upstream provider returned an error |
| 503 | `MODEL_UNAVAILABLE` | Requested model is not available |

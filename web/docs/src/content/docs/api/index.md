---
title: REST API
description: FunctionFly REST API reference — authentication, functions, execution, and more
---

## Overview

The FunctionFly REST API is the primary interface for programmatic access to
the platform. All endpoints are versioned under `/v1/` and require
authentication via API key or session token.

**Base URL:** `https://api.functionfly.com/v1`

## Sections

- [Authentication](/api/authentication/) — API keys, session tokens, OAuth
- [Functions](/api/functions/) — CRUD for functions, versions, manifests
- [Execution](/api/execution/) — Invoke functions, streaming, async execution

## Full OpenAPI Reference

For the complete auto-generated API reference with every endpoint, schema,
and example, see the [OpenAPI Reference](/api-reference/).

## Authentication

All API requests require authentication. Use one of:

| Method | Header | Example |
|--------|--------|---------|
| API Key | `X-API-Key: <key>` | `X-API-Key: ffp_v1_abc...` |
| API Key (Auth header) | `Authorization: ApiKey <key>` | `Authorization: ApiKey ffp_v1_abc...` |
| Session JWT | `Authorization: Bearer <token>` | `Authorization: Bearer eyJhbG...` |

See [Authentication](/api/authentication/) for details.

## Rate Limits

| Plan | Requests/min | Requests/hour | Requests/day |
|------|-------------|---------------|--------------|
| Free | 60 | 3,600 | 86,400 |
| Starter | 120 | 7,200 | 172,800 |
| Professional | 300 | 18,000 | 432,000 |
| Enterprise | 1,000 | 60,000 | 1,440,000 |

Per-key rate limits can be configured independently (see [API Keys](/api-keys/)).

Rate limit headers are included in every response:

```
X-RateLimit-Limit: 300
X-RateLimit-Remaining: 287
X-RateLimit-Reset: 1719705660
```

## Error Format

All errors return JSON:

```json
{
  "error": "not_found",
  "message": "Function not found",
  "status": 404
}
```

## Pagination

List endpoints support pagination via `limit` and `offset`:

```bash
curl "https://api.functionfly.com/v1/functions?limit=20&offset=40" \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

## Next Steps

- [Authentication](/api/authentication/) — Auth methods and token exchange
- [Functions](/api/functions/) — Function CRUD
- [Execution](/api/execution/) — Invoking functions
- [API Keys](/api-keys/) — Managing API keys

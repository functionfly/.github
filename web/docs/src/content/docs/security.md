---
title: "SECURITY"
---

# Security Features

FunctionFly implements comprehensive security measures to protect against common web vulnerabilities and ensure secure API access.

## Overview

The security implementation includes:

- HMAC request signing for API authentication
- Rate limiting to prevent abuse
- Input validation and sanitization
- CORS and security headers

## HMAC Request Signing

FunctionFly supports HMAC-SHA256 request signing for enhanced API security. This is particularly useful for sensitive operations and API integrations.

### How It Works

Requests are signed using the formula:

```
HMAC_SHA256(sharedSecret, timestamp + method + path + bodyHash)
```

### Headers Required

- `X-FFLY-Timestamp`: Unix timestamp (must be within 5 minutes of server time)
- `X-FFLY-Signature`: Hex-encoded HMAC-SHA256 signature

### Configuration

Set the shared secret via environment variable:

```bash
API_SHARED_SECRET=your-api-shared-secret-here
```

### Routes Protected by HMAC

HMAC signing is required for:

- Deployment operations (`POST /v1/apps/{appId}/deploy`, `POST /deployments/{deploymentId}/rollback`)
- Admin write operations (tenant/user management)

### Example Usage

```javascript
const crypto = require('crypto');

function signRequest(method, path, body, sharedSecret) {
  const timestamp = Math.floor(Date.now() / 1000);
  const bodyHash = crypto.createHash('sha256').update(body || '').digest('hex');
  const signatureString = `${timestamp}${method}${path}${bodyHash}`;

  const signature = crypto.createHmac('sha256', sharedSecret)
    .update(signatureString)
    .digest('hex');

  return {
    'X-FFLY-Timestamp': timestamp.toString(),
    'X-FFLY-Signature': signature
  };
}

// Usage
const headers = signRequest('POST', '/v1/apps/123/deploy', '{"version":"1.0"}', 'your-secret');
```

## Rate Limiting

Rate limiting prevents abuse by limiting the number of requests per client IP within a time window.

### Configuration

- `RATE_LIMIT_REQUESTS`: Maximum requests per window (default: 100)
- `RATE_LIMIT_WINDOW_SECONDS`: Time window in seconds (default: 60)

### Rate Limit Headers

Successful requests include:

- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Unix timestamp when the limit resets

### Rate Limited Response

When limit is exceeded (HTTP 429):

```json
{
  "error": "Rate limit exceeded"
}
```

## Input Validation and Sanitization

All requests undergo comprehensive input validation and sanitization.

### Path Validation

- Prevents path traversal attacks (`../`)
- Rejects null bytes
- Limits path length (2048 characters)
- Blocks suspicious characters

### Query Parameter Validation

- Validates parameter names (alphanumeric, underscores, hyphens only)
- Limits parameter values (1000 characters max)
- HTML escapes values to prevent XSS

### Header Validation

- Validates header names (printable ASCII only)
- Rejects null bytes in header values
- Limits header value length (4096 characters)

## CORS and Security Headers

### CORS Configuration

Configure CORS via environment variables:

- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed origins (default: `*` in development; production must list real HTTPS origins, e.g. `https://admin.functionfly.com` for the admin SPA, `https://auth.functionfly.com` for the auth service)
- `CORS_ALLOWED_METHODS`: Allowed HTTP methods (default: `GET, POST, PUT, PATCH, DELETE, OPTIONS`)
- `CORS_ALLOWED_HEADERS`: Allowed headers

### Security Headers

Always included:

- `X-Frame-Options: DENY` - Prevents clickjacking
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing
- `X-XSS-Protection: 1; mode=block` - Enables XSS protection
- `Referrer-Policy: strict-origin-when-cross-origin` - Controls referrer information

### Content Security Policy

Configure via `CONTENT_SECURITY_POLICY` environment variable:

```
default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self';
```

### HTTP Strict Transport Security (HSTS)

Enabled for HTTPS connections:

- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- Configure max-age via `HSTS_MAX_AGE` (default: 31536000 seconds / 1 year)

## Environment Variables

All security features are configurable via environment variables:

```bash
# HMAC Signing
API_SHARED_SECRET=your-api-shared-secret-here

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60

# CORS
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET, POST, PUT, PATCH, DELETE, OPTIONS
CORS_ALLOWED_HEADERS=Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-FFLY-Timestamp, X-FFLY-Signature

# Security Headers
CONTENT_SECURITY_POLICY=default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self';
HSTS_MAX_AGE=31536000
```

## Security Best Practices

1. **Always use HTTPS** in production
2. **Rotate API shared secrets** regularly
3. **Monitor rate limiting logs** for abuse patterns
4. **Keep dependencies updated** for security patches
5. **Use strong, unique secrets** for HMAC signing
6. **Configure CORS** appropriately for your use case
7. **Review security headers** for your specific requirements

## Monitoring and Logging

Security events are logged with structured logging:

- HMAC verification failures
- Rate limit violations
- Input validation errors
- Invalid headers/paths

Monitor these logs to detect and respond to security threats.

## Access Control & Seat Management

Account sharing and seat enforcement are governed by per-plan user limits. See [`docs/ACCOUNT_SHARING.md`](ACCOUNT_SHARING.md) for the full operational reference and [`plans/ACCOUNT_SHARING.md`](../plans/ACCOUNT_SHARING.md) for the design spec.

## Information Leakage Prevention (Tenant Mismatch)

State Fabric repository methods use a deliberate error-masking pattern when a resource exists but belongs to a different tenant. Instead of returning "unauthorized" or "forbidden", the API returns "state fabric not found" (HTTP 404).

**Rationale:** Returning "not found" instead of "unauthorized" prevents information leakage about whether a resource ID exists in the system. An attacker cannot enumerate valid resource IDs by observing error responses.

**Affected operations** (in `internal/storage/statefabric/repository.go`):
- `GetFabric` — tenant ID check at line 346
- `UpdateFabric` — tenant ID check at line 360
- `DeleteFabric` — tenant ID check at line 391
- `GetMetrics` — tenant ID check at line 433
- `GetPipeline` — tenant ID check at line 444
- `UpdatePipeline` — tenant ID check at line 678
- `DeletePipeline` — tenant ID check at line 755
- `SetFabricSuspended` — tenant ID check at line 787
- `GetSettings` — tenant ID check at line 851
- `UpdateSettings` — tenant ID check at line 869
- `GetAuditLog` — tenant ID check at line 897

This is intentional and should not be changed without careful security review.

## Before Making the Repo Public

- **No real credentials in docs or examples**: Staging/production DB passwords, Neon project IDs, and connection strings must use placeholders (e.g. `<from Neon Console>`, `ep-staging-xxxxx.us-east-1.aws.neon.tech`). Real values belong in a secrets manager or private runbooks only.
- **Env files**: Only `*.env.example` (and similar) are committed. `.env`, `.env.production`, `.env.staging`, and `deploy/database/production.env` are in `.gitignore` and must never be committed.
- **Supabase / third-party**: `.env.example` uses placeholders for `VITE_SUPABASE_URL` and `VITE_SUPABASE_ANON_KEY`; do not commit real project URLs or keys.
- **Test accounts**: Docs may reference a dev-only account (e.g. `admin@functionfly.local`); ensure production deployments use strong passwords and that this is clearly documented as dev-only.
- **Edge / deploy**: `deploy/edge/certs-in/` and `deploy/edge/certs-out/` are gitignored; no TLS private keys or certs in the repo.

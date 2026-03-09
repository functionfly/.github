# Versioning System Documentation

This document provides comprehensive documentation for the FunctionFly versioning system, covering API versioning, function versioning, rollback procedures, deprecation policies, and edge provider compatibility.

---

## Table of Contents

1. [Overview](#overview)
2. [API Version Management](#api-version-management)
3. [Function Version Lifecycle](#function-version-lifecycle)
4. [Rollback Procedures](#rollback-procedures)
5. [Deprecation Policy](#deprecation-policy)
6. [Edge Provider Compatibility](#edge-provider-compatibility)
7. [API Reference](#api-reference)
8. [Examples](#examples)

---

## Overview

The FunctionFly versioning system provides a comprehensive solution for managing versions at two levels:

1. **API Versioning**: Platform-level API version management
2. **Function Versioning**: Per-function version management with rollback capabilities

### Key Features

- Semantic versioning support (semver)
- Version aliases (latest, stable, draft)
- Deprecation with sunset dates
- Rollback with multiple strategies
- Deployment tracking across providers
- Service contract versioning

---

## API Version Management

### Version States

| State | Description |
|-------|-------------|
| `active` | Current, supported version |
| `deprecated` | Still functional but not recommended |
| `sunset` | No longer available, returns 410 Gone |
| `archived` | Historical version, read-only |

### Version Format

API versions use a simple `v{N}` format:
- `v1` - First stable API version
- `v2` - Second API version
- `v3` - Third API version

### Request Version Detection

The middleware detects the requested version from (in priority order):

1. **URL Path**: `/v1/functions`, `/v2/functions`
2. **Accept-Version Header**: `Accept-Version: v1`
3. **Default Version**: `v1`

### Deprecation Headers

When using a deprecated API version, the following headers are added to responses:

```
Deprecation: true
X-API-Warning: This API version is deprecated
Sunset: Sat, 01 Jan 2025 00:00:00 GMT
Link: </v2/>; rel="successor-version"
X-API-Sunset-Message: Migration guide: https://docs.example.com/migration
```

---

## Function Version Lifecycle

### Version States

| State | Description |
|-------|-------------|
| `draft` | Work in progress, not deployed |
| `published` | Deployed and available |
| `deprecated` | Still functional, migration recommended |
| `archived` | No longer available |

### Version Aliases

FunctionFly supports the following version aliases:

- `latest` - Most recently published version
- `stable` - Latest non-prerelease version
- `draft` - Unpublished work

### State Transitions

```
draft → published → deprecated → archived
         ↘________________↗
```

### Version Naming

Function versions follow semantic versioning:
- `v1.0.0` - Initial release
- `v1.1.0` - Minor update (backward compatible)
- `v2.0.0` - Major version (breaking changes)
- `v1.0.0-beta` - Prerelease
- `v1.0.0-rc.1` - Release candidate

---

## Rollback Procedures

### Rollback Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `immediate` | Instant switch to target version | Critical hotfix |
| `gradual` | Percentage-based traffic shift | A/B testing rollback |
| `canary` | Route small subset first | Conservative rollback |

### Rollback API

```bash
# Rollback to previous version
curl -X POST https://api.functionfly.com/v1/functions/{functionId}/rollback \
  -H "Authorization: Bearer {token}" \
  -d '{"toVersion": "v1.0.0", "strategy": "immediate"}'
```

### Rollback Response

```json
{
  "rollbackId": "rb_abc123",
  "functionId": "fn_xyz789",
  "fromVersion": "v2.0.0",
  "toVersion": "v1.0.0",
  "strategy": "immediate",
  "status": "completed",
  "initiatedAt": "2024-01-15T10:30:00Z",
  "completedAt": "2024-01-15T10:30:05Z"
}
```

---

## Deprecation Policy

### Deprecation Timeline

1. **Announcement**: Version marked as deprecated
2. **Grace Period**: 90 days default (configurable)
3. **Sunset Date**: Version returns 410 Gone
4. **Archive**: Version marked as archived (read-only)

### Deprecation Request

```bash
curl -X POST https://api.functionfly.com/v1/functions/{functionId}/versions/{version}/deprecate \
  -H "Authorization: Bearer {token}" \
  -d '{
    "reason": "Security vulnerability in authentication",
    "replacedBy": "v2.0.0",
    "migrationGuide": "https://docs.example.com/migration",
    "gracePeriodDays": 30
  }'
```

### Migration Headers

When calling deprecated endpoints:

```http
HTTP/1.1 200 OK
Deprecation: true
X-API-Warning: This API version is deprecated
Sunset: Sat, 15 Feb 2025 00:00:00 GMT
Link: </v2/>; rel="successor-version"
```

---

## Edge Provider Compatibility

### Cloudflare Workers

FunctionFly versions are deployed as Cloudflare Workers with the following version handling:

- **Version in Worker Name**: `function-v1.0.0`
- **Routing**: Workers routing based on version header
- **KV Storage**: Version metadata stored in KV namespace

```javascript
// Worker route configuration
addEventListener('fetch', event => {
  const version = event.request.headers.get('x-function-version') || 'latest';
  // Route to appropriate worker
});
```

### Vercel

Vercel deployments use function aliases for versioning:

- **Version as Alias**: `my-function-v1.0.0`
- **Routing**: Vercel Routes for version-based routing
- **Environment**: Version-specific environment variables

```json
// vercel.json
{
  "routes": [
    { "src": "/v1/(.*)", "dest": "/api/v1/$1" },
    { "src": "/v2/(.*)", "dest": "/api/v2/$1" }
  ]
}
```

### Fly.io

Fly.io uses region-specific deployments with version tracking:

- **Version Labels**: Fly volume and deployment labels
- **Routing**: Fly proxy with version-based rules
- **Regions**: Version-specific region assignments

```toml
# fly.toml
[deploy]
  strategy = " Rolling"

[metadata]
  function-version = "v1.0.0"
```

### Deno Deploy

Deno Deploy uses import maps for version management:

- **Version in Import Map**: Specified in deno.json
- **Routing**: Deno Deploy routes by path prefix
- **Cache**: Version-specific KV cache

```json
// deno.json
{
  "imports": {
    "my-function": "https://deno.example.com/functions/my-function@v1.0.0/mod.ts"
  }
}
```

---

## API Reference

### API Versions

#### List API Versions

```bash
GET /v1/api-versions
```

#### Create API Version

```bash
POST /v1/api-versions
{
  "version": "v3",
  "pathPrefix": "/v3",
  "status": "active",
  "releasedAt": "2024-01-15T00:00:00Z"
}
```

#### Deprecate API Version

```bash
POST /v1/api-versions/{version}/deprecate
{
  "sunsetAt": "2025-01-15T00:00:00Z",
  "sunsetMessage": "Use v3 API"
}
```

### Function Versions

#### List Function Versions

```bash
GET /v1/functions/{functionId}/versions
```

#### Publish Version

```bash
POST /v1/functions/{functionId}/versions/{version}/publish
{
  "setAsLatest": true,
  "setAsStable": true
}
```

#### Deprecate Version

```bash
POST /v1/functions/{functionId}/versions/{version}/deprecate
{
  "reason": "Use v2.0.0 instead",
  "replacedBy": "v2.0.0",
  "migrationGuide": "https://docs.example.com/migration",
  "gracePeriodDays": 30
}
```

#### Rollback Version

```bash
POST /v1/functions/{functionId}/rollback
{
  "toVersion": "v1.0.0",
  "strategy": "immediate"
}
```

### Aliases

#### Set Version Alias

```bash
POST /v1/functions/{functionId}/aliases/latest
{
  "version": "v2.0.0"
}
```

#### Get Version Alias

```bash
GET /v1/functions/{functionId}/aliases/stable
```

### Changelog

#### Create Changelog Entry

```bash
POST /v1/functions/{functionId}/changelog
{
  "version": "v2.0.0",
  "changeType": "major",
  "changeCategory": "api",
  "description": "Complete API redesign",
  "breakingChanges": ["Removed legacy endpoints"],
  "migrationSteps": ["Update client library"]
}
```

### Contracts

#### List Service Contracts

```bash
GET /v1/contracts
```

#### Create Service Contract

```bash
POST /v1/contracts
{
  "serviceName": "user-service",
  "contractVersion": "1.0.0",
  "contractType": "rest",
  "schema": {...}
}
```

---

## Examples

### Publishing a New Version

```bash
# 1. Create a new version (draft)
curl -X POST https://api.functionfly.com/v1/functions/{functionId}/versions \
  -H "Authorization: Bearer {token}" \
  -d '{"version": "v1.1.0", "runtime": "nodejs18"}'

# 2. Add changelog entry
curl -X POST https://api.functionfly.com/v1/functions/{functionId}/changelog \
  -H "Authorization: Bearer {token}" \
  -d '{
    "version": "v1.1.0",
    "changeType": "minor",
    "changeCategory": "feature",
    "description": "Added support for async operations"
  }'

# 3. Publish the version
curl -X POST https://api.functionfly.com/v1/functions/{functionId}/versions/v1.1.0/publish \
  -H "Authorization: Bearer {token}" \
  -d '{"setAsLatest": true}'
```

### Rolling Back a Problematic Release

```bash
# Immediate rollback
curl -X POST https://api.functionfly.com/v1/functions/{functionId}/rollback \
  -H "Authorization: Bearer {token}" \
  -d '{
    "toVersion": "v1.0.0",
    "strategy": "immediate"
  }'
```

### Deprecating an Old Version

```bash
curl -X POST https://api.functionfly.com/v1/functions/{functionId}/versions/v1.0.0/deprecate \
  -H "Authorization: Bearer {token}" \
  -d '{
    "reason": "Critical security vulnerability",
    "replacedBy": "v2.0.0",
    "migrationGuide": "https://docs.example.com/security-fix",
    "gracePeriodDays": 7
  }'
```

### Using Version Aliases

```bash
# Always get the latest version
curl -X GET https://api.functionfly.com/v1/functions/{functionId} \
  -H "X-Function-Version: latest"

# Get the stable version
curl -X GET https://api.functionfly.com/v1/functions/{functionId} \
  -H "X-Function-Version: stable"
```

---

## Error Codes

| Code | Description |
|------|-------------|
| `VERSION_NOT_FOUND` | Requested version does not exist |
| `VERSION_CONFLICT` | Version already exists |
| `VERSION_DEPRECATED` | Version is deprecated |
| `VERSION_SUNSET` | Version has been sunset |
| `ROLLBACK_FAILED` | Rollback operation failed |
| `INVALID_VERSION_STATE` | Invalid state transition |

---

## Related Documentation

- [Architecture Overview](../plans/ARCHITECTURE.md)
- [API Specification](../plans/API_SPEC.md)
- [Migration Guide](./MIGRATION_GUIDE.md)

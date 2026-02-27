# API Migration Guide: v1 to v2

## Overview

This guide helps you migrate from the FunctionFly Registry API v1 to v2. The v2 API includes improvements to field naming conventions, new features, and better performance.

### Timeline

| Milestone | Date |
|-----------|------|
| v2 API Released | January 1, 2026 |
| v1 Deprecation Notice | January 1, 2026 |
| v1 Sunset Date | January 1, 2027 |

## What's Changed

### Field Naming Convention

The most significant change is the adoption of **camelCase** for field names in API responses. v1 used **snake_case**.

#### v1 Response (snake_case)

```json
{
  "popularity_score": 42,
  "deterministic_score": 0.95,
  "latest_version": "1.0.0",
  "published_at": "2025-12-01T00:00:00Z"
}
```

#### v2 Response (camelCase)

```json
{
  "popularityScore": 42.0,
  "deterministicScore": 0.95,
  "latestVersion": "1.0.0",
  "publishedAt": "2025-12-01T00:00:00Z"
}
```

### Type Changes

| Field | v1 Type | v2 Type | Migration |
|-------|---------|---------|-----------|
| popularity_score | integer | float | Cast to float: `parseFloat()` |
| trust_score | N/A | float | New field, default 0.0 |
| execution_count | N/A | integer | New field, default 0 |
| verified | N/A | boolean | New field, default false |

## Endpoint Changes

### GET /v1/registry/functions → /v2/registry/functions

**Breaking Changes:**

- Field `popularity_score` renamed to `popularityScore` (type changed to float)
- Field `deterministic_score` renamed to `deterministicScore`

**New Fields:**

- `trust_score` (float): New trust metric (0.0-1.0)
- `execution_count` (integer): Number of times the function was executed
- `last_executed_at` (timestamp): ISO 8601 timestamp of last execution

**Migration Example:**

```javascript
// v1
const score = data.popularity_score;

// v2
const score = parseFloat(data.popularityScore);
```

### GET /v1/registry/functions/{author}/{name} → /v2/registry/functions/{author}/{name}

**Breaking Changes:**

- Field `latest_version` renamed to `latestVersion`

**New Fields:**

- `trust_score` (float): Function trust score
- `verified` (boolean): Whether the function is verified
- `signature_info` (object): Cryptographic signature information

### GET /v1/registry/search → /v2/registry/search

**New Fields:**

- `relevance_score` (float): Search relevance score
- `highlights` (array): Matching text highlights

## Migration Steps

### Step 1: Update Endpoint URLs

Replace `/v1/` with `/v2/` in all your API calls:

```javascript
// Before (v1)
const response = await fetch('https://api.functionfly.dev/v1/registry/functions');

// After (v2)
const response = await fetch('https://api.functionfly.dev/v2/registry/functions');
```

### Step 2: Update Field References

Update your code to use camelCase field names:

```javascript
// Before (v1)
const popularity = func.popularity_score;
const version = func.latest_version;

// After (v2)
const popularity = parseFloat(func.popularityScore);
const version = func.latestVersion;
```

### Step 3: Handle Type Changes

The `popularityScore` field is now a float:

```javascript
// Handle float type
const score = parseFloat(data.popularityScore) || 0;

// Or use Number()
const score = Number(data.popularityScore) || 0;
```

### Step 4: Handle New Fields

Add handling for new v2 fields:

```javascript
// Trust score (new in v2)
const trustScore = data.trustScore ?? 0.0;

// Verified status (new in v2)
const isVerified = data.verified ?? false;

// Execution count (new in v2)
const execCount = data.executionCount ?? 0;
```

## Deprecation Headers

When calling v1 endpoints, you'll receive deprecation warnings:

```http
Deprecation: true
Sunset: Sat, 01 Jan 2027 00:00:00 GMT
Link: <https://api.functionfly.dev/v2/registry/functions>; rel="successor-version"
X-API-Warning: This API version is deprecated. Please migrate to v2.
```

## JavaScript/TypeScript Migration Example

```typescript
// v1 client
interface FunctionV1 {
  popularity_score: number;
  deterministic_score: number;
  latest_version: string;
}

// v2 client
interface FunctionV2 {
  popularityScore: number;
  deterministicScore: number;
  latestVersion: string;
  trustScore: number;
  verified: boolean;
  executionCount: number;
  lastExecutedAt: string | null;
}

// Migration helper
function migrateFunctionV1ToV2(v1: FunctionV1): FunctionV2 {
  return {
    popularityScore: Number(v1.popularity_score),
    deterministicScore: v1.deterministic_score,
    latestVersion: v1.latest_version,
    trustScore: 0.0,
    verified: false,
    executionCount: 0,
    lastExecutedAt: null
  };
}
```

## Python Migration Example

```python
# v1
def get_popularity(func):
    return func['popularity_score']

# v2
def get_popularity(func):
    return float(func['popularityScore'])
```

## Frequently Asked Questions

### Q: Will v1 stop working immediately?

**A:** No, v1 will continue to work until the sunset date (January 1, 2027), but we recommend migrating as soon as possible.

### Q: Do I need to update my code right away?

**A:** No immediate action is required, but you should plan your migration to v2 before the sunset date to avoid service interruptions.

### Q: How do I handle the popularity score type change?

**A:** Use `parseFloat()` in JavaScript or `float()` in Python to convert the string/number to a float value.

### Q: What happens if I don't migrate?

**A:** After the sunset date (January 1, 2027), v1 endpoints will return 410 Gone status.

### Q: Can I use both v1 and v2 simultaneously?

**A:** Yes, you can gradually migrate your integrations. Many customers run both versions during transition.

## Support

- **Documentation**: <https://docs.functionfly.com>
- **API Reference**: <https://docs.functionfly.com/api-reference>
- **Support Email**: <support@functionfly.com>
- **Community Slack**: <https://functionfly.slack.com>

## Changelog

| Date | Change |
|------|--------|
| 2026-01-01 | v2 API released |
| 2026-01-01 | v1 deprecated with warning headers |
| 2027-01-01 | v1 sunset - endpoints return 410 Gone |

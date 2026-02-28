# Production-Ready URLs Plan

## Overview

This document outlines recommendations for making URLs across the FunctionFly application more robust, type-safe, and production-ready using npm libraries.

## Current State Analysis

### Frontend (web/dashboard)
- **Routing**: Uses `react-router-dom` v7
- **URL Handling**: Basic route constants in [`constants.ts`](web/dashboard/src/lib/constants.ts:4)
- **Dynamic Parameters**: `:author`, `:name`, `:id`, `:username`, `:slug`, etc.
- **No URL validation** on dynamic parameters

### Backend (Go)
- **Routing**: Uses `gorilla/mux`
- **API Versioning**: `/v1`, `/v2` prefixes
- **Structured Routes**: Functions, registry, auth, monitoring, etc.
- **URL Utilities Available**:
  - [`slugify`](publish_slugify.json) - Convert text to URL-friendly slugs
  - [`url-encode`](publish_url_encode.json) - URL percent-encoding
  - [`url-decode`](publish_url_decode.json) - URL percent-decoding

---

## NPM Library Recommendations

### 1. URL Parsing & Building

| Library | Purpose | Bundle Size |
|---------|---------|-------------|
| [`query-string`](https://www.npmjs.com/package/query-string) | Parse/stringify URL query strings | ~2KB |
| [`url`](https://nodejs.org/api/url.html) | Node.js URL polyfill (built-in) | N/A |
| [`regexparam`](https://www.npmjs.com/package/regexparam) | Convert path patterns to regex | ~500B |
| [`path-to-regexp`](https://www.npmjs.com/package/path-to-regexp) | Convert path strings to regex | ~6KB |

**Recommendation**: Add `query-string` for robust query parameter handling.

```bash
cd web/dashboard && npm install query-string
```

### 2. Type-Safe Routing

| Library | Purpose |
|---------|---------|
| [`@tanstack/router`](https://www.npmjs.com/package/@tanstack/router) | Type-safe routing for React |
| [`type-route`](https://www.npmjs.com/package/type-route) | Type-safe routing |

**Recommendation**: Consider migrating to `@tanstack/router` for full type safety (major refactor).

### 3. URL Validation & Sanitization

| Library | Purpose |
|---------|---------|
| [`validator`](https://www.npmjs.com/package/validator) | URL validation & sanitization |
| [`slugify`](https://www.npmjs.com/package/slugify) | String to slug conversion |

**Recommendation**: Add `validator` for URL validation on user inputs.

```bash
cd web/dashboard && npm install validator slugify
```

### 4. Current Dependencies Already Useful

These packages in [`package.json`](web/dashboard/package.json) already help with URLs:
- `react-router-dom` - Already handles routing
- `zod` - For URL parameter validation schemas
- `axios` - Handles URL encoding in requests

---

## Implementation Plan

### Phase 1: URL Utilities Library

Create a new URL utilities module at [`web/dashboard/src/lib/url-utils.ts`](web/dashboard/src/lib/url-utils.ts):

```typescript
// Recommended: web/dashboard/src/lib/url-utils.ts
import queryString from 'query-string';
import slugify from 'slugify';
import validator from 'validator';

// Query parameter utilities
export function parseQueryParams(search: string): Record<string, string> {
  return queryString.parse(search) as Record<string, string>;
}

export function buildQueryString(params: Record<string, unknown>): string {
  return queryString.stringify(params);
}

// Slug utilities
export function createSlug(text: string, options?: { maxLength?: number; separator?: string }): string {
  const separator = options?.separator || '-';
  const maxLength = options?.maxLength || 100;
  
  return slugify(text, {
    lower: true,
    strict: true,
    separator,
    trim: true,
  }).slice(0, maxLength);
}

// URL validation
export function isValidUrl(url: string): boolean {
  return validator.isURL(url, {
    require_protocol: true,
    require_valid_protocol: true,
    protocols: ['http', 'https'],
  });
}

export function sanitizeUrl(url: string): string {
  return validator.trim(validator.escape(url));
}

// Path builder with type safety
export function buildPath(base: string, ...parts: string[]): string {
  const cleanParts = parts.filter(Boolean).map(p => p.replace(/^\/|\/$/g, ''));
  return `/${[base, ...cleanParts].filter(Boolean).join('/')}`;
}
```

### Phase 2: Route Parameter Validation

Add validation utilities for route parameters using Zod:

```typescript
// Recommended: web/dashboard/src/lib/route-validators.ts
import { z } from 'zod';

// Common route parameter schemas
export const authorSchema = z.string()
  .min(1)
  .max(50)
  .regex(/^[a-zA-Z0-9_-]+$/, 'Invalid author name');

export const functionNameSchema = z.string()
  .min(1)
  .max(100)
  .regex(/^[a-zA-Z0-9_-]+$/, 'Invalid function name');

export const uuidSchema = z.string().uuid();

export const slugSchema = z.string()
  .min(1)
  .max(200)
  .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, 'Invalid slug');

// Route parameter extraction with validation
export function validateRouteParams<T>(params: unknown, schema: z.ZodSchema<T>): T {
  return schema.parse(params);
}
```

### Phase 3: Enhanced Route Constants

Update [`constants.ts`](web/dashboard/src/lib/constants.ts) with builder functions:

```typescript
// Enhanced ROUTES object with builder functions
export const ROUTES = {
  // ... existing routes ...
  
  // Builder functions for dynamic routes
  function: (author: string, name: string) => `/fx/${author}/${name}`,
  playground: (author: string, name: string) => `/run/${author}/${name}`,
  replay: (execId: string) => `/replay/${execId}`,
  userProfile: (username: string) => `/u/${username}`,
  blogPost: (slug: string) => `/blog/${slug}`,
  docs: (slug?: string) => slug ? `/docs/${slug}` : '/docs',
  
  // Query string builders
  registrySearch: (query: string, page = 1) => 
    `/registry?q=${encodeURIComponent(query)}&page=${page}`,
} as const;
```

### Phase 4: API URL Builder

Create an API URL builder for consistent backend communication:

```typescript
// Recommended: web/dashboard/src/lib/api-urls.ts
import { API_BASE_URL } from './constants';

const API_VERSION = 'v1';

export const API_URLS = {
  // Auth
  login: `${API_BASE_URL}/v1/auth/login`,
  signup: `${API_BASE_URL}/v1/auth/signup`,
  logout: `${API_BASE_URL}/v1/auth/logout`,
  
  // Functions
  listFunctions: (page = 1, limit = 20) => 
    `${API_BASE_URL}/v1/registry/functions?page=${page}&limit=${limit}`,
  getFunction: (author: string, name: string) => 
    `${API_BASE_URL}/v1/registry/functions/${author}/${name}`,
  searchFunctions: (query: string) => 
    `${API_BASE_URL}/v1/registry/search?q=${encodeURIComponent(query)}`,
  
  // Execution
  execute: (author: string, name: string) => 
    `${API_BASE_URL}/v1/fx/${author}/${name}`,
  executeWithVersion: (author: string, name: string, version: string) => 
    `${API_BASE_URL}/v1/fx/${author}/${name}@${version}`,
  
  // Replay
  replay: (execId: string) => 
    `${API_BASE_URL}/v1/replay/${execId}`,
} as const;
```

---

## Backend Improvements (Go)

### URL Structure Recommendations

Current patterns are good. Consider these additions:

1. **Versioned API with content negotiation**:
   ```
   Accept: application/vnd.functionfly.v1+json
   ```

2. **Rate-limited function URLs**:
   ```
   /v1/fx/{author}/{name}@.{version} (semver)
   ```

3. **Function aliases**:
   ```
   /v1/fx/{alias} -> /v1/fx/{author}/{name}
   ```

### Middleware for URL Validation

Add validation in [`internal/api/middleware/`](internal/api/middleware):

```go
// Validate URL path parameters
func ValidatePathParams(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        
        // Validate author name
        if author, ok := vars["author"]; ok {
            if !isValidAuthorName(author) {
                http.Error(w, "Invalid author name", http.StatusBadRequest)
                return
            }
        }
        
        // Validate function name
        if name, ok := vars["name"]; ok {
            if !isValidFunctionName(name) {
                http.Error(w, "Invalid function name", http.StatusBadRequest)
                return
            }
        }
        
        next.ServeHTTP(w, r)
    })
}

func isValidAuthorName(s string) bool {
    return regexp.MustCompile(`^[a-zA-Z0-9_-]{1,50}$`).MatchString(s)
}

func isValidFunctionName(s string) bool {
    return regexp.MustCompile(`^[a-zA-Z0-9_-]{1,100}$`).MatchString(s)
}
```

---

## Summary of Recommended npm Packages

| Package | Purpose | Priority |
|---------|---------|----------|
| `query-string` | Query param parsing/stringifying | High |
| `slugify` | String to URL slug | High |
| `validator` | URL validation & sanitization | High |
| `path-to-regexp` | Path pattern matching | Medium |
| `@tanstack/router` | Type-safe routing | Low (major refactor) |

---

## Installation Commands

```bash
# Install recommended packages in dashboard
cd web/dashboard
npm install query-string slugify validator path-to-regexp

# Install in site (if needed)
cd ../site
npm install query-string slugify
```

---

## Files to Create/Modify

1. **Create**: `web/dashboard/src/lib/url-utils.ts` - URL utilities
2. **Create**: `web/dashboard/src/lib/route-validators.ts` - Route parameter validators
3. **Modify**: `web/dashboard/src/lib/constants.ts` - Add route builder functions
4. **Create**: `web/dashboard/src/lib/api-urls.ts` - API URL constants
5. **Modify**: `web/dashboard/src/App.tsx` - Add parameter validation
6. **Create**: `internal/api/middleware/validation.go` - Backend validation (Go)

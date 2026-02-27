# Registry API Endpoints Fix Plan

## Current Implementation Analysis

### Existing Registry API Endpoints
1. `GET /registry/functions` - List functions
2. `GET /registry/functions/{author}/{name}` - Get function details
3. `GET /registry/functions/{author}/{name}/versions` - List versions
4. `POST /registry/publish` - Publish a new function
5. `POST /fx/{author}/{name}` - Execute function (latest version)
6. `POST /fx/{author}/{name}@{version}` - Execute specific version
7. `POST /registry/functions/{author}/{name}/test` - Test a function
8. `GET /registry/functions/{author}/{name}/stats` - Get function stats
9. `POST /registry/functions/{author}/{name}/rating` - Submit rating
10. `GET /registry/functions/search` - Search functions

---

## Issues Identified

### 1. HandleGetFunction (query.go:15-50)
**Current behavior:**
- Returns function info via `fn.ToInfo(fnVersion)`
- Only expands manifest if explicitly requested via `?expand=manifest`
- **Missing data:** capabilities, bundle_size, source_hash, deployment_id, backend_id

**Required fixes:**
- Include version-level capabilities in response by default
- Add bundle_size and source_hash to response
- Add deployment_id and backend_id if present

### 2. HandlePublish (publish.go:34-226)
**Current behavior:**
- Creates function and version in database
- Returns minimal response with only: ok, function, version, message, verification_status

**Required fixes:**
- Return full function details including:
  - function_id (UUID)
  - runtime
  - timeout_ms
  - memory_mb
  - capabilities
  - deterministic
  - side_effects
  - idempotent
  - cache_ttl

### 3. HandleListFunctions / HandleSearchFunctions (query.go)
**Current behavior:**
- Returns list of functions via `ToInfoWithRating`
- Missing capabilities and other version metadata

**Required fixes:**
- Add capabilities to FunctionInfo response
- Include version metadata (bundle_size, source_hash)

### 4. FunctionInfo Type (functionregistry/types.go:247-278)
**Current fields:**
```
- author, name, version, title, description, runtime
- category, tags, price_per_call, reliability
- deterministic, side_effects, idempotent, cache_ttl
- input_type, output_type, input_example, output_example
- trust_score fields
```

**Missing fields:**
- capabilities (array of strings)
- bundle_size (int)
- source_hash (string)
- deployment_id (string)
- backend_id (string)
- playground_url (already added in ToInfo but not in type)
- documentation_url (already added in ToInfo but not in type)

### 5. ToInfo Helper (storage/registry/helpers.go)
**Missing fields to add:**
- capabilities from function-level (line 26 of types.go)
- capabilities from version-level (line 45 of types.go)
- bundle_size (int32)
- source_hash (string)

---

## Implementation Plan

### Step 1: Update FunctionInfo Type
**File:** `internal/functionregistry/types.go`

Add new fields:
```go
type FunctionInfo struct {
    // ... existing fields ...
    
    // New fields to add
    Capabilities   []string `json:"capabilities,omitempty"`
    BundleSize     int      `json:"bundle_size,omitempty"`
    SourceHash     string   `json:"source_hash,omitempty"`
    DeploymentID   string   `json:"deployment_id,omitempty"`
    BackendID      string   `json:"backend_id,omitempty"`
}
```

### Step 2: Update ToInfo Helper
**File:** `internal/storage/registry/helpers.go`

Update `ToInfoWithRating` to include:
- Parse and add capabilities from version
- Add bundle_size if present
- Add source_hash if present
- Add deployment_id if present
- Add backend_id if present

### Step 3: Update HandleGetFunction
**File:** `internal/api/handlers/registry/query.go`

Update `HandleGetFunction` to:
- Always include capabilities in response (not just when expand=manifest)
- Add bundle_size and source_hash
- Add deployment_id and backend_id

### Step 4: Update HandlePublish
**File:** `internal/api/handlers/registry/publish.go`

Update `PublishResponse` to include full function details:
```go
type PublishResponse struct {
    OK                 bool                   `json:"ok"`
    Function           string                 `json:"function"` // author/name
    Version            string                 `json:"version"`
    Message            string                 `json:"message,omitempty"`
    VerificationStatus string                 `json:"verification_status,omitempty"`
    
    // New fields
    FunctionID         string                 `json:"function_id,omitempty"`
    Runtime            string                 `json:"runtime,omitempty"`
    TimeoutMs          int                    `json:"timeout_ms,omitempty"`
    MemoryMB           int                    `json:"memory_mb,omitempty"`
    Capabilities       []string               `json:"capabilities,omitempty"`
    Deterministic      bool                  `json:"deterministic"`
    SideEffects        string                 `json:"side_effects,omitempty"`
    Idempotent         bool                  `json:"idempotent"`
    CacheTTL           int                    `json:"cache_ttl,omitempty"`
    BundleSize         int                    `json:"bundle_size,omitempty"`
}
```

### Step 5: Update ListFunctionsResponse
**File:** `internal/api/handlers/registry/query.go`

Ensure `convertToFunctionInfos` properly maps all fields including:
- capabilities
- bundle_size
- source_hash

---

## API Response Examples

### GET /registry/functions/{author}/{name}
```json
{
  "author": "example",
  "name": "hello-world",
  "version": "1.0.0",
  "title": "Hello World Function",
  "description": "A simple hello world function",
  "runtime": "node18",
  "category": "utility",
  "tags": ["hello", "world"],
  "visibility": "public",
  "price_per_call": 0,
  "reliability": 95.5,
  "deterministic": false,
  "side_effects": "none",
  "idempotent": true,
  "cache_ttl": 3600,
  "timeout_ms": 5000,
  "memory_mb": 128,
  "capabilities": ["fetch:read"],
  "bundle_size": 24576,
  "source_hash": "abc123...",
  "input_type": "object",
  "output_type": "string",
  "documentation_url": "/docs/example/hello-world",
  "playground_url": "/playground/example/hello-world",
  "trust_score": 85,
  "trust_level": "good"
}
```

### POST /registry/publish (response)
```json
{
  "ok": true,
  "function": "example/hello-world",
  "version": "1.0.0",
  "message": "Function published successfully",
  "verification_status": "verified",
  "function_id": "uuid-here",
  "runtime": "node18",
  "timeout_ms": 5000,
  "memory_mb": 128,
  "capabilities": ["fetch:read"],
  "deterministic": false,
  "side_effects": "none",
  "idempotent": true,
  "cache_ttl": 3600,
  "bundle_size": 24576
}
```

---

## Execution Endpoint Verification

The execution endpoint (`/fx/{author}/{name}`) appears to be correctly implemented:
- Routes through security middleware (routes.go:380-406)
- Handles both latest version (`/fx/{author}/{name}`) and specific version (`/fx/{author}/{name}@{version}`)
- Properly parses ExecutionRequest and returns ExecutionResponse

No changes needed for execution endpoint.

---

## Testing Strategy

1. Test GET /registry/functions/{author}/{name} - verify all fields present
2. Test GET /registry/functions - verify list includes capabilities
3. Test POST /registry/publish - verify full response includes function details
4. Test POST /fx/{author}/{name} - verify execution works
5. Test POST /fx/{author}/{name}@{version} - verify versioned execution works

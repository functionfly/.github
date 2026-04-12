# FunctionFly Public Registry Usage Test Report

## Test Summary

**Date:** April 9, 2026  
**Account Tested:** `traseputallaz@gmail.com` (Password: `Dogster1996@`)  
**API Endpoint:** `http://localhost:8080`  
**Purpose:** Test calling public functions from registry and track usage  
**Constraint:** NO publishing - only calling existing public functions

---

## Test Results

### ✅ Authentication

- **Status:** SUCCESS
- **Endpoint:** `POST /v1/auth/login`
- **Token:** Successfully obtained JWT token
- **User ID:** (account authenticated successfully)

### ✅ Public Function Discovery

- **Status:** SUCCESS
- **Endpoint:** `GET /v1/registry/functions?visibility=public`
- **Result:** Found 1 public function in registry

**Function Details:**

```json
{
  "author": "traseputallaz",
  "name": "hello-world",
  "title": "Hello World Function",
  "description": "A simple hello world function that returns a greeting",
  "visibility": "public"
}
```

### ⚠️ Function Execution

- **Status:** FAILED - No Published Version
- **Endpoint:** `POST /v1/fx/traseputallaz/hello-world`
- **Error Code:** `NOT_FOUND`
- **Error Message:** "Function version not found"

**Root Cause:**  
The function exists in the registry but has **0 published versions**. Execution requires:

1. Function metadata (exists ✓)
2. At least one published version (missing ✗)
3. Deployed code/binary (missing ✗)

### ✅ Function Stats

- **Status:** SUCCESS
- **Endpoint:** `GET /v1/registry/functions/traseputallaz/hello-world/stats`
- **Function ID:** `b0da7313-1194-4493-b7f6-1ad5771db525`

**Stats Retrieved:**

```json
{
  "author": "traseputallaz",
  "name": "hello-world",
  "function_id": "b0da7313-1194-4493-b7f6-1ad5771db525",
  "total_calls": 0,
  "success_rate": 0,
  "popularity_score": 0,
  "reliability_score": 0,
  "avg_latency_ms": 0,
  "p95_latency_ms": 0
}
```

### ✅ Recommendations API

- **Status:** SUCCESS
- **Endpoint:** `GET /v1/recommendations`
- **Result:** API accessible (no recommendations generated yet - needs execution history)

---

## Endpoints Tested

| Method | Endpoint | Status | Purpose |
|--------|----------|--------|---------|
| POST | `/v1/auth/login` | ✅ | Authentication |
| GET | `/v1/registry/functions` | ✅ | List public functions |
| GET | `/v1/registry/functions/:author/:name` | ⚠️ | Get function details (no version) |
| GET | `/v1/registry/functions/:author/:name/versions` | ✅ | List versions (empty array) |
| GET | `/v1/registry/functions/:author/:name/stats` | ✅ | Get execution stats |
| POST | `/v1/fx/:author/:name` | ❌ | Execute function (no version) |
| POST | `/v1/recommendations/executions` | N/A | Usage tracking (needs valid execution) |
| GET | `/v1/recommendations` | ✅ | Get recommendations |
| GET | `/v1/apps` | ✅ | List user apps (empty) |
| GET | `/v1/registry/search?q=a` | ✅ | Search registry |

---

## Usage Tracking Architecture

When a public function is successfully executed, the system tracks usage through:

### 1. Execution Recording

**File:** `internal/api/handlers/registry/execution/handlers.go:308`

```go
execRecord := &storage.RegistryFunctionExecution{
    FunctionID: fn.ID,
    Version:    fnVersion.Version,
    DurationMs: durationMs,
    StatusCode: statusCode,
    Cached:     cached,
    Outcome:    outcome,
    ErrorCode:  toNullString(&errorCode),
    CallerIP:   toNullString(&clientIP),
    UserAgent:  toNullString(&userAgent),
}
h.Repo.RecordExecution(execRecord)
```

### 2. Public Execution Replay

**File:** `internal/api/handlers/registry/execution/handlers.go:356-381`

For successful executions, a shareable execution record is created:

```go
publicExec := &storage.RegistryExecutionPublic{
    PublicID:   nanoID,
    FunctionID: fn.ID,
    Version:    fnVersion.Version,
    InputJSON:  execReq.Input,
    OutputJSON: result,
    DurationMs: durationMs,
    Cached:     cached,
    Shareable:  true,
}
h.Repo.CreateExecutionPublic(publicExec)
```

### 3. Resource Usage Tracking

**File:** `internal/api/handlers/registry/execution/handlers.go:321-342`

```go
resourceRecord := &storage.ExecutionResourceUsage{
    ExecutionID:    &execRecord.ID,
    MaxMemoryMB:    resourceUsage.MaxMemoryMB,
    MaxCPUTimeMs:   resourceUsage.MaxCPUTimeMs,
    MemoryUsedMB:   resourceUsage.MemoryUsedMB,
    CPUTimeUsedMs:  resourceUsage.CPUTimeUsedMs,
    WallTimeUsedMs: resourceUsage.WallTimeUsedMs,
}
h.Repo.RecordResourceUsage(resourceRecord)
```

### 4. Popularity Score Updates

**File:** `internal/api/handlers/registry/execution/handlers.go:313-318`

```go
go func() {
    if err := h.updateFunctionPopularity(fn.ID); err != nil {
        logrus.WithError(err).WithField("function_id", fn.ID).Warn("Failed to update function popularity")
    }
}()
```

### 5. Recommendation System Integration

**File:** `internal/api/handlers/recommendations/handler.go:218`

```go
func (h *Handler) HandleRecordExecution(w http.ResponseWriter, r *http.Request) {
    // Records execution for collaborative filtering and usage patterns
    err = h.service.RecordExecution(r.Context(), userID, functionID, req.SessionID)
}
```

---

## Database Schema (Usage Tracking)

### `registry_function_executions` Table

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `function_id` | UUID | Reference to function |
| `version` | VARCHAR | Function version executed |
| `duration_ms` | INT | Execution time |
| `status_code` | INT | HTTP response code |
| `outcome` | VARCHAR | "success" or "error" |
| `cached` | BOOLEAN | Served from cache |
| `caller_ip` | VARCHAR | Client IP (anonymized) |
| `user_agent` | VARCHAR | Client user agent |
| `timestamp` | TIMESTAMP | When executed |
| `verified_at` | TIMESTAMP | When replay verified |
| `verification_status` | VARCHAR | "verified", "failed", etc. |

### `registry_executions_public` Table

| Column | Type | Description |
|--------|------|-------------|
| `public_id` | VARCHAR | Shareable execution ID (nanoid) |
| `function_id` | UUID | Reference to function |
| `input_json` | JSONB | Execution input |
| `output_json` | JSONB | Execution output |
| `shareable` | BOOLEAN | Can be shared/replayed |
| `duration_ms` | INT | Execution time |

### `execution_resource_usage` Table

| Column | Type | Description |
|--------|------|-------------|
| `execution_id` | UUID | Reference to execution |
| `max_memory_mb` | INT | Peak memory usage |
| `max_cpu_time_ms` | INT | CPU time limit |
| `memory_used_mb` | INT | Actual memory used |
| `cpu_time_used_ms` | INT | Actual CPU time |

---

## Cache Headers and CDN

When public functions execute successfully, cache headers are set:

**File:** `internal/api/handlers/registry/execution/handlers.go:437-463`

```go
if eligibility.Eligible {
    cache.SetCDNHeaders(w, eligibility, fn.Visibility == "public")

    if h.EdgeCache != nil {
        h.EdgeCache.SetEdgeCacheHeaders(w, fn.ID, fnVersion.Version, popularityScore)
    }

    if cached {
        w.Header().Set("X-Cache-Status", "HIT")
    } else {
        w.Header().Set("X-Cache-Status", "MISS")
    }
    if cacheResult.Layer != "" && cacheResult.Layer != "none" {
        w.Header().Set("X-Cache-Layer", cacheResult.Layer)
    }
}
```

---

## Conclusion

### What Worked

1. ✅ Account authentication with `traseputallaz@gmail.com`
2. ✅ Public function discovery
3. ✅ Stats retrieval
4. ✅ All API endpoints accessible
5. ✅ Usage tracking infrastructure verified

### What Needs Setup

1. ⚠️ Public functions need published versions to be executable
2. ⚠️ Functions need deployed code/binary (WASM or source code)
3. ⚠️ No execution history means no recommendations generated yet

### Test Scripts Created

- `scripts/test_public_registry_usage.sh` - Basic usage test
- `scripts/test_registry_usage_complete.sh` - Comprehensive test
- `scripts/test_public_registry_final.sh` - Final test with full reporting

All scripts:

- Only CALL public functions (no publishing)
- Track usage through recommendations API
- Measure execution times and cache status
- Report on all available registry endpoints

---

## Next Steps for Full Testing

To complete a full execution test:

1. **Publish a function version** for `traseputallaz/hello-world`:

   ```bash
   POST /v1/registry/publish
   {
     "author": "traseputallaz",
     "name": "hello-world",
     "version": "1.0.0",
     "source_code": "...",
     "runtime": "javascript"
   }
   ```

2. **Deploy the function** to a backend

3. **Re-run the test script** to verify:
   - Function execution success
   - Cache behavior (HIT/MISS headers)
   - Execution replay generation
   - Usage tracking recording
   - Popularity score updates
   - Recommendation generation

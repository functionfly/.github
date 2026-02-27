# Instant Playground URLs - Implementation Plan

## Overview

Every function automatically gets a public playground URL: `functionfly.com/run/{app-slug}/{function-name}` that opens a UI to test it in the browser. This creates viral shareability - users can share links like "Try my AI tool" with a single URL.

## URL Structure

```
https://functionfly.com/run/{app-slug}/{function-name}
```

**Examples:**

- `functionfly.com/run/myapp/sentiment-analyzer`
- `functionfly.com/run/chatgpt-clone/assistant`

## Architecture

```mermaid
flowchart TD
    A[User visits /run/{appSlug}/{functionName}] --> B[Caddy receives request]
    B --> C{Static SPA or API?}
    C -->|HTML/JS| D[Serve Playground UI]
    C -->|API| E[Proxy to deployed function]
    D --> F[User enters input in UI]
    F --> G[POST /run/{appSlug}/{functionName}/execute]
    E --> H[Execute function]
    H --> I[Return result to UI]
    G --> I
```

## Implementation Steps

### 1. Database Migration

**File:** `internal/storage/sql/migrations/`

Add `playground_enabled` column to functions table:

- `playground_enabled` (BOOLEAN, DEFAULT TRUE)
- `playground_config` (JSONB, optional - stores default params, title, description)

### 2. Repository Layer

**File:** `internal/storage/function_repository.go`

Add new method:

```go
GetFunctionByAppSlugAndName(ctx context.Context, appSlug, functionName string) (*FunctionConfig, error)
```

### 3. API Handler

**File:** `internal/api/handlers/playground/playground.go` (NEW)

Endpoints:

- `GET /run/{appSlug}/{functionName}` - Serve playground UI HTML
- `GET /run/{appSlug}/{functionName}/info` - Get function metadata (public info)
- `POST /run/{appSlug}/{functionName}/execute` - Execute function with provided input

### 4. Route Registration

**File:** `internal/api/routes.go`

```go
// Public playground routes (no auth required)
api.HandleFunc("/run/{appSlug}/{functionName}", playgroundHandler.HandlePlaygroundUI).Methods("GET")
api.HandleFunc("/run/{appSlug}/{functionName}/info", playgroundHandler.HandleGetFunctionInfo).Methods("GET")
api.HandleFunc("/run/{appSlug}/{functionName}/execute", playgroundHandler.HandleExecute).Methods("POST", "OPTIONS")
```

### 5. Frontend Playground UI

**Files:** `web/dashboard/src/pages/Playground.tsx` (NEW)

Components:

- Input form (JSON or form fields based on function signature)
- Execute button
- Response viewer (JSON/text with syntax highlighting)
- Share button (copies URL with optional pre-filled params)
- History of recent executions (localStorage)

### 6. Caddy Configuration

**File:** `deploy/caddy/Caddyfile`

Add route for `/run/` path:

```
handle /run/* {
    reverse_proxy orchestrator-api:8080
}
```

### 7. Security Considerations

- **Rate limiting:** More permissive than authenticated endpoints but still limited (e.g., 60 req/min per IP)
- **Function validation:** Only deployed functions can be playground-enabled
- **Environment vars:** Functions run with restricted env vars in playground mode (no secrets)
- **CORS:** Allow cross-origin requests for embedding in external sites

### 8. Analytics

Track playground usage:

- Playground opens (unique URL visits)
- Function executions (grouped by function)
- Share link generation

## Shareable Link Format

Pre-filled parameters can be encoded in URL:

```
functionfly.com/run/myapp/sentiment-analyzer?input=Hello%20world&title=Sentiment%20Demo
```

The UI should parse these and pre-fill the input fields.

## UI Features

1. **Function Info Panel:**
   - Function name and description
   - Provider/deployment status indicator
   - Version info

2. **Input Section:**
   - Auto-detect input format (JSON/form)
   - Schema-based form generation if function provides OpenAPI spec
   - File upload support for functions that accept files

3. **Output Section:**
   - JSON syntax highlighting
   - Response time display
   - Error formatting
   - Copy result button

4. **Share Panel:**
   - Generate shareable link
   - Optional: Add custom title/description for the share preview

## Implementation Order

1. Create migration
2. Add repository method
3. Create API handler with basic proxy logic
4. Register routes
5. Create frontend UI
6. Update Caddy config
7. Add rate limiting
8. Test end-to-end

## File Changes Summary

| Action | File |
|--------|------|
| CREATE | `internal/storage/sql/migrations/YYYYMMDDHHMMSS_add_playground_support.up.sql` |
| CREATE | `internal/storage/sql/migrations/YYYYMMDDHHMMSS_add_playground_support.down.sql` |
| MODIFY | `internal/storage/function_repository.go` |
| CREATE | `internal/api/handlers/playground/playground.go` |
| MODIFY | `internal/api/routes.go` |
| CREATE | `web/dashboard/src/pages/Playground.tsx` |
| CREATE | `web/dashboard/src/api/playground.ts` |
| MODIFY | `deploy/caddy/Caddyfile` |

## Notes

- The playground UI should work standalone without authentication
- Functions must be deployed to be accessible via playground
- Consider adding a "badge" or widget that can be embedded in external sites
- Track playground usage in analytics for viral coefficient measurement

# API Error Handling Hardening - Implementation Design

**Date:** 2026-06-18
**Status:** Draft
**Scope:** Eliminate leak of internal error details (`err.Error()`) to API clients across all handlers; consolidate on the existing `internal/apierror` package.

---

## Problem

A scan of the Go codebase finds **180+ sites** that return `err.Error()` text directly to API clients across ~50 files. Additional **1447** `http.Error(w, ...)` calls in 110 files bypass the canonical `internal/apierror` package even when they use hardcoded (safe) messages, producing an inconsistent error response shape across the API.

**Leak patterns identified:**

| Pattern | Sites | Example |
|---------|-------|---------|
| `writeError(w, status, code, err.Error())` | 110 | `agent/execute.go`, `billing/*` |
| `respondError(w, status, err.Error())` | 43 | `gba/plugins/{webauthn,scim}/handlers.go` |
| `writeJSONError(w, status, err.Error())` | 15 | `marketplace`, `auth/saml` |
| `apierror.NewXxx(err.Error())` | 32 | `trustapi/billing_handlers.go`, `admin/dedicated_db.go` |
| `http.Error(w, ..., err.Error())` | 12 | `schedule/handler.go`, `gba/handlers.go` |
| `apierror.WriteError(..., apierror.NewXxx("ctx: "+err.Error()))` | 30 | `admin/dedicated_db.go` |
| **Total leaks** | **~180** | |

**Hardcoded but inconsistent (no actual leak, but bypass canonical package):** ~1447 `http.Error(w, "msg", status)` calls in 110 files.

### Critical finding

`internal/api/middleware/error_normalizer.go` defines `ErrorNormalizerMiddleware` (with tests) that intercepts `http.Error` responses and converts them to the structured `apierror.APIError` JSON format. **It is not currently registered in the route chain in `internal/api/routes.go`.** Wiring it up would provide an instant blanket fix for the `http.Error` leak pattern, but it does not catch the other 168 leak sites that use `respondError`, `writeError`, `writeJSONError`, or `apierror.NewXxx` directly.

---

## Goals

1. **Zero leaks of internal error details** to clients, including stack traces, SQL error text, file paths, library names, and connection strings.
2. **One canonical error response shape** across the entire API (the existing `apierror.APIError` JSON envelope).
3. **Server-side observability preserved** — every internal error is still logged with full context (request id, handler context, original err).
4. **Defense in depth** — both handler-level and middleware-level controls so that future code cannot accidentally regress.
5. **Zero behavior change for legitimate client-visible messages** (e.g. "User not found", "Invalid API version") — only `err.Error()` and internal context gets sanitized.

---

## Design

### Architecture

```
HTTP Request
     │
     ▼
┌──────────────────────────────────────────────────────────────┐
│ middleware.RecoveryMiddleware          (panic → 500)        │
│ middleware.ErrorNormalizerMiddleware   (NEW: sanitizes)     │  ← safety net
│ middleware.TracingMiddleware           (request id)         │
│ middleware.EnvironmentMiddleware                             │
│ ... other middleware ...                                     │
└────────────────────────┬─────────────────────────────────────┘
                         ▼
┌──────────────────────────────────────────────────────────────┐
│ Handler:                                                    │
│   apierror.LogAndInternal(r, err, "create agent")            │
│   apierror.WriteError(w, apierror.NewNotFound("User not found"))│
│   http.Error(w, "...", 500)  // safety net catches this     │
└────────────────────────┬─────────────────────────────────────┘
                         ▼
        ┌────────────────┴────────────────┐
        ▼                                  ▼
   Server log:                        Client response:
   ERROR context=...                  {"code":"INTERNAL_ERROR",
          error="pq: duplicate               "message":"Internal server
          key value..."                            error",
          request_id=req_abc123                 "request_id":"req_abc123"}
```

### Two layers of defense

**Layer 1: Handler-level (primary)**
- New `apierror.LogAndXxx(r, err, ctx)` family of helpers. These log the full `err` with context + request id server-side and emit a sanitized response with a generic message.
- New `apierror.FromError(r, err)` mapper that inspects an error and produces the appropriate `*APIError` (with sane default of `INTERNAL_ERROR` + generic message + `Detail` populated only in dev).

**Layer 2: Middleware (safety net)**
- `ErrorNormalizerMiddleware` wired into the route chain immediately after `RecoveryMiddleware`.
- Catches any remaining `http.Error(w, ..., err.Error(), ...)` calls and rewrites the body to a sanitized apierror envelope, preserving only status code and request id.
- Already has a `DISABLE_ERROR_NORMALIZER=true` escape hatch and skips WebSocket upgrades.

### `apierror` package additions

**File:** `internal/apierror/logging.go` (new)

```go
package apierror

// LogAndInternal logs err server-side with full context, then writes a
// generic 500 response to the client. Use this for any 5xx error where
// the underlying err is from a dependency (DB, network, third-party API).
func LogAndInternal(r *http.Request, err error, contextMsg string)

// LogAndBadRequest — 400 with err logged
func LogAndBadRequest(r *http.Request, err error, contextMsg string)

// LogAndNotFound — 404 with err logged
func LogAndNotFound(r *http.Request, err error, contextMsg string)

// LogAndConflict — 409 with err logged
func LogAndConflict(r *http.Request, err error, contextMsg string)

// LogAndForbidden — 403 with err logged
func LogAndForbidden(r *http.Request, err error, contextMsg string)

// LogAndServiceUnavailable — 503 with err logged
func LogAndServiceUnavailable(r *http.Request, err error, contextMsg string)

// FromError inspects err and returns the best-fit *APIError. The returned
// APIError always uses a generic Message; the original err is logged by
// the caller via LogAndInternal. Detail is populated only in development.
func FromError(err error) *APIError

// SanitizeMessage returns a generic message appropriate for the status code.
// If body is a static, hand-written string, it can be passed through; if it
// looks like a raw err message (contains pq:, sql:, json:, etc.) it is replaced.
func SanitizeMessage(status int, body string) string
```

**Shared logger hook:** `logAPIError(r *http.Request, err error, ctx, level string)` uses `logrus.WithError(err).WithFields(...)` so request_id and context are always present.

### Per-package helper consolidation

Files with local helpers (e.g. `internal/api/handlers/agent/handler.go:6` defines `writeError(w, status, code, message string)`) keep their signature for minimal diff, but their body delegates to `apierror.WriteError` underneath. For err-bearing paths, callers switch to `apierror.LogAndInternal(r, err, ctx)`.

Examples of local helpers to consolidate (representative — full list in PR3):
- `agent/{handler,execute,daemon,sebg,evolution,lifecycle,...}.go` — all use local `writeError`
- `apikeys/create.go`
- `billing/{state_usage_handler,usage_handlers,export_handlers,cost_allocation_handlers,external_billing_handlers,wallet_export_handlers}.go`
- `deploykeys/handler.go`
- `function_webhooks/handler.go`
- `registry/{execution/handlers,verification_pipeline}.go`
- `trustapi/webhook_handlers.go`
- `agentruntime/handler.go`
- `gba/plugins/{webauthn,scim}/handlers.go` — uses `respondError`
- `gba/handlers.go` — uses `http.Error(w, ...err.Error()...)`
- `frg/handlers.go`
- `auth/{webauthn,saml,oauth_handlers,signup_handlers,password_handlers}.go` — uses `writeJSONError`
- `marketplace/handler.go`
- `payouts/handler_extended.go`
- `studio/projects_handler.go`
- `tax_handlers.go`
- `provisioning/handler.go`
- `mcp/global_settings.go`
- `schedule/handler.go`
- `functions/paste.go`
- `docs/docs.go`

### CI guard

**File:** `scripts/check-error-leaks.sh` (new, executable)

```bash
#!/usr/bin/env bash
set -euo pipefail
PATTERNS=(
  'http\.Error\(w, .*err\.Error\(\)'
  'respondError\(w, [^,]+, [^,]*(err\.Error\(\)|fmt\.Sprintf.*err\.Error\(\))'
  'writeError\(w, [^,]+, [^,]+, .*err\.Error\(\)'
  'writeJSONError\(w, [^,]+, .*err\.Error\(\)'
  'apierror\.New[A-Z][a-zA-Z]*\(.*err\.Error\(\)'
  'apierror\.New[A-Z][a-zA-Z]*\(.*\+ ?err\.Error\(\)'
)
# ... grep + fail if any matches
```

Wired into CI alongside `golangci-lint`. Allows existing sites to be grandfathered via `// nolint:errorleak` comment on the line above, which the migration PRs remove.

---

## Migration Order

Four PRs, ordered smallest to largest so each is independently reviewable.

### PR 1 — Wire up the normalizer (instant blanket fix)

- **File:** `internal/api/routes.go`
- Register `middleware.ErrorNormalizerMiddleware` immediately after `middleware.RecoveryMiddleware` (line ~XXX).
- **Tests:** Add a test in `internal/api/middleware/error_normalizer_test.go` that mounts a handler calling `http.Error(w, err.Error(), 500)` and asserts the body does not contain the err text.
- **Risk:** Low. The middleware has existed and been tested since before, but was never enabled. It already skips WebSocket upgrades and has a kill switch.
- **Size:** ~10 lines changed.

### PR 2 — Add `LogAndXxx` helpers and `FromError` mapper

- **Files:**
  - `internal/apierror/logging.go` (new, ~120 lines)
  - `internal/apierror/logging_test.go` (new, ~150 lines)
- Pure addition. No callers yet. Tests for each helper and `FromError`.
- **Size:** ~270 lines added.

### PR 3 — Migrate the 180 leak sites

- One commit per domain. ~10 sub-PRs:
  1. `agent/` (~30 sites)
  2. `billing/` (~25 sites)
  3. `auth/` and `gba/` (~25 sites)
  4. `trustapi/`, `apikeys/`, `deploykeys/`, `function_webhooks/` (~20 sites)
  5. `registry/` (~10 sites)
  6. `marketplace/`, `payouts/`, `studio/`, `tax_handlers/` (~10 sites)
  7. `mcp/`, `schedule/`, `functions/`, `docs/` (~10 sites)
  8. `frg/`, `provisioning/`, `agentruntime/` (~10 sites)
  9. Middleware-level leaks (`execution_security.go`)
  10. apierror leaks in `admin/`, `vault/`, `version/`, `trustapi/billing_handlers.go`
- Each leak site: replace with `apierror.LogAndInternal(r, err, "<context>")` (or appropriate status). Add a one-line `logrus.WithError(err).Error("context")` for callers that need extra fields.
- **Size:** ~50 files changed, ~180 sites, mechanical.

### PR 4 — Migrate the 1447 hardcoded `http.Error` calls + add CI guard

- One commit per domain. ~10 sub-PRs.
- `http.Error(w, "msg", status)` → `apierror.WriteError(w, apierror.NewXxx("msg"))`
- Update per-package `writeError`/`writeJSONError` helpers to delegate to `apierror` underneath.
- **Add:** `scripts/check-error-leaks.sh` and wire it into CI to prevent regression.
- **Size:** ~110 files, ~1447 sites.

---

## Testing

- `internal/api/middleware/error_normalizer_test.go` — already exists; expand with leak-site cases.
- `internal/apierror/logging_test.go` (new) — unit tests for each `LogAndXxx` and `FromError`.
- `internal/api/handlers/agent/execute_test.go` (or equivalent) — add a "leak prevention" test that mocks a dependency to return a known `err.Error()` (e.g. `"pq: duplicate key"`) and asserts the response body is `INTERNAL_ERROR` + generic message.
- CI guard `scripts/check-error-leaks.sh` runs on every PR.

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Middleware breaks streaming or upgrade responses | Already skips WebSocket; has `DISABLE_ERROR_NORMALIZER` escape hatch. Will add explicit test for streaming chunked responses. |
| Middleware perf cost (parse body) | Only runs on 4xx/5xx; success path is passthrough. No body buffer until status >= 400. |
| Migration PRs too large to review | Split by domain (~10 sub-PRs each for PR3 and PR4). |
| `apierror` package becomes a god package | Helpers are thin; `FromError` lives in a separate `internal/apierror/mapper.go`. |
| Existing clients depend on old error shape | The existing normalizer already produces the same shape; new helpers preserve all field names (`code`, `message`, `request_id`, `detail`, `field`). |
| Tests that assert specific err text in response body will break | Documented in PR3; test authors update to assert sanitized output. |
| 30+ `apierror.NewXxx("ctx: "+err.Error())` calls need the same treatment as raw leaks | Covered by PR3.10 (apierror leaks sub-PR). |

---

## Out of Scope

- Changes to error response shape (preserved as-is).
- New error codes or status codes.
- Telemetry/metrics on error rates (already handled by `monitoringPkg.HTTPMetricsMiddleware`).
- Frontend error display changes (the dashboard already handles the structured envelope).
- Refactoring of handlers themselves beyond the error-return line.

---

## Success Criteria

- Zero matches of the leak patterns in the CI guard script (after the migration is complete).
- All 1447 `http.Error` calls replaced with `apierror.WriteError` + `apierror.NewXxx`.
- `ErrorNormalizerMiddleware` registered in `routes.go` and tested.
- All existing tests pass.
- New `apierror.LogAndXxx` and `apierror.FromError` helpers covered by unit tests.
- CI guard script runs on every PR and fails on new leak sites.

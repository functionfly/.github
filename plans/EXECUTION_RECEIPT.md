# The Execution Receipt — Architecture Plan

**Feature:** A public, shareable URL generated after every successful function execution that doubles as a viral distribution surface.
**Why now:** Every function execution is currently a private, terminal event. We get zero distribution from a developer's most viral moment — "look what I just built and ran." The receipt turns each run into a permanent, indexable, runnable artifact.
**Effort:** Medium. **Impact:** 10x distribution. **Timeframe:** Week 1.

---

## 0. The Big Shortcut — What's Already There

Audit of the current codebase (`internal/api/handlers/registry/execution/handlers.go`, `internal/storage/registry/types.go`, `web/dashboard/src/pages/ReplayPage/`) shows the foundation is **80% built**:

| Already exists | Where | Reuse strategy |
|---|---|---|
| `RegistryExecutionPublic` table (nanoid `PublicID`, input/output, duration, cached, verification) | `internal/storage/registry/types.go:244-266` | Receipt = row in this table, with a few denormalized columns added |
| `HandleGetReplay` returns the execution to the SPA | `handlers.go:852-892` | Wrap in a richer `/v1/receipts/:id` handler; keep replay as the deeplink in dashboards |
| `execution_id` (nanoid) returned in every successful execution response | `writeSuccessResponse`, `handlers.go:672-735` | Receipt URL = `${PUBLIC_BASE}/r/${executionID}` |
| Privacy-sanitized input/output for shareable executions | `PrivacyService.SanitizeInputOutput` | Receipt inherits the same gate |
| `/replay/:execId` SPA route + page (shadcn/ui, Navbar, Footer) | `App.tsx:527`, `ReplayPage/index.tsx` | New `/r/:id` route that mounts a richer `<ReceiptPage>` — copy ReplayPage as the structural template |
| SPA catch-all matcher covers `/replay/*` and `/fx/*` for no-auth GETs | `routes.go:1198-1209` | Add `r` and `receipt` to the path list |
| `realtimeUsageTracker` (Redis-backed, has `RecordExecution`) | `services/realtime_usage_tracker.go` | Hook milestone detection here |
| `function_webhook_service` (`internal/storage/function_webhook_service.go`) | `internal/storage/` | Use to fan out the tweet/notify event |
| Existing scheduler framework (`internal/scheduler/`) | `internal/scheduler/` | Add a new `ReceiptMilestoneScheduler` |
| `notificationSvc` + `notificationRepo` | `routes.go:190-195` | Drop in-app + email notifications from milestone |
| `Resend` (email) already integrated | AGENTS.md | Email "your function just ran 100×" |
| Twitter/X web intent URL (`https://twitter.com/intent/tweet?text=...`) | n/a | Zero-OAuth tweet; user clicks → pre-filled tweet |

**Conclusion:** The data model, capture point, public page, privacy gate, and scheduler plumbing are all in place. The "Receipt" is a **new consumer of the existing public-execution infrastructure** plus (a) denormalized function metadata, (b) a "live call" panel, (c) a fork endpoint, (d) a milestone worker, and (e) a redesigned public page.

---

## 1. Microservices Decomposition

We do **not** introduce new services. We add a single new bounded context — `receipt` — that owns all Receipt logic and reads from the existing registry + function-registry + notification services. This keeps blast radius small and respects "compose, don't create."

```
┌──────────────────────────────────────────────────────────────────────┐
│  Public web (Vercel-hosted, Astro for /r, React SPA for /app/r)     │
│  └── /r/:id  → React SPA page (RSC-friendly static shell)            │
└────────────┬─────────────────────────────────────────────────────────┘
             │ GET /v1/receipts/:id   (public, no auth, CORS open)
             │ POST /v1/receipts/:id/run  (public, no auth, rate-limited)
             │ GET /v1/receipts/:id/fork-payload  (public, no auth)
             ▼
┌──────────────────────────────────────────────────────────────────────┐
│  Go Orchestrator API (existing)                                      │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │  NEW: internal/api/handlers/receipt/                            │  │
│  │   - receipt_handler.go  (GetReceipt, RunReceipt, ForkPayload)   │  │
│  │   - milestone_worker.go (fire on 1, 10, 100, 1000, 10k runs)    │  │
│  └────────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │  EXISTING: registry/execution/HandleExecute (handlers.go:27)    │  │
│  │   ↳ generateExecutionID() (handlers.go:548) — entry point       │  │
│  └────────────────────────────────────────────────────────────────┘  │
└────────────┬─────────────────────────────────────────────────────────┘
             │
   ┌─────────┴──────────┐
   ▼                    ▼
┌─────────┐      ┌────────────────┐
│Postgres │      │ Redis (existing)│
│ + GORM  │      │  realtimeUsage  │
└─────────┘      │  rateLimit      │
                └────────────────┘
```

### Boundaries

- **Public API surface (new):** `/v1/receipts/:id*` — 3 routes, no auth, mounted on the existing `s.router` (root) so the SPA catch-all keeps working and the existing CORS middleware (in `advancedSecurityMiddleware.CORSMiddleware`) applies. We add `r` and `receipt` to the SPA path-prefix list at `routes.go:1207-1208`.
- **Internal only (new):** milestone scheduler — runs in-process on the orchestrator, scheduled via the existing `internal/scheduler/` framework.
- **Read models:** Receipt handler reads from the existing `RegistryRepository` (no new repo needed). We add a thin `receipt_repository.go` only if the join complexity warrants it (it does — see §3).

### Why not a separate service?

1. Receipts are read-mostly, single-source-of-truth (Postgres), and the write path is one extra line in an existing handler. A new service adds a network hop on the hot execution path.
2. The existing `internal/functionregistry` is the right home for execution-related domain logic; the receipt is a projection of that domain.
3. The system prompt's "scale from day one" doesn't mean "create a new service for every feature" — it means "design for scale." Receipt reads will scale via: (a) Postgres read replica (`s.postgresDB.ReadPool` already exists in the codebase), (b) Redis cache layer (3-min TTL — see §6), (c) CDN cache headers (`cache.SetCDNHeaders` already used in `writeSuccessResponse`).

---

## 2. Data Flow

### Capture (write path)

```
1. Client hits POST /v1/fx/{author}/{name}  (or scheduled/triggered execution)
2. HandleExecute runs the function via RuntimeRouter  (handlers.go:62-98)
3. On success, generateExecutionID() creates a RegistryExecutionPublic row
   ↳ THIS is our existing receipt row. We extend it.
4. NEW: After row is committed, increment a Redis counter
   INCR  ff:rcpt:milestone:{function_id}  EX 90d
5. NEW: Spawn go-routine h.checkMilestone(fn.ID, ownerID)
   - if counter crossed {1, 10, 100, 1000, 10000}, enqueue receipt_milestone_event
   - dedupe with SETNX ff:rcpt:milestone:fired:{function_id}:{threshold} EX 30d
6. NEW: Receipt row gets denormalized function_name/author/runtime/schema
   (moved from join into the row at write time — see §3)
```

### Read (public path)

```
Client → GET /r/:nanoid  (Vercel/edge, no auth)
   ↓
React SPA (web/dashboard) mounts <ReceiptPage>
   ↓ useQuery(["receipt", id])
GET /v1/receipts/:id  (no auth, CORS open, rate-limited 30/min/IP)
   ↓
Handler.GetReceipt:
  - L1: Redis GET ff:rcpt:body:{id}  (TTL 60s)
  - L2: Postgres SELECT receipt + function name/runtime/schema
  - Hydrate response → set L1 → return
   ↓
React renders. User clicks "Run with these inputs":
   POST /v1/receipts/:id/run  (body: { input: ... })
   ↓
Handler.RunReceipt:
  - Validate function is public
  - Reuse existing /v1/fx/{author}/{name} execution path
  - Charge wallet if paid (existing)
  - Return ExecutionResponse with new execution_id → client builds new /r/:id URL
```

### Milestone → tweet path

```
RealtimeUsageTracker.RecordExecution (existing)
  ↓ NEW: at the end, fire receiptMilestoneHook(functionID, tenantID, totalRuns)
  ↓
receiptMilestoneHook:
  - if totalRuns ∈ {1,10,100,1000,10000} AND not already fired for this bucket
  - enqueue Asynq job (or just go-routine + notificationRepo.Insert)
  - job: build tweet intent URL + email body, insert into notifications table,
    fire email via existing notificationSvc / Resend
  - dashboard shows in-app toast for the owner next time they log in
```

### Fork path

```
Client → "Fork this function" button on /r/:id
   ↓
window.location = `${APP_URL}/functions/new?fork=${base64(function_source)}&name=${name}&author=${author}`
   ↓
Existing /functions/new editor (FunctionEditorPage) detects `?fork=` param
   ↓
Pre-fills the editor with the source; user is NOT logged in → routed to /auth/signup?next=/functions/new?fork=...
   ↓ on signup
Existing onboarding flow picks up the next URL, lands them in the editor with the source pre-loaded
```

---

## 3. Database Schema

### Migration: `20260601180000_execution_receipt.up.sql`

```sql
-- A. Extend the existing shareable-execution table with denormalized receipt fields.
-- This avoids a join on every public read (hot path) and lets the row stand alone.
ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS function_name     TEXT,
  ADD COLUMN IF NOT EXISTS function_author   TEXT,
  ADD COLUMN IF NOT EXISTS runtime           TEXT,         -- e.g. "python3.11", "node20", "wasm"
  ADD COLUMN IF NOT EXISTS input_schema      JSONB,        -- JSON Schema of input
  ADD COLUMN IF NOT EXISTS output_schema     JSONB,        -- JSON Schema of output
  ADD COLUMN IF NOT EXISTS function_visibility TEXT DEFAULT 'public',  -- mirrors fn.Visibility
  ADD COLUMN IF NOT EXISTS description       TEXT,         -- short marketing blurb
  ADD COLUMN IF NOT EXISTS fork_count        INTEGER NOT NULL DEFAULT 0, -- materialized counter
  ADD COLUMN IF NOT EXISTS view_count        INTEGER NOT NULL DEFAULT 0, -- for analytics + trending
  ADD COLUMN IF NOT EXISTS last_viewed_at    TIMESTAMPTZ;

-- B. Backfill from registry_functions / registry_function_versions
UPDATE registry_executions_public ep
SET
  function_name   = f.name,
  function_author = f.author,
  runtime         = v.runtime,
  function_visibility = f.visibility,
  description     = f.description
FROM registry_functions f
JOIN registry_function_versions v
  ON v.function_id = f.id AND v.version = ep.version
WHERE ep.function_id = f.id
  AND ep.function_name IS NULL;

-- Add NOT NULL after backfill (safe because backfill covers all existing rows)
ALTER TABLE registry_executions_public
  ALTER COLUMN function_name   SET NOT NULL,
  ALTER COLUMN function_author SET NOT NULL,
  ALTER COLUMN runtime         SET NOT NULL;

-- C. Indexes for the two hot public read paths
CREATE INDEX IF NOT EXISTS idx_rcpt_public_id_active
  ON registry_executions_public (public_id)
  WHERE shareable = TRUE;                                 -- partial: only receipts

CREATE INDEX IF NOT EXISTS idx_rcpt_function_created
  ON registry_executions_public (function_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_rcpt_view_count
  ON registry_executions_public (view_count DESC, created_at DESC)
  WHERE shareable = TRUE;                                 -- "trending receipts" list

-- D. New table: per-recipient milestone log (idempotent fan-out)
CREATE TABLE IF NOT EXISTS receipt_milestone_events (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  function_id       UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
  tenant_id         UUID,                                 -- owner
  threshold         INTEGER NOT NULL,                    -- 1, 10, 100, 1000, 10000
  total_runs_at     INTEGER NOT NULL,
  public_id         TEXT NOT NULL,                        -- pointer to the canonical receipt
  fired_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  channels_fired    TEXT[] NOT NULL DEFAULT '{}',         -- {'inapp','email','tweet_intent'}
  dedupe_key        TEXT NOT NULL UNIQUE,                 -- function_id:threshold — one-shot
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_milestone_function
  ON receipt_milestone_events (function_id, fired_at DESC);
```

### Down migration: `20260601180000_execution_receipt.down.sql`
Reverses all of the above (drop indexes, drop table, drop columns).

### Why denormalize?
The receipt is the **hottest public read path in the product** (anyone with a link can hit it, and we want them to, and CDN will cache it). Denormalizing `function_name`, `author`, `runtime`, `description` onto the row removes a join on every read and lets the L1 Redis cache be a true 1-key GET. The cost is a slightly larger write (one extra UPDATE during `generateExecutionID`) — acceptable on a write path that already does an INSERT.

---

## 4. API Contract (all new)

All routes mounted on the **root router** (not `/v1`) via a new `registerReceiptRoutes(s.router, ...)` in `internal/api/handlers/receipt/router.go`. Public, no auth. Rate-limited. CORS open (the existing CORS middleware is permissive for `/api/*` — we mirror the same allow-list).

| Method | Path | Auth | Rate limit | Purpose |
|---|---|---|---|---|
| `GET`  | `/v1/receipts/:id` | none | 60/min/IP, 600/min/global | Full receipt payload |
| `POST` | `/v1/receipts/:id/run` | none | 10/min/IP, function-call-cost | Re-execute the function with the receipt's inputs (or override) |
| `GET`  | `/v1/receipts/:id/fork-payload` | none | 30/min/IP | Returns the function source + manifest for the "fork" CTA |
| `POST` | `/v1/receipts/:id/view` | none | best-effort, 60/min/IP | Increments `view_count` + `last_viewed_at` for trending |
| `GET`  | `/v1/receipts/trending` | none | 30/min/IP | "Trending receipts" for a homepage widget (cross-link) |

### `GET /v1/receipts/:id` — response shape

```json
{
  "id": "V1StGXR8_Z5jHi3B-myT",
  "function": {
    "name": "summarize-url",
    "author": "ada",
    "runtime": "python3.11",
    "version": "1.4.2",
    "visibility": "public",
    "description": "Fetches a URL and returns a 3-sentence summary.",
    "input_schema":  { "type": "object", "properties": { "url": { "type": "string" } }, "required": ["url"] },
    "output_schema": { "type": "object", "properties": { "summary": { "type": "string" } } }
  },
  "execution": {
    "input":         { "url": "https://example.com" },
    "output":        { "summary": "An example domain used in illustrations." },
    "duration_ms":   142,
    "cached":        false,
    "created_at":    "2026-06-01T18:42:11Z",
    "verification":  { "status": "verified", "verified_at": "..." }
  },
  "share": {
    "url":               "https://functionfly.com/r/V1StGXR8_Z5jHi3B-myT",
    "embed_url":         "https://functionfly.com/r/V1StGXR8_Z5jHi3B-myT/embed",
    "tweet_intent_url":  "https://twitter.com/intent/tweet?text=I+just+ran+ada%2Fsummarize-url+on+%40functionfly+%E2%9C%A8&url=https%3A%2F%2Ffunctionfly.com%2Fr%2FV1StGXR8_Z5jHi3B-myT",
    "og_meta": {
      "title":       "ada/summarize-url ran in 142ms · FunctionFly",
      "description": "I just ran ada/summarize-url — fetched a URL and returned a 3-sentence summary. Try it yourself.",
      "image":       "https://functionfly.com/api/og/receipt/V1StGXR8_Z5jHi3B-myT.png"
    }
  },
  "can_run":      true,                  // false if fn is private or removed
  "is_paid":      false,
  "price_per_call_usd": 0
}
```

### `POST /v1/receipts/:id/run`

Request:
```json
{ "input": { "url": "https://example.com" } }   // optional override
```

Response: **same as the existing `ExecutionResponse`** — `{ ok, data, cached, duration_ms, version, execution_id }`. The returned `execution_id` is a NEW nanoid that points to a new `RegistryExecutionPublic` row, so the user can immediately share the new run.

### `GET /v1/receipts/:id/fork-payload`

Returns the function source (read from the function version, not the receipt row) and a base64-encoded source blob. The frontend builds a `?fork=` URL for the editor.

### Errors (consistent with the rest of the codebase)

| Status | Code | When |
|---|---|---|
| 404 | `RECEIPT_NOT_FOUND` | nanoid unknown or `shareable = false` |
| 410 | `RECEIPT_REVOKED` | owner revoked the receipt |
| 403 | `FUNCTION_PRIVATE` | function is `visibility != public` |
| 429 | `RATE_LIMITED` | over the bucket |
| 502 | `RUN_FAILED` | re-run failed (UI shows output in error state, still 200 if user wants to see the run) |

### Handler skeleton

```go
// internal/api/handlers/receipt/handler.go
package receipt

type Handler struct {
    Repo           *registry.RegistryRepository
    Cache          *cache.CacheService           // existing
    RateLimiter    *middleware.DistributedRateLimiter  // existing
    UsageTracker   *services.RealtimeUsageTracker // existing
    NotificationSvc *notifications.Service
    Logger         *logrus.Logger
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request)  { ... }
func (h *Handler) Run(w http.ResponseWriter, r *http.Request)  { ... }
func (h *Handler) ForkPayload(w http.ResponseWriter, r *http.Request) { ... }
func (h *Handler) View(w http.ResponseWriter, r *http.Request) { ... }
```

---

## 5. Backend Wiring (precise change list)

### 5.1 New file: `internal/api/handlers/receipt/handler.go`
- `Get`, `Run`, `ForkPayload`, `View`, `Trending`
- Reuses: `cache.CacheService`, `registry.RegistryRepository.GetFunctionByAuthorName`, `registry.GetFunctionVersion`, `privacy.PrivacyService`, `middleware.DistributedRateLimiter`
- The `Run` handler **delegates to the existing execution path** rather than re-implementing it — refactor `HandleExecute` to extract `executeCore(ctx, fn, fnVersion, input, r)` that both `/v1/fx/...` and `/v1/receipts/:id/run` can call. This keeps billing, quota, verification, and execution semantics consistent.

### 5.2 New file: `internal/api/handlers/receipt/milestone.go`
- `func (h *ReceiptMilestoneWorker) OnExecution(ctx, functionID, tenantID, publicID)` — called from a single hook
- Reads/writes `receipt_milestone_events` (idempotent via `dedupe_key UNIQUE`)
- Builds tweet intent URL + email body; inserts a `notifications` row; calls `notificationSvc.SendEmail` for the owner

### 5.3 New file: `internal/storage/receipt/receipt_repository.go`
- `GetReceiptByID(ctx, publicID) (*Receipt, error)` — single denormalized read
- `IncrementViewCount(ctx, publicID) error` — fire-and-forget
- `GetTrendingReceipts(ctx, limit) ([]Receipt, error)` — backed by the partial index

### 5.4 Extend `internal/api/handlers/registry/execution/handlers.go`
Two surgical edits, both inside `HandleExecute`:

```go
// after generateExecutionID returns &publicExec.PublicID  (line 343)
if executionID != nil {
    // backfill the denormalized fields we just added to registry_executions_public
    if err := h.Repo.BackfillReceiptMetadata(*executionID, fn, fnVersion); err != nil {
        logrus.WithError(err).Warn("Failed to backfill receipt metadata")
    }
    // fire-and-forget milestone check
    if h.ReceiptMilestoneHook != nil {
        go h.ReceiptMilestoneHook(r.Context(), fn.ID, fn.TenantID, *executionID)
    }
}
```

`BackfillReceiptMetadata` is a single `UPDATE registry_executions_public SET function_name=..., function_author=..., runtime=..., input_schema=..., output_schema=..., description=... WHERE public_id = $1`. Existing rows from before this feature are covered by the backfill in the migration.

### 5.5 Extend `internal/api/routes.go`
- Import `receiptHandler "github.com/functionfly/functionfly/internal/api/handlers/receipt"`
- Initialize: `receiptRepo := receipt.NewRepository(s.postgresDB.GORM); receiptHandler := receipt.NewHandler(receiptRepo, ...)`
- Add `registerReceiptRoutes(s.router, receiptHandler, rateLimiter)` call (after `registerPublicWebhookRoutes`, before `registerRegistryRoutes` — the SPA matcher must come AFTER API routes so it never shadows a real route)
- Add `r` and `receipt` to the SPA path-prefix list in the `MatcherFunc` block at `routes.go:1207-1208`

### 5.6 New migration files
- `migrations/20260601180000_execution_receipt.up.sql`
- `migrations/20260601180000_execution_receipt.down.sql`
Run via `./scripts/create-migration.sh "execution_receipt"` and validate with `./scripts/validate-migrations.sh`.

### 5.7 Scheduler registration
- `internal/scheduler/receipt_milestone_scheduler.go` — sweeps `registry_executions_public` daily to detect function-level milestones that may have been missed during downtime (defensive belt-and-braces; primary path is real-time hook).

---

## 6. Caching, Rate-Limiting, Scale

| Layer | Mechanism | TTL | Owner |
|---|---|---|---|
| L0 | Cloudflare/Vercel CDN on `/r/:id` (static shell) | 24h, revalidate on demand | New |
| L1 | Redis `ff:rcpt:body:{id}` (full JSON payload) | 60s | New |
| L2 | Postgres (source of truth) | n/a | Existing |
| L3 | `Cache-Control: public, max-age=30, s-maxage=300, stale-while-revalidate=86400` on `/v1/receipts/:id` | 5m CDN | New |

| Concern | Mechanism | Code location |
|---|---|---|
| Public rate limit (read) | 60/min/IP via `middleware.DistributedRateLimiter` (Redis already) | `receipt/router.go` |
| Public rate limit (run) | 10/min/IP + function call cost (paid functions) | `receipt/router.go` |
| Anti-scraping | Cloudflare Turnstile on `/run` (existing `captcha.CaptchaService`) | `handler.go:Run` |
| Privacy gate | Inherit existing `PrivacyService.SanitizeInputOutput` and `ShouldStoreInputOutput` (already on the write path) | `generateExecutionID` |
| PII redaction | Reuse `SanitizeInputOutput` (already in `handlers.go:576`) | unchanged |
| Owner can revoke | New `revoked_at TIMESTAMPTZ` column; UI button on `FunctionPage` | New column, `receipt_repository.go:Revoke` |

---

## 7. Frontend — Complete Build Spec

### 7.1 Routing (new + changed)

In `web/dashboard/src/App.tsx`, add (place after line 527 alongside the existing `/replay` route):

```tsx
<Route path="/r/:execId" element={<ReceiptPage />} />
<Route path="/r/:execId/embed" element={<ReceiptEmbedPage />} />
<Route path="/r/:execId/run" element={<ReceiptRunPage />} />
<Route path="/receipt/:execId" element={<ReceiptPage />} />   // alias for crawlers
```

**No auth gate.** The existing `routes.go` SPA catch-all already covers `/r/*` for unauthenticated GET — we only need to add the prefix to the matcher (see §5.5).

### 7.2 File tree (new)

```
web/dashboard/src/
├── pages/
│   └── ReceiptPage/
│       ├── index.tsx                    # <ReceiptPage> main component
│       ├── ReceiptPage.test.tsx
│       ├── components/
│       │   ├── ReceiptHeader.tsx        # function name, author avatar, runtime badge
│       │   ├── ReceiptSchemaViewer.tsx  # input/output JSON tree (custom)
│       │   ├── ReceiptInputOutput.tsx   # shows the recorded input + output
│       │   ├── ReceiptRunPanel.tsx      # "Run with this input" form (custom)
│       │   ├── ReceiptStats.tsx         # duration, cached, timestamp
│       │   ├── ReceiptPoweredBy.tsx     # subtle footer badge (custom)
│       │   ├── ReceiptShareBar.tsx      # copy link / tweet intent / QR
│       │   ├── ReceiptForkCTA.tsx       # "Deploy your own function — free" hero
│       │   └── ReceiptSkeleton.tsx      # loading state
│       ├── hooks/
│       │   ├── useReceipt.ts            # TanStack Query wrapper
│       │   ├── useReceiptRun.ts         # mutation: POST /v1/receipts/:id/run
│       │   └── useReceiptFork.ts        # builds the /functions/new?fork=... URL
│       ├── lib/
│       │   ├── og-meta.ts               # builds share text variants
│       │   ├── runtime-badge.ts         # color/icon per runtime
│       │   └── schema-render.ts         # JSON Schema → tree
│       └── types.ts
└── lib/
    └── receipt.ts                       # <meta> injection, OG image fallback
```

### 7.3 Custom React components (all new)

#### `<ReceiptPage>` — orchestrator
- Mounts `<Navbar variant="landing">` + `<Footer>` (same as `ReplayPage`).
- `useReceipt(id)` → `{ data, isLoading, error }`.
- Layout: 2-column on desktop (`md:grid-cols-3`), single column mobile.
- Renders: `<ReceiptHeader>`, `<ReceiptStats>`, `<ReceiptSchemaViewer>`, `<ReceiptInputOutput>`, `<ReceiptRunPanel>`, `<ReceiptForkCTA>`, `<ReceiptShareBar>`, `<ReceiptPoweredBy>`.
- On mount, fire-and-forget `POST /v1/receipts/:id/view` (analytics only).
- Injects OG meta tags via `react-helmet-async` (or whatever the codebase already uses — check `web/dashboard/src/App.tsx` for existing SEO helpers; if none, add `<Helmet>` from `react-helmet-async` already in `package.json`).

#### `<ReceiptHeader>` — custom
- Shows: function avatar, `name`, `author` (linked to `/u/:author`), `@version` badge, `runtime` badge, `verified` shield if `verification.status === "verified"`.
- Reuses `<Avatar>`, `<Badge>`, `<Button>` from `web/dashboard/src/components/ui/`.
- Author link: if owner has a profile, `<Link to="/u/${author}">`; else, plain text.

#### `<ReceiptSchemaViewer>` — custom, JSON-Schema-aware
- Tabs: "Input Schema" / "Output Schema".
- Renders the JSON Schema as a tree: `name: type` (required badge, default value, description).
- Uses shadcn `<Tabs>`, `<Collapsible>`, lucide icons (`ChevronRight`, `ChevronDown`, `Braces`).
- Falls back to pretty-printed raw JSON if no schema.
- Hover any leaf → shows inferred example.

#### `<ReceiptInputOutput>` — custom
- Two `<Card>`s side-by-side (collapses to tabs on mobile).
- Each renders the recorded value via `react-json-view` (or a tiny custom pretty-printer using the existing `<Textarea readOnly>` pattern from `ReplayPage`).
- Copy-to-clipboard button per card (uses existing `navigator.clipboard`).
- Truncates values > 8 KB with "show more" toggle.

#### `<ReceiptRunPanel>` — custom, the killer feature
- "Run with this input" header + CTA `<Button>`. Default = original input.
- Click → inline `<Textarea>` pre-filled with the original input (JSON), editable.
- Submit → `useReceiptRun` mutation → `POST /v1/receipts/:id/run`.
- On success: replace the output panel with the new run, show toast "New receipt: `/r/{newId}`" with a "Share" action.
- On failure: show the error inline; do not lose the original receipt.
- Disabled if `can_run === false` or `is_paid === true && !authed`. Paid functions show a modal: "This function is paid — sign in to run." (Routes to `/auth/login?next=/r/:id`.)
- Auth check uses the existing `useAuthStore` / `useAuth` hook in the dashboard.

#### `<ReceiptForkCTA>` — custom, the distribution lever
- Hero `<Card>` at the bottom: "Deploy your own function — free."
- Subtitle: "Fork this function into your own account in one click. No credit card required."
- Button label: "Fork this function" → calls `useReceiptFork()` → `window.location = ${APP_URL}/functions/new?fork=${base64(source)}&name=${name}&author=${author}`.
- If user is unauthenticated, route includes `/auth/signup?next=${urlencoded}` — the existing auth flow honors `?next=` (verify in `AuthPage`).
- Inline 3-icon row: "1. Fork" → "2. Edit" → "3. Deploy".

#### `<ReceiptPoweredBy>` — custom, the branding lever
- Small, fixed-bottom-right pill: "Powered by **FunctionFly**" with the FF wordmark.
- Whole pill is a link to `https://functionfly.com/?utm_source=receipt&utm_medium=virality`.
- Opacity 70% on idle → 100% on hover. NO animation (per spec: "subtle, not obnoxious").
- `<a href="..." target="_blank" rel="noopener noreferrer sponsored">` (sponsored rel for transparency).

#### `<ReceiptShareBar>` — custom
- Buttons: "Copy link" (`navigator.clipboard`), "Tweet" (opens `tweet_intent_url` from response in a new tab), "QR" (renders a QR via existing `qrcode` lib if installed, else shows a button that links to a static QR generator).
- Twitter intent URL is **server-computed** in the receipt response (no client-side URL building — easier to evolve copy and gives us analytics).

#### `<ReceiptStats>` — custom
- 3 metric cards (same pattern as `ReplayPage:144-181`): duration, cached/live, executed-at.
- Adds: a 4th card "Views: N" (from `view_count`) — social proof.

#### `<ReceiptSkeleton>` — custom
- Shimmer placeholders matching the real layout to avoid CLS.
- Use the existing `LoadingSpinner` for the spinner, or `tailwindcss-animate` if installed.

### 7.4 Styling / system alignment

- **Use the existing shadcn/ui primitives** (already in `components/ui/`): `Card`, `Button`, `Badge`, `Tabs`, `Textarea`, `Avatar`, `Tooltip`, `Skeleton`.
- **Tailwind** — already configured. Follow the same `bg-bg-primary` / `text-muted-foreground` tokens used in `ReplayPage`.
- **lucide-react** icons — already used everywhere.
- **No new dependencies.** If a package is missing for QR or JSON view, check `web/dashboard/package.json` first; if absent, prefer hand-rolled over adding deps in week 1.
- **Dark mode** — follow existing token-based approach (no hardcoded colors).
- **Responsive** — mobile-first; the run panel must work on a phone (most shares will be opened on mobile).

### 7.5 Embed variant (`/r/:id/embed`)

A minimal, iframe-safe variant:
- No `<Navbar>`, no `<Footer>`, no `<ReceiptPoweredBy>`.
- Just `<ReceiptHeader>`, `<ReceiptSchemaViewer>`, `<ReceiptInputOutput>`, `<ReceiptRunPanel>` in a 380px-wide card.
- The existing embed service (`internal/api/handlers/registry/embed.go` + `web/dashboard/src/components/embed/`) is the natural extension point — follow its `X-Embed-Origin` / `RateLimitPerHour` pattern.

### 7.6 `useReceipt` hook (TanStack Query)

```ts
// web/dashboard/src/pages/ReceiptPage/hooks/useReceipt.ts
export function useReceipt(execId: string) {
  return useQuery<Receipt, ApiError>({
    queryKey: ['receipt', execId],
    queryFn: () => apiGet(API_URLS.receipt.get(execId)),
    enabled: !!execId,
    staleTime: 30_000,           // match CDN edge TTL
    gcTime: 5 * 60_000,
    retry: (count, err) => err.status !== 404 && count < 2,
  });
}
```

Add to `web/dashboard/src/lib/api-urls.ts`:

```ts
receipt: {
  get:    (id: string) => `${API}/receipts/${id}`,
  run:    (id: string) => `${API}/receipts/${id}/run`,
  fork:   (id: string) => `${API}/receipts/${id}/fork-payload`,
  view:   (id: string) => `${API}/receipts/${id}/view`,
  trending: ()          => `${API}/receipts/trending`,
},
```

### 7.7 SEO + sharing

- **Server-rendered OG image:** the existing `web/site` (Astro, port 4321) can serve a static OG image at `/og/receipt/:id.png` generated at deploy time or by a tiny Astro endpoint. Vercel `og` library is already in the stack (or use a simple Satori-based renderer). Fallback to a generic FunctionFly OG if the dynamic one fails.
- **`<title>` and meta tags** injected client-side via `react-helmet-async` (or whatever the codebase uses) using `share.og_meta` from the response.
- **`robots.txt`:** allow `/r/*`, disallow `/r/*/run` (POST pages aren't crawlable anyway but explicit is safer).
- **Sitemap:** add `/r/:id` URLs for all `shareable = true` rows from the last 7 days, regenerated daily by a small Asynq job.

### 7.8 Tests

- **Unit:** `<ReceiptPage>` renders, calls hook, handles 404/410.
- **Unit:** `useReceiptRun` mutation: success → toast + state update, failure → inline error.
- **Unit:** `ReceiptForkCTA` builds correct `/functions/new?fork=...` URL.
- **Unit:** `ReceiptPoweredBy` has the correct href, `rel="noopener noreferrer sponsored"`, and opacity transitions.
- **E2E (Playwright):** open `/r/:id` (created via API), click "Run with this input", assert new `execution_id` returned and a new `/r/:id` URL appears.
- **E2E:** unauthenticated user clicks "Fork this function" → lands on `/auth/signup?next=/functions/new?fork=...`.
- **E2E (Playwright):** social share meta tags present in the head.

---

## 8. Milestone Worker

### When to fire
- After `RecordExecution` succeeds in the realtime usage tracker.
- Thresholds: **1, 10, 100, 1000, 10000** (configurable via `RECEIPT_MILESTONE_THRESHOLDS` env var, comma-separated).

### What to send

| Channel | Implementation | Cost |
|---|---|---|
| In-app notification | Insert into `notifications` table (existing) — toast on next dashboard load | Free |
| Email | `notificationSvc.SendEmail` via Resend (existing) — "Your function just ran 100× — here's the public receipt" | Free tier |
| Tweet intent | **NO auto-tweet.** Provide a `tweet_intent_url` in the notification payload. The owner clicks → pre-filled tweet → they tweet. Zero OAuth, zero liability. | Free |
| Slack/Discord webhook | Out of scope week 1; row in `receipt_milestone_events.channels_fired` is extensible | — |

### Idempotency
- `receipt_milestone_events.dedupe_key = function_id:threshold` is `UNIQUE`. The hook is `INSERT ... ON CONFLICT DO NOTHING`. Retries are safe.

### Owner opt-out
- New column `users.receipt_milestones_enabled BOOLEAN DEFAULT TRUE` in a follow-up migration. Respect it in the worker.

### Configuration env vars

| Var | Default | Purpose |
|---|---|---|
| `RECEIPT_ENABLED` | `true` | Master kill switch |
| `RECEIPT_MILESTONE_THRESHOLDS` | `1,10,100,1000,10000` | CSV of run-count thresholds |
| `RECEIPT_MILESTONE_CHANNELS` | `inapp,email` | Which channels to fire |
| `RECEIPT_PUBLIC_BASE_URL` | `https://functionfly.com/r` | Public receipt URL prefix |
| `RECEIPT_TWITTER_HANDLE` | `functionfly` | For `@functionfly` mention in tweet text |
| `RECEIPT_OG_BASE_URL` | `https://functionfly.com/og/receipt` | OG image base |

---

## 9. Rollout & Feature Flags

| Flag | Default | Effect |
|---|---|---|
| `RECEIPT_ENABLED` | `true` in dev, `false` in prod (week 1) | Hard kill switch |
| `RECEIPT_AUTO_GENERATE` | `true` | If `false`, receipts are only generated for `SideEffects == "none"` (legacy behavior) — start with this; flip when confident |
| `RECEIPT_MILESTONE_ENABLED` | `false` in prod week 1 | Don't spam owners while validating |
| `RECEIPT_PII_REDACT_STRICT` | `true` | If `true`, force `SanitizeInputOutput` even if privacy service is disabled |

**Week 1 rollout:**
1. Ship the migration, the handler, and the page behind `RECEIPT_ENABLED=false`.
2. Manually create a receipt for an internal function, screenshot it, share with team.
3. Enable for staging → `RECEIPT_AUTO_GENERATE=true`.
4. Enable milestone worker in staging → verify email + in-app.
5. Enable in prod with `RECEIPT_MILESTONE_ENABLED=false` for 48h.
6. Flip the milestone flag.

---

## 10. Observability

- **Metrics** (Prometheus, existing `/metrics`): `receipt_get_total`, `receipt_run_total{status}`, `receipt_fork_total`, `receipt_view_total`, `receipt_milestone_fired_total{threshold,channel}`.
- **Logs** (existing logrus): one structured log per receipt hit with `receipt_id`, `function_id`, `cache_layer`, `view_count`.
- **Dashboard** (Grafana, existing): add a "Receipt funnel" panel — views → runs → forks → signups. The last hop (signups attributed to a referral cookie `?r={id}`) is the conversion metric for the whole feature.
- **Alerts**: page if `receipt_run_total{status="5xx"}` rate > 1% for 5m.

---

## 11. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Public pages expose function internals (input data leakage) | Existing `PrivacyService.SanitizeInputOutput` + `ShouldStoreInputOutput` already gate this; owner can revoke per-row via new `revoked_at` |
| Public pages become an attack surface (run with arbitrary input = $$ cost on paid functions) | Public `run` is **only enabled for `visibility == "public"` and `PricePerCall == 0`** for week 1; paid runs go through auth + wallet (existing). Add Turnstile on `run` if abuse appears. |
| `/r/:id` becomes a hot path → Postgres melt | Redis L1 (60s), CDN L0 (5m), partial index, denormalized row, read replica |
| Owner gets spammed by milestone emails | `RECEIPT_MILESTONE_ENABLED=false` for first 48h, opt-out column, configurable thresholds |
| Twitter intent URL leaks API design | All copy is server-rendered in the receipt response; copy changes are config deploys, not code deploys |
| Existing `ReplayPage` users get a worse experience | Keep `/replay/:id` as a redirect to `/r/:id`. ReplayPage is the seed, ReceiptPage is the upgrade. |
| SEO: Google indexes thousands of `/r/:id` URLs → crawl budget | `noindex` for runs with `< 10ms` duration (probably tests); `last_viewed_at` index for the sitemap (only include last-7-day high-view-count receipts) |

---

## 12. Day-1 / Day-7 Delivery Plan

### Day 1-2 (Backend skeleton)
- [ ] Create migration `20260601180000_execution_receipt` (up + down) via `./scripts/create-migration.sh`
- [ ] Apply migration locally
- [ ] Add `internal/storage/receipt/receipt_repository.go` (`Get`, `IncrementViewCount`, `GetTrending`, `Revoke`)
- [ ] Add `internal/api/handlers/receipt/handler.go` with stub methods
- [ ] Wire `registerReceiptRoutes` into `routes.go`; add `r` and `receipt` to the SPA matcher
- [ ] Refactor `HandleExecute` → extract `executeCore()` so both `/v1/fx/...` and `/v1/receipts/:id/run` use the same path
- [ ] Hook backfill + milestone check in `HandleExecute` behind `RECEIPT_ENABLED` env flag

### Day 3-4 (Frontend)
- [ ] Add routes `/r/:id`, `/r/:id/embed`, `/r/:id/run` to `App.tsx`
- [ ] Add `receipt` to `api-urls.ts`
- [ ] Build `<ReceiptPage>` + all 9 subcomponents (using `shadcn/ui` primitives)
- [ ] Add `useReceipt`, `useReceiptRun`, `useReceiptFork` hooks
- [ ] Add `react-helmet-async` OG meta injection (or whatever is in use)
- [ ] Style pass + dark mode + mobile responsive
- [ ] Add `ReceiptPoweredBy` subtle badge

### Day 5 (Milestones)
- [ ] Add `internal/scheduler/receipt_milestone_scheduler.go` (real-time hook + daily sweep)
- [ ] Add `internal/api/handlers/receipt/milestone.go`
- [ ] Build the email template "Your function just ran N times"
- [ ] Add `users.receipt_milestones_enabled` column (follow-up migration if scope creep)
- [ ] Wire tweet intent URL into the notification payload

### Day 6 (Embed + SEO + tests)
- [ ] Build `<ReceiptEmbedPage>` (`/r/:id/embed`)
- [ ] Generate dynamic OG images (Astro endpoint or Vercel `og` lib)
- [ ] Add `/r/:id` to `robots.txt` and sitemap
- [ ] Unit + E2E tests (Playwright)
- [ ] Lighthouse pass on `/r/:id` (target: >95 perf, >90 SEO)

### Day 7 (Rollout)
- [ ] Behind `RECEIPT_ENABLED=false` in prod → smoke test
- [ ] Enable `RECEIPT_AUTO_GENERATE=true`
- [ ] Monitor for 48h → enable `RECEIPT_MILESTONE_ENABLED=true`
- [ ] Add a "Trending receipts" widget to `web/site` (Astro) homepage
- [ ] Launch tweet

---

## 13. Out of Scope (Week 1)

- Server-side rendered `/r/:id` (Astro) — current plan keeps the React SPA page. Revisit in week 2 if Lighthouse < 90 on mobile.
- Auto-generated Open Graph images via a worker (use a static image week 1).
- Owner dashboard "All my receipts" view — wait for analytics from week 1 to confirm product-market fit.
- Slack/Discord webhook for milestones.
- "Remix" CTA (uses the receipt as a template, not a fork) — too overlapping with Fork in v1.

---

## 14. Open Questions (none blocking)

The plan is implementation-ready. The only judgment calls that benefit from product input are in §11 (risks) and §7.5 (embed), and both have safe defaults. No new architecture decisions are required from the user.

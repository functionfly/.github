# Two-Protocol Gateway Plan

FunctionFly is the load-bearing gateway for both **MCP** (vertical: agent → tools/data) and **A2A** (horizontal: agent → agent). One endpoint, two protocol adapters, one receipt spine, one registry.

> **DRY principle:** Both protocol adapters are thin shells over a single `GatewayCore`. The protocol layer is plumbing; the value lives in the auth, capability routing, rate-limit brokering, fallback chain, and the signed Receipt. Build the core once. An **A2A Task IS a long-lived receipt** — same table, same HMAC, same revocation path. There is no `a2a_tasks` table.

---

## 1. Current state (audit)

### 1.1 MCP — **production-ready** ✅

| Surface | File | Status |
|---|---|---|
| JSON-RPC 2.0 transport | `internal/api/handlers/mcp/jsonrpc.go` | ✅ spec-compliant (`tools/list`, `tools/call`, `initialize`, `ping`, batch) |
| `POST /v1/mcp` (streamable-HTTP) | `internal/api/routes_mcp.go:65` | ✅ |
| `GET /v1/mcp/manifest` | `internal/api/handlers/mcp/handler.go:139` | ✅ server identity, cacheable |
| `GET /v1/mcp/tools` | `internal/api/handlers/mcp/handler.go:187` | ✅ public tool index, SEO |
| `tools/call` auth gating | `internal/api/handlers/mcp/auth.go:135` | ✅ bearer (JWT + `ff*_` API keys) |
| Per-function allowlist + rate limit | `registry/mcp.go:282` | ✅ |
| MCP invocation telemetry | `registry/mcp.go` `RecordMCPInvocation` | ✅ |
| `tools/call` executor | `internal/api/handlers/mcp/tools_call.go:18` | ✅ wires to existing registry executor |
| Tests | `internal/api/handlers/mcp/mcp_test.go` | ✅ |
| Sitemap + OG | `mcp/sitemap.go`, `mcp/og.go` | ✅ |

### 1.2 A2A — **does not exist** ❌

Grep results across all `.go`:
- `a2a|agent_card|AgentCard|tasks/send|tasks/get|/.well-known/agent` → **0 matches**.
- The "agent-to-agent" code in `internal/agent/swarm/messages.go:59` is a **private** HMAC signing scheme for FunctionFly's own swarm, not the A2A standard.
- No Agent Card discovery, no Task lifecycle, no `/.well-known/agent.json`, no `tasks/send`, `tasks/get`, `tasks/cancel`, no JSON-RPC-over-HTTP for A2A, no peer-agent routing.

**This is the gap.** Everything else in this plan is the foundation A2A should plug into.

### 1.3 Execution Receipts — **production-ready, MCP-only** ✅⚠️

| Surface | File | Status |
|---|---|---|
| `GET /v1/receipts/{id}` | `handlers/receipt/handler.go:324` | ✅ public, HMAC-signed IDs, cacheable |
| `POST /v1/receipts/{id}/run` | `handlers/receipt/handler.go:376` | ✅ re-executes via `/v1/fx/.../execute` |
| `GET /v1/receipts/{id}/fork-payload` | `handlers/receipt/handler.go:473` | ✅ |
| `POST /v1/receipts/{id}/revoke` (owner) | `handlers/receipt/handler.go:643` | ✅ |
| Trending + per-function list | `handler.go:556`, `:587` | ✅ |
| HMAC signing | `handler.go:936` `HMACSign`/`HMACVerify` | ✅ |
| Milestone worker | `handlers/receipt/milestone.go` + `scheduler/receipt_milestone_scheduler.go` | ✅ |
| **DRE on-chain anchoring** | `internal/dre/cert/anchoring_eth.go` | ✅ Ethereum anchor |
| Share OG + tweet intent | `handler.go:837` | ✅ |
| Storage | `internal/storage/receipt/receipt_repository.go` | ✅ |

⚠️ **Limitation:** Source-of-truth is `registry_executions_public` (function execution only). A2A task executions will be **the same table** with a `protocol` enum + a `state` column for the A2A state machine. See §5.

### 1.4 Layer 2 (Gateway) — **partly there** ⚠️

| Capability | File | Status |
|---|---|---|
| Auth proxy (JWT + API keys) | `internal/auth/`, `middleware/auth.go` | ✅ |
| Per-scope rate limiters | `middleware/*_rate_limit.go` | ✅ (auth/vault/provider/wallet/mfa/public/agent/admin) |
| Quota middleware | `middleware/quota_middleware.go` | ✅ |
| Verification middleware (ClamAV+YARA+trust) | `middleware/verification.go` | ✅ |
| Execution coordinator | `middleware/execution_coordinator.go` | ✅ |
| Anti-loop | `middleware/antiloop.go` | ✅ |
| **Capability-based routing** | – | ❌ not a first-class concept |
| **Fallback chains** | `registryexecution.RuntimeRouter` is per-function runtime, not gateway | ⚠️ partial |
| **A2A brokering** | – | ❌ |

### 1.5 Layer 4 (Registry & Discovery) — **strong for MCP, A2A-shaped hole** ⚠️

| Capability | File | Status |
|---|---|---|
| MCP Server Registry | `internal/api/handlers/registry/mcp_settings.go`, `storage/registry/mcp.go` | ✅ trust score, version, transport, allowlist |
| Trust scores, ratings, reviews | `registry/trust.go`, `registry/reviews.go` | ✅ |
| Bundles, versions, canary | `registry/version_management.go`, `canary.go` | ✅ |
| `/.well-known/functionfly.json` | `handlers/wellknown/handler.go:94` | ✅ LLM/agent discovery (FunctionFly-native) |
| **A2A Agent Card Directory** | – | ❌ |
| **`/.well-known/agent.json`** | – | ❌ |
| **Protocol negotiation in well-known** | – | ❌ no `supported_protocols` block |

### 1.6 Frontend (dashboard)

| Surface | Path | Status |
|---|---|---|
| MCP tool browser | implicit via registry UI | ⚠️ no first-class "MCP server registry" page |
| Receipts page | `web/dashboard/src/pages/ReceiptPage` | ✅ |
| A2A agent browser | – | ❌ |

### 1.7 Existing code smells to fix in P0 (DRY violations)

These are concrete duplications/crossings the GatewayCore refactor must consolidate, not paper over:

| Smell | Locations | Move to |
|---|---|---|
| `HMACSign`/`HMACVerify` defined as **method** on `receipt.Handler` (cannot be reused by other packages) | `internal/api/handlers/receipt/handler.go:936` | `internal/gateway/receipt.go` as **package-level functions** |
| `setCORSHeaders` duplicated | `internal/api/handlers/mcp/handler.go:700` and `internal/api/handlers/wellknown/handler.go:242` | `internal/gateway/cors.go` |
| DRE anchoring + milestone worker logic is called only from receipt handler, but the *trigger* is "an execution finished" — protocol-agnostic | `internal/dre/cert/anchoring_eth.go` + `scheduler/receipt_milestone_scheduler.go` | `internal/gateway/receipt/` (public API: `Emit(req, result)`) |
| The `Executor` interface in MCP is a near-perfect `GatewayCore` half — but lives inside the MCP package, blocking reuse | `internal/api/handlers/mcp/handler.go:54` | `internal/gateway/core.go` (`CallRequest`/`CallResult`) |

**P0 ships when these four are gone.**

---

## 2. The moat, in one sentence

> **FunctionFly is the only gateway that authenticates once, enforces quota + verification once, routes once, executes once, and emits one signed, optionally on-chain-anchored receipt for both protocol paths — and a tool call can hand off to a peer agent without the caller knowing.**

The "without the caller knowing" is the load-bearing part. See §4.6.

---

## 3. The DRY core: `internal/gateway/`

Extract a `GatewayCore` from the existing `mcp/handler.go:496 dispatch` and the `registry/handlers.go HandleExecute` so both MCP and A2A call the same code path. **No copy-paste.**

```
internal/gateway/
├── core.go              # GatewayCore: the single execution path (CallRequest → CallResult)
├── capability.go        # CapabilitySet + data-driven capability resolution
├── ratelimit.go         # shared quota/rate-limit brokering
├── fallback.go          # chain definition + health-aware failover
├── receipt.go           # package-level HMACSign/HMACVerify (was: receipt.Handler methods)
├── receipt/
│   ├── emit.go          # Emit(req, result) — single entry point
│   ├── milestone.go     # moved from internal/api/handlers/receipt/milestone.go
│   └── anchor.go        # wraps internal/dre/cert/anchoring_eth.go
├── auth.go              # identity resolution (JWT | API key | agent signing key | A2A peer JWT)
├── cors.go              # shared CORS (was: duplicated in mcp + wellknown)
├── version.go           # supported protocol versions (negotiation)
└── adapters/
    ├── mcp.go           # thin adapter on top of Core
    └── a2a.go           # thin adapter on top of Core
```

### 3.1 `GatewayCore` interface (the only thing adapters need)

```go
// internal/gateway/core.go
type CallRequest struct {
    Protocol     Protocol              // ProtocolMCP | ProtocolA2A
    Caller       Caller                // resolved identity
    Target       Target                // Resolved: (function) | (agent)
    Inputs       json.RawMessage
    Session      *SessionCtx           // call-depth, parent-task, trace
    Quota        QuotaHint
    Capabilities []string              // requested capabilities
    Metadata     map[string]string     // protocol-specific passthrough
}

type CallResult struct {
    Output        json.RawMessage
    Status        int
    DurationMs    int
    ReceiptID     string               // always populated (nanoid, HMAC-signed)
    ReceiptSig    string               // HMAC
    AnchoredTx    string               // empty if not anchored
    FallbackChain []string             // which adapter/route served it
    Cached        bool                 // true if served from cache hit
    State         string               // for A2A: "submitted"|"working"|"input-required"|"completed"|"failed"|"canceled"
}
```

Both `mcp.Handler` (already has `Executor` interface — `mcp/handler.go:54`) and the future `a2a.Handler` will pass through `GatewayCore.Call(ctx, CallRequest)`. **The Executor interface in MCP is already 80% of this; finish the extraction.**

### 3.2 Capability-based routing (data-driven, not code)

`internal/gateway/capability.go`:
- Capabilities live in **DB tables** (`api_key_capabilities`, `agent_capabilities`, `role_capabilities`) loaded at request time.
- No Go switch statements. Adding a new capability is a row, not a deploy.
- Core capability namespaces:
  - `mcp:tools:list`, `mcp:tools:call`
  - `a2a:tasks:send`, `a2a:tasks:get`, `a2a:tasks:cancel`, `a2a:tasks:subscribe`
  - `a2a:agent:invoke` (peer-to-peer)
  - `a2a:agent:delegate` (hand off a tool result to a peer — see §4.6)
  - `receipt:read`, `receipt:write`, `receipt:revoke`
  - `bundle:publish`, `bundle:install`

### 3.3 Fallback chains

Today: `registryexecution.RuntimeRouter` chooses per-function runtime (WASM → CPython → sandbox).  
Tomorrow: `GatewayCore` also chooses **per-call**:
- Primary: in-region FunctionFly worker
- Fallback 1: peer agent via A2A (if MCP target is unhealthy)
- Fallback 2: cached output (if input hash matches and TTL valid)
- Fallback 3: degraded mode (return last known good)

The chain decision is logged in `Receipt.FallbackChain` **and** in the metric `gateway_fallback_fired_total{from,to,reason}` (see §10) so the SRE team can alert on degradation.

### 3.4 Auth (one resolver, multiple credential types)

`gateway/auth.go` resolves a `Caller` from any of:
- JWT (existing) — user
- `ff*_` API key (existing) — function caller
- `X-Agent-API-Key` (existing in `agent/execute.go:307`) — agent
- **A2A peer JWT** (new) — peer agent from another framework, with JWKS URL on their `agent_card`

The rest of the pipeline is identical.

### 3.5 Protocol version negotiation (NEW)

`internal/gateway/version.go` exposes the versions this gateway speaks:

```go
var SupportedProtocols = ProtocolSet{
    "mcp": semver.MustParse("2025-03-26"),
    "a2a": semver.MustParse("0.3.0"),
}
```

Surfaced in:
- `GET /.well-known/functionfly.json` (extended — see §6.2)
- `GET /.well-known/agent.json` (A2A spec, see §4.2)
- The `ServerManifest` returned by `GET /v1/mcp/manifest`

A2A is still moving (ACP merged late 2025). Pin the version, then negotiate. Without this, version-drift bugs will hit production.

---

## 4. A2A build-out (the missing half)

### 4.1 Spec surface to implement

A2A core methods (per Google's 2025 spec, now LF-governed with ACP merged in):

| Method | Path | Auth |
|---|---|---|
| `GET /.well-known/agent.json` | public | none |
| `POST /v1/a2a/{agent_id}/tasks/send` | send a task | bearer |
| `GET /v1/a2a/tasks/{task_id}` | poll task | bearer |
| `POST /v1/a2a/tasks/{task_id}/cancel` | cancel | bearer |
| `POST /v1/a2a/tasks/{task_id}/subscribe` | SSE stream | bearer |
| `POST /v1/a2a/{agent_id}/message/send` | short message | bearer |

### 4.2 Agent Card shape (server-side, served by FunctionFly)

```json
{
  "name": "functionfly",
  "description": "FunctionFly agent gateway",
  "url": "https://api.functionfly.com/v1/a2a",
  "version": "1.0",
  "protocolVersion": "0.3.0",
  "capabilities": ["streaming", "push-notifications"],
  "skills": [
    { "id": "execute", "description": "Run any registered function" },
    { "id": "delegate", "description": "Hand off a tool result to a peer agent" }
  ],
  "authentication": { "schemes": ["bearer", "agent-apikey"] },
  "defaultInputModes": ["application/json"],
  "defaultOutputModes": ["application/json"]
}
```

Plus per-agent cards at `GET /v1/a2a/agents/{agent_id}/card` (sourced from the existing `agent/identity` repo + `agent/tools` registry).

### 4.3 A2A → GatewayCore mapping (Task = Receipt, no new table)

> **Decision:** An A2A Task is a long-lived Receipt with a state machine. We do **not** create an `a2a_tasks` table. This collapses two storage layers into one and gives A2A tasks the same HMAC-signed share URLs, trending, milestones, and revocation for free.

| A2A concept | FunctionFly internals |
|---|---|
| Agent Card | `agent/identity` + a new `agent_card` view |
| **Task** | **row in `registry_executions_public` with `protocol='a2a'`** + a new `state` column |
| Task state machine | `a2a.TaskEngine` updates the same row's `state` column; the row's `output` column fills on completion |
| Task identity (public) | the `public_id` (nanoid) — same shape as MCP receipts, single `isValidPublicID` check |
| Task identity (internal) | UUID primary key, same as MCP |
| Message parts | JSON or text; map to `Inputs` |
| Artifact | `Output` (stored on the same row) |
| Push notifications | reuse `notification.Service` |
| **Revocation** | **same `/v1/receipts/{id}/revoke` endpoint** — works for both protocols |
| **Trending** | **same `/v1/receipts/trending`** — A2A receipts appear alongside MCP |

### 4.4 Files to create

```
internal/a2a/
├── handler.go              # HTTP handlers for A2A paths (thin)
├── card.go                 # Agent Card model + serializer
├── task_engine.go          # State machine — updates registry_executions_public.state
├── stream.go               # SSE for tasks/subscribe
└── storage/
    └── a2a_card_repository.go   # only for agent_cards (Task storage REUSES registry)

internal/gateway/
└── adapters/
    └── a2a.go              # translates A2A requests into GatewayCore.Call

internal/api/routes_a2a.go
```

Note the **absence** of `a2a_task_repository.go` and `a2a/storage/`. The Task storage is the existing `registry.RegistryRepository` — no duplication.

### 4.5 Wiring (DRY)

```go
// internal/api/routes_a2a.go (sketch — DO NOT WRITE YET)
core := gateway.NewCore(gateway.Deps{
    Registry: registryRepo, Auth: authSvc, Quota: quotaMW,
    Receipt: receiptSvc, Fallback: fallbackChain, Versions: gateway.SupportedProtocols,
})
api.HandleFunc("/.well-known/agent.json", a2aCardHandler.ServeGatewayCard)
api.HandleFunc("/v1/a2a/agents/{agent_id}/card", a2aCardHandler.ServeAgentCard)
a2aHandler := a2a.NewHandler(core, a2aRepo, agentIdentityRepo, notificationSvc)
api.HandleFunc("/v1/a2a/{agent_id}/tasks/send", a2aHandler.SendTask)
api.HandleFunc("/v1/a2a/tasks/{task_id}", a2aHandler.GetTask)
api.HandleFunc("/v1/a2a/tasks/{task_id}/cancel", a2aHandler.CancelTask)
api.HandleFunc("/v1/a2a/tasks/{task_id}/subscribe", a2aHandler.SubscribeSSE)
```

`a2a.Handler` is **only** A2A plumbing (JSON shaping, task state transitions, SSE). It calls `core.Call(ctx, CallRequest{Protocol: ProtocolA2A, ...})` for the actual execution and delegates state updates to `a2a.TaskEngine` (which writes to `registry_executions_public.state`). **Zero execution logic in the A2A package.**

### 4.6 The killer use case: `tools/call` → delegate to peer

> **The "load-bearing for both protocols" claim becomes concrete here.** An MCP caller invokes a tool; the tool's result is a *handoff* to a peer A2A agent. The caller never knows.

**Flow:**

1. Peer agent `A` calls MCP `tools/call` on `fx://functionfly/summarize_doc` with `{doc_url, ask_peer: true, peer: "research-agent@other-framework"}`.
2. `GatewayCore.Call(req)` runs the function. The function returns a normal MCP result.
3. If the function's result includes a `delegate` block (a new optional return shape), `GatewayCore`:
   - Creates a **second receipt** (also via `Emit`) with `protocol='a2a'`, `parent_task_id=<first_receipt_id>`, `state='submitted'`.
   - Calls `a2a.TaskEngine.Send(req)` to dispatch to the peer.
   - Returns to the original MCP caller a `CallResult` that links to the A2A receipt.
4. The MCP caller can now poll the A2A receipt (`/v1/receipts/{id}` with `?protocol=a2a`) — or subscribe via SSE.

**What this needs:**

- New capability `a2a:agent:delegate` (added to the capability table — no code change).
- New optional return shape on function outputs: `{ "data": <normal>, "delegate": { "peer_card_id": "...", "input": {...} } }`.
- The `delegate` block is parsed by `GatewayCore` after the function returns; if present, `TaskEngine.Send` is invoked before `CallResult` is returned to the caller.
- Both receipts (the MCP one and the A2A one) appear in `/v1/receipts/{author}/{name}` trending — they share a `function_id` (the MCP function) and a `parent_task_id` (the A2A task).

**This is Phase P2.5** — it ships between the A2A spec scaffolding (P2) and the full A2A task lifecycle (P3), and it's the demo we lead with.

---

## 5. Cross-protocol Receipts

### 5.1 Schema change (idempotent migration)

```sql
-- Add a protocol column to the public-execution source-of-truth.
-- A2A Tasks live in the SAME table as MCP receipts. The protocol column
-- is the discriminator; the state column holds the A2A state machine.

ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS protocol TEXT NOT NULL DEFAULT 'mcp'
    CHECK (protocol IN ('mcp', 'a2a'));

ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS state TEXT NOT NULL DEFAULT 'completed'
    CHECK (state IN (
      'submitted', 'working', 'input-required',
      'completed', 'failed', 'canceled'
    ));

ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS parent_task_id UUID NULL
    REFERENCES registry_executions_public(public_id) ON DELETE SET NULL;

ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS fallback_chain TEXT[] NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_exec_public_protocol
  ON registry_executions_public(protocol, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exec_public_state
  ON registry_executions_public(protocol, state)
  WHERE state IN ('submitted', 'working', 'input-required');

CREATE INDEX IF NOT EXISTS idx_exec_public_parent
  ON registry_executions_public(parent_task_id)
  WHERE parent_task_id IS NOT NULL;
```

**No `a2a_tasks` table.** This was the headline DRY win.

### 5.2 Receipt emission in `GatewayCore`

`internal/gateway/receipt/emit.go Emit(req, result)`:
1. Generate `public_id` (nanoid) — same algorithm as `receipt.IsValidPublicID` (`receipt/handler.go:861`).
2. Compute `receipt_sig` via package-level `gateway.HMACSign(public_id)` (moved from `receipt.Handler.HMACSign`).
3. Insert `registry_executions_public` with `protocol`, `state`, `parent_task_id` (if delegation), `fallback_chain`, `function_id` (nullable for A2A peer-to-peer), `input`, `output`, `duration_ms`, `cached`, `shareable`, `verification_status`.
4. Async: `gateway/receipt/milestone.go` (moved from `internal/api/handlers/receipt/milestone.go`).
5. Async: `gateway/receipt/anchor.go` (wraps `internal/dre/cert/anchoring_eth.go`) if `protocol ∈ {mcp, a2a}` and `DRE_ANCHORING_ENABLED=true`.

`receipt.Handler.HandleGet` (existing) auto-serves both — no API surface change for the public reader. The response gains a `protocol` and `state` field (cosmetic).

### 5.3 Receipt URL convention

- MCP: `https://functionfly.com/r/{id}` (unchanged)
- A2A:  `https://functionfly.com/r/{id}?protocol=a2a&task={id}` (same URL, query disambiguates; `?task=` is a no-op for now but reserved for future deep-linking into the task timeline)

Single canonical URL, one share link, one trending surface, one revocation path.

### 5.4 Files to touch

- `internal/storage/registry/registry_models.go` — add `Protocol`, `State`, `ParentTaskID`, `FallbackChain` to `RegistryExecutionPublic`.
- `internal/api/handlers/receipt/handler.go` — surface `protocol`, `state`, `parent_task_id` in `PublicResponse`. **Note:** `HMACSign`/`HMACVerify` are removed from this handler (now in `internal/gateway/receipt.go`).
- New migration under `migrations/YYYYMMDDHHMMSS_receipt_protocol_state_columns.up.sql`.
- New `internal/gateway/receipt/emit.go` (the new emitter).
- Move `internal/api/handlers/receipt/milestone.go` → `internal/gateway/receipt/milestone.go` and update imports.
- Move DRE-anchoring trigger logic from receipt handler to `internal/gateway/receipt/anchor.go`.

---

## 6. Registry & Discovery additions

### 6.1 Agent Card Directory

A registry **of** agent cards, separate from the function registry. Mirrors the MCP Server Registry structure:

| Endpoint | Purpose |
|---|---|
| `GET /v1/agents/cards` | Browse agent cards (paginated, filter by capability/skill) |
| `GET /v1/agents/cards/{agent_id}` | Single card |
| `POST /v1/agents/cards` | Publish/update card (owner) |
| `DELETE /v1/agents/cards/{agent_id}` | Unpublish |
| `GET /v1/agents/cards/{agent_id}/versions` | Version history |

Data model: `agent_cards(agent_id, version, name, description, skills JSONB, capabilities TEXT[], url, auth_schemes TEXT[], input_modes TEXT[], output_modes TEXT[], trust_score, published_at, peer_jwks_url, ...)` — parallel to `mcp_settings` (`storage/registry/mcp.go`).

### 6.2 `/.well-known/agent.json` (A2A discovery) + protocol negotiation

Served at root by the existing `wellknown.Handler` pattern (`wellknown/handler.go:94`). Returns FunctionFly's own gateway card. Per-agent cards are at `/.well-known/{agent_id}/agent.json` (lower priority — A2A spec primarily expects the gateway card).

**Extended `/.well-known/functionfly.json`** — add a `supported_protocols` block so callers can feature-detect:

```json
{
  "schema_version": "1.0",
  "provider": "functionfly",
  "supported_protocols": {
    "mcp": "2025-03-26",
    "a2a": "0.3.0"
  },
  "endpoints": {
    "mcp_streamable_http": "https://api.functionfly.com/v1/mcp",
    "a2a_tasks_send":      "https://api.functionfly.com/v1/a2a/{agent_id}/tasks/send",
    "a2a_agent_card":      "https://api.functionfly.com/v1/a2a/agents/{agent_id}/card",
    "well_known_agent":    "https://api.functionfly.com/.well-known/agent.json",
    "receipts_public":     "https://api.functionfly.com/v1/receipts/{id}"
  }
}
```

Sourced from `internal/gateway/version.go` — single source of truth.

### 6.3 Bundles (curated agent + function packs) — moved to **P5**

Reuse `internal/bundler/`. Add `bundle_kind = 'mcp+agent'` for cross-protocol bundles. The bundle's `bundle.yaml` lists both `mcp_servers` and `agent_cards`. **This is the only artifact that ties the two protocol stories into a single shippable thing for partners** — it's the marketable demo and the partner onboarding primitive. Promoting it to P5 (right after the load-bearing core work) is deliberate.

---

## 7. Dashboard surfaces (frontend)

### 7.1 New pages

- `web/dashboard/src/pages/MCPRegistryPage.tsx` — first-class browser for MCP servers, sourced from `/v1/mcp/manifest` + `/v1/mcp/tools`. Sitemap, OG previews, per-tool trust score.
- `web/dashboard/src/pages/AgentCardsPage.tsx` — browse agent cards, see trust score, install.
- `web/dashboard/src/pages/A2AExplorerPage.tsx` — A2A playground: send a task, watch lifecycle (uses `<TaskTimeline />`), view task, view receipt.
- `web/dashboard/src/pages/BundlesPage.tsx` — browse `mcp+agent` bundles, one-click install.

### 7.2 Shared components

- `<ProtocolBadge protocol="mcp|a2a" />` — used everywhere a call appears.
- `<ReceiptCard />` (existing) — extend with `protocol` and `state` fields. Same component renders both MCP receipts and A2A tasks.
- `<TaskTimeline />` — A2A task lifecycle viewer (`submitted → working → input-required → completed` / `failed` / `canceled`). Receives a `ReceiptCard` and overlays the state transitions.
- `<CapabilityPicker />` — for the bundle install + agent card publish flows; loads from `GET /v1/capabilities`.

---

## 8. Phased rollout

| Phase | Scope | Done-when | Risk |
|---|---|---|---|
| **P0** | Extract `internal/gateway/core.go` from `mcp.Handler.dispatch` + `registry.Handler.HandleExecute`. **Also fix the four DRY violations in §1.7** (move HMAC, CORS, DRE+milestone into `gateway/`). MCP keeps using the new Core. | MCP `tools/call` paths pass `mcp_test.go` AND the **contract test replay** (§8.1) green. | Low — pure refactor, contract test is the safety net. |
| **P0.5** | Contract test gate (see §8.1) — every existing `mcp_test.go` JSON-RPC frame is replayed through the new `GatewayCore` path; both old and new paths must produce byte-identical responses. | CI step "contract-replay" green. | Low. |
| **P1** | Schema migration: add `protocol`, `state`, `parent_task_id`, `fallback_chain` to `registry_executions_public`. `gateway/receipt/emit.go` populates them (always `mcp`/`completed` for now). | Receipts surface `protocol: "mcp"`, `state: "completed"`. `parent_task_id` and `fallback_chain` nullable. | Low. |
| **P2** | A2A spec scaffolding: Agent Card model, `/.well-known/agent.json`, `/v1/a2a/agents/{id}/card`. Extend `/.well-known/functionfly.json` with `supported_protocols`. No execution yet. | `curl /.well-known/agent.json` returns valid card per A2A 0.3.0; well-known advertises both protocols. | Low. |
| **P2.5** | **The killer use case:** `tools/call` → delegate to peer A2A agent (see §4.6). Adds `a2a:agent:delegate` capability. Two-receipt chain (MCP receipt + child A2A receipt with `parent_task_id`). | Demo: peer agent receives a delegated task from a tool call, both receipts shareable. | Medium. |
| **P3** | A2A Task state machine + `tasks/send` + `tasks/get` + `tasks/cancel`. Updates `registry_executions_public.state`. Calls `GatewayCore.Call(ProtocolA2A)`. | End-to-end: peer agent sends task, gets receipt, can poll + cancel + see state transitions. | Medium. |
| **P4** | SSE streaming (`tasks/subscribe`). State changes emit SSE events. | Subscribe returns Server-Sent Events for `working` → `completed`. | Medium. |
| **P5** | **Capability-based routing (data-driven) + fallback chains + `mcp+agent` bundles.** `CapabilitySet` loaded from DB tables, not code. `gateway_fallback_fired_total{from,to,reason}` metric. Bundle kind `mcp+agent` ships. | Same input + degraded downstream → Receipt shows `fallback_chain: ["primary:down", "cache:hit"]`. Bundle install pulls a tool and an agent in one click. | Medium-High. |
| **P6** | Agent Card Directory endpoints + `agent_cards` table. | Publish + browse + version agent cards. | Low. |
| **P7** | Dashboard: `MCPRegistryPage`, `AgentCardsPage`, `A2AExplorerPage`, `BundlesPage`. | UI shipped, dark/light parity, `<TaskTimeline />` overlay works. | Low. |

**P0–P3 are the load-bearing work** for the "load-bearing for both protocols" claim. P2.5 is the demo. P5 is the moat (capabilities + bundles). P6–P7 strengthen the surface.

### 8.1 The contract test gate (P0.5)

The P0 refactor is dangerous because the MCP path is on the hot path for every agent. A pure unit test is not enough — subtle `*http.Request` mutations and context propagation issues slip through. **The contract test is the safety net:**

- Capture every JSON-RPC frame sent in `mcp_test.go` (and any new MCP integration tests).
- For each frame, record (status, body, headers) from the **pre-refactor** handler.
- After P0, replay each frame through the new `GatewayCore`-backed handler.
- Assert byte-equality of body and headers; status must match.

CI step name: `contract-replay`. Fails the build on any divergence. Lives in `internal/gateway/contract_test.go` so it runs alongside the unit tests.

---

## 9. Open questions (decide before P2)

1. **A2A spec version pin** — **must** be answered before P2 starts. Pin to `0.3.0` (the LF-governed version after ACP merge); revisit quarterly. Update `internal/gateway/version.go` only.
2. **ACP compatibility:** A2A is now the superset. We implement A2A-only; the REST shape is already ACP-friendly. No `/v1/acp` paths.
3. **Peer-to-peer auth:** A2A peers authenticate via FunctionFly-issued JWTs **or** their own JWT (verified via `peer_jwks_url` on their `agent_card`). Trust is recorded in the `agent_cards` table.
4. **DRE anchoring cost:** Default to L2 (Base). Ethereum L1 is opt-in per-receipt via the existing `chain` field on `EthereumAnchoringService`.
5. **State Fabric ↔ A2A:** State Fabric is the longer-lived state substrate. A2A Tasks are short-lived protocol artifacts in `registry_executions_public`. **Add a `state_ref` field** (nullable UUID pointing into State Fabric) so long-running A2A tasks can hand off durable state without coupling the two.
6. **`/.well-known/{agent_id}/agent.json`:** A2A spec primarily expects the gateway card. Per-agent cards at `/.well-known/...` are lower-priority. Confirm before P2.

---

## 10. Receipts as a product (observability + SLOs)

Receipts are not plumbing — they're a **product surface** with OG images, share URLs, milestones, view counts, and trending. They are FunctionFly's viral loop. Treat them accordingly.

**SLOs:**

| Path | SLO |
|---|---|
| `GET /v1/receipts/{id}` | p50 < 20ms, p99 < 100ms (Redis cache hit), p99 < 200ms (DB hit) |
| `GET /v1/receipts/{id}` (after revocation) | p99 < 50ms (must be fast — the user clicked a revoke link) |
| `POST /v1/receipts/{id}/run` | p99 < 500ms excluding the underlying function execution time |
| Receipt emission (from `GatewayCore.Call`) | p99 overhead < 5ms (must not slow down the hot path) |

**Metrics** (`/metrics` endpoint, scraped by Prometheus):

```
receipt_emit_total{protocol,state}                  # counter
receipt_emit_duration_seconds{protocol,state}       # histogram
receipt_get_total{cache,status}                     # counter (already exists)
receipt_milestone_fired_total{threshold,channel}    # counter (already exists)
receipt_milestone_duplicates_total                  # counter (already exists)
gateway_fallback_fired_total{from,to,reason}        # counter (NEW)
gateway_capability_check_total{capability,result}   # counter (NEW)
a2a_task_state_total{from_state,to_state}           # counter (NEW)
dre_anchor_total{protocol,chain,result}             # counter (NEW)
```

**Dashboards:** one for "Receipt health" (latency by cache layer, error rate, milestone fan-out rate) and one for "Two-protocol gateway" (calls per protocol, fallback rate, delegation chains).

**Backlog item:** SLO alerts — p99 `GET /v1/receipts/{id}` > 200ms for 5min pages the on-call.

---

## 11. The "no copy-paste" checklist

When the A2A work starts, **none** of the following should be re-implemented:
- ❌ New rate limiter (use `middleware/distributed_rate_limiter.go`, wrapped by `gateway/ratelimit.go`)
- ❌ New auth resolver (use `gateway/auth.go`)
- ❌ New receipt generator (use `gateway/receipt/emit.go`)
- ❌ New HMAC code (use package-level `gateway.HMACSign`/`HMACVerify`)
- ❌ New CORS handler (use `gateway/cors.go`)
- ❌ New DRE anchor trigger (use `gateway/receipt/anchor.go`)
- ❌ New milestone fan-out (use `gateway/receipt/milestone.go`)
- ❌ New execution path (use `GatewayCore.Call`)
- ❌ New verification flow (use `middleware/verification.go`)
- ❌ New capability allowlist code (use DB-loaded `CapabilitySet`)
- ❌ New task storage (use `registry_executions_public` with `protocol='a2a'`)
- ❌ New share/trending/revoke surface (use existing `/v1/receipts/*` endpoints)
- ❌ New well-known manifest (extend the existing one; share `gateway/version.go`)

The A2A package contains **only**: A2A spec shapes, task state machine (`a2a.TaskEngine`), SSE, peer-card validation, delegation parsing. Nothing else.

---

## 12. Naming for the market

The product is the **FunctionFly Two-Protocol Gateway** (or **MCP + A2A Gateway**). The endpoint is `api.functionfly.com`. The well-known manifest is `functionfly.json`. The discovery surfaces are `functionfly.com/mcp` and `functionfly.com/agents`. Use this name in the docs site, dashboard, partner onboarding, and the bundle kind `mcp+agent` README.

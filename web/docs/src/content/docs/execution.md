---
title: Execution
description: How function invocations work — lifecycle, runtimes, caching, and monitoring
---

## What Is an Execution

An **execution** is a single invocation of a registered FunctionFly function.
Every call — whether from an HTTP request, webhook, scheduled trigger, or
agent — is tracked with full metadata: duration, outcome, status code,
caching, verification status, and resource usage.

## Execution Lifecycle

```
Request → Parse → Lookup → Auth → Quota → Cache? → Route → Execute → Respond
                                     │         │        │
                                     │         │        └─ Sandbox timeout/memory
                                     │         └─ L1 deterministic cache
                                     └─ Rate limits, daily/hourly/minute caps
```

1. **Request parsing** — Extract `{author}/{name}/{version}` from URL, decode JSON body (max 10MB)
2. **Function lookup** — Resolve function from the registry
3. **Payment** — If the function has a price, check wallet balance and deduct
4. **Version resolution** — Resolve to latest or specified version; hydrate source from object storage
5. **Security checks** — Rate limits, quotas, abuse detection
6. **Cache check** — Deterministic functions may return a cached result (L1 Redis cache)
7. **Runtime routing** — Select the execution engine based on runtime type and plan tier
8. **Sandboxed execution** — Run the function with memory/CPU limits and timeout
9. **Post-execution** (async) — Verification, tracing, billing, DNA analysis

## Runtimes

| Runtime | Engine | Cold Start | Notes |
|---------|--------|------------|-------|
| Python | CPython-WASM / Firecracker MicroVM | 50–200ms | Enterprise uses MicroVM |
| Node.js | V8 isolate | 30–100ms | Default for JS/TS |
| Bun | Bun runtime | 20–80ms | Fast JS alternative |
| Deno | Deno runtime | 30–100ms | Secure by default |
| Go | Compiled binary | 10–50ms | Near-native speed |
| Rust/C/C++ | Wasmtime WASM | 5–30ms | Fastest cold starts |
| Ruby | Ruby runtime | 80–200ms | Full Ruby support |
| Kotlin/Swift | WASM | 10–50ms | Via WebAssembly |

## Cold Starts vs. Warm Invocations

**Cold start** — A new runtime instance is created. The runtime initializes,
loads function code, then invokes the handler. Tracked via the `cold_start`
flag in execution metrics.

**Warm invocation** — Reuses an already-loaded runtime instance. Files in
`/tmp` persist across warm invocations. No initialization overhead.

### Prewarming

FunctionFly's ML-powered prewarming (Holt-Winters forecasting) predicts
demand and pre-initializes runtime instances, reducing cold starts by
40–60%.

## Timeouts

Every execution has a timeout. If the function doesn't return within the
limit, the execution is killed with an `timeout` outcome.

| Runtime | Default Timeout | Max Timeout |
|---------|----------------|-------------|
| Most runtimes | 30s | 300s |
| Python external | 300s | 600s |
| MicroVM (Enterprise) | 30s | 300s |

Configurable per function via `timeout_ms` in the function version settings.

## Caching

Deterministic functions (same input → same output) benefit from L1 caching:

- Cache key: `sha256(input)` + function version
- Storage: Redis with configurable TTL
- Cached executions return instantly with `cached: true`

## Execution Queuing

When a node is under pressure, executions can be queued instead of rejected:

- Enabled via `EXECUTION_QUEUE_ENABLED=true`
- Triggered by `X-Queue-If-Busy: true` header
- Uses RabbitMQ or in-memory priority queue
- Prevents 503s during traffic spikes

## Outcomes

| Outcome | Description |
|---------|-------------|
| `success` | Function returned a valid response |
| `error` | Function threw an error or returned an error status |
| `timeout` | Execution exceeded the timeout limit |
| `oom` | Execution exceeded memory limits |
| `cancelled` | Execution was cancelled (e.g. client disconnect) |

## Verification & Certificates

### Replay Verification

Deterministic functions are automatically replayed to verify that the same
input produces the same output. Results are recorded in `drift_reports`.

### Merkle Execution Graph (MEG)

Every execution generates a Merkle tree hash over the input, output,
environment, and resource usage. This provides tamper-proof evidence of
what happened during execution.

### FXCERT Certificates

Executions can generate cryptographic certificates (FXCERTs) with:

- Input/output hashes
- Execution root hash (Merkle tree)
- Node signature (Ed25519)
- Platform signature
- Optional blockchain anchoring

## Billing

Each execution is metered for billing:

- **Per-call fee** — Fixed price per invocation
- **Compute fee** — Based on duration × memory (ms-GB)
- **Platform fee** — Percentage markup

Paid functions require authentication and sufficient wallet balance.

## Monitoring Executions

### Dashboard

- **Execution Explorer** — Full history with filtering by version, outcome, and verification status
- **Function detail page** — Recent executions tab, execution count, latest root hash
- **Dashboard home** — Execution rate chart (last 24h), trends

### API

```bash
# List executions for a function
curl https://api.functionfly.com/v1/registry/{author}/{name}/executions \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"

# Get a specific execution replay
curl https://api.functionfly.com/v1/registry/replay/{executionId}
```

### Metrics

FunctionFly exposes Prometheus metrics:

- `functionfly_function_invocations_total` — Total invocations by function/outcome
- Execution duration histograms
- Cache hit/miss ratios

## Data Retention

| Table | Default Retention |
|-------|-------------------|
| Execution log | 90 days |
| Public executions | 30 days |
| Resource usage | 90 days |
| MEG records | 365 days |
| Certificates | 365 days |
| Drift reports | 365 days |

Retention is configurable. Data under legal hold is never deleted.

## Next Steps

- [Functions](/functions/) — Writing and deploying functions
- [CLI](/cli/) — CLI reference
- [Analytics](/analytics/) — Execution analytics
- [Monitoring & Observability](/guides/monitoring/) — Production monitoring

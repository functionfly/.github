# internal/wasmpool/client

WasmPoolClient SDK for the orchestrator. Routes WASM execution to either
the in-process `wasm.InstancePool` (current behavior) or the external
`wasm-pool-service` (Phase 1+).

## Wiring (production)

In `cmd/orchestrator-api/main.go` (or wherever the server boots), after
the local pool is initialized:

```go
wasmpoolMgr, err := wasmpoolclient.NewManagerFromConfig(wasm.PerTenantPools)
if err != nil {
    log.Fatalf("wasm pool manager: %v", err)
}
defer wasmpoolMgr.Close()

// Optional: start the prewarm reconciler.
reconciler := wasmpoolclient.NewReconciler(
    wasmpoolclient.ReconcilerConfig{
        Tenants:    tenantList,
        Runtimes:   []string{"python"},
        MaxConc:    10,
        TickPeriod: 30 * time.Second,
    },
    wasmpoolclient.NewHTTPPrewarmClient(),
    wasmpoolMgr.Router().(*WasmPoolRouter).external, // see manager.go
)
reconciler.Start(ctx)
defer reconciler.Stop()
```

## Call sites

Replace direct `e.pool.Get` / `pool.Put` calls in `engines.go` and
`wasm_integration.go` with:

```go
resp, err := wasmpoolMgr.Execute(ctx, &wasmpoolclient.Request{
    TenantID: tenantID.String(),
    Runtime:  runtimeType,
    Input:    input,
    Timeout:  30 * time.Second,
})
```

This is deferred to a follow-up because the working tree has many
uncommitted changes; the SDK is wired and tested but the call sites
still use the pool directly. The router defaults to Local when
`WASM_POOL_EXTERNAL_PERCENT=0`, so there's no behavior change at the
default setting.

## Env vars (all optional, defaults shown)

| Var | Default | Purpose |
|-----|---------|---------|
| `WASM_POOL_EXTERNAL_PERCENT` | `0` | 0–100; 0 = always Local |
| `WASM_POOL_EXTERNAL_TENANTS` | empty | comma-separated tenant IDs always routed to External |
| `WASM_POOL_LOCAL_TENANTS` | empty | comma-separated tenant IDs always routed to Local (wins on conflict) |
| `WASM_POOL_EXTERNAL_DRY_RUN` | `false` | Local authoritative, External best-effort |
| `WASM_POOL_SERVICE_ADDR` | `wasm-pool-service:8084` | headless service DNS name |
| `WASM_POOL_GRPC_TLS` | `false` | enable mTLS |
| `WASM_POOL_GRPC_CERT_FILE` | — | mTLS client cert |
| `WASM_POOL_GRPC_KEY_FILE` | — | mTLS client key |
| `WASM_POOL_GRPC_CA_FILE` | — | mTLS server CA |
| `WASM_POOL_GRPC_AUTH_TOKEN` | — | HMAC shared secret (dev) |

## Metrics

| Metric | Labels | Meaning |
|--------|--------|---------|
| `wasm_pool_router_routing_decisions_total` | `decision`, `reason` | Per-request routing decision |
| `wasm_pool_client_latency_seconds` | `target`, `runtime` | Client-side execute latency |
| `wasm_pool_breaker_state` | `state` | Circuit breaker state (0/1/2) |
| `wasm_pool_dry_run_divergences_total` | `field` | Dry-run divergences by field |

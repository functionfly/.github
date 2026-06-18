# WASM Pool Externalization — Runbook

This runbook covers the rollout of the `wasm-pool-service` (replaces the
in-process WASM pools in the orchestrator) and the chaos / rollback
procedures described in
`.kilo/plans/externalize-wasm-pool-service.md` Phase 3.

For the design, see `.kilo/plans/externalize-wasm-pool-service.md`.
For the SDK, see `internal/wasmpool/client/README.md`.
For the service, see the [`wasm-pool-service` repo](../wasm-pool-service/README.md).

## Architecture (steady state)

```
┌────────────────────┐       gRPC (port 8084)       ┌─────────────────────┐
│  orchestrator-api  │ ───────────────────────────▶ │  wasm-pool-service  │
│  (3–10 replicas)   │     consistent hash on       │  (3–20 replicas)    │
│                    │     tenantID                  │                     │
└────────────────────┘                              └─────────────────────┘
        │                                                   │
        │ in-process fallback (WASM_POOL_LOCAL_FALLBACK_SIZE=1)
        ▼
   wasm.InstancePool (deprecated, kept for emergency)
```

When `WASM_POOL_EXTERNAL_PERCENT=100` and the external pool is healthy, no
in-process instance is ever created. The local pool is sized to 1
instance per tenant and exists only as a circuit-breaker fallback.

## Rollout checklist

| Day | Action | Verify | Rollback |
|-----|--------|--------|----------|
| 0 (staging) | Deploy `wasm-pool-service` (3 replicas, headless Service) | `kubectl get pods -l app=wasm-pool-service` shows 3/3 Running; `/health` returns 200 on all | `kubectl delete deploy wasm-pool-service` |
| 0 (staging) | Synthetic load: 100 RPS, 50 tenants | p99 < 50 ms added latency; cold-start rate < 5%; 0 errors | Lower RPS in `tests/load/wasm-pool.go` |
| 0 (staging) | Enable `WASM_POOL_EXTERNAL_PERCENT=1` in staging orchestrator | `wasm_pool_router_routing_decisions_total{reason="percentage"}` shows ~1% external | `kubectl set env deploy/orchestrator-api WASM_POOL_EXTERNAL_PERCENT=0` |
| 0 (prod)    | Canary tenant via `WASM_POOL_EXTERNAL_TENANTS` + `WASM_POOL_EXTERNAL_PERCENT=1` | Canary tenant shows external latency on dashboard; control cohort stable | `kubectl set env deploy/orchestrator-api WASM_POOL_EXTERNAL_TENANTS=` |
| 1 | `WASM_POOL_EXTERNAL_PERCENT=5` | Per-tenant latency delta < 10 ms vs control | `WASM_POOL_EXTERNAL_PERCENT=0` |
| 2 | `WASM_POOL_EXTERNAL_PERCENT=25` | p99 within 2× baseline | `WASM_POOL_EXTERNAL_PERCENT=0` |
| 3 | `WASM_POOL_EXTERNAL_PERCENT=50` | Cold-start count drops 2–5× vs baseline (the whole point) | `WASM_POOL_EXTERNAL_PERCENT=0` |
| 4–7 | `WASM_POOL_EXTERNAL_PERCENT=100`; unlock HPA to 3–20 | `wasm_pool_breaker_state{state="open"}` stays at 0 for 24h | `WASM_POOL_EXTERNAL_PERCENT=0` |
| 8+ | Steady state: monitor for 1–2 weeks | Zero SEV-2 attributed to pool | n/a |
| T+30 | Decommission: set `WASM_POOL_LOCAL_FALLBACK_SIZE=1`; deprecate in-process pool | Memory per orchestrator pod drops ~4 GB | n/a |

## Rollback procedures (in priority order — pick the smallest change first)

### 1. Per-tenant rollback (instant, no deploy)

Remove the tenant from `WASM_POOL_EXTERNAL_TENANTS`:

```bash
kubectl set env deploy/orchestrator-api \
  WASM_POOL_EXTERNAL_TENANTS=acme-corp,beta-corp
# To remove acme-corp:
kubectl set env deploy/orchestrator-api \
  WASM_POOL_EXTERNAL_TENANTS=beta-corp
```

Effect: within 30 s (next ring refresh + reconciler tick), acme-corp routes
back to Local.

### 2. Global rollback (instant, no deploy)

```bash
kubectl set env deploy/orchestrator-api WASM_POOL_EXTERNAL_PERCENT=0
```

Effect: every request goes Local immediately. Circuit-breaker state is
irrelevant. No traffic to the pool service.

### 3. Pool service down (automatic)

The SDK's circuit breaker opens after 5 consecutive failures / 30 s and
all traffic falls back to Local. The pool service can be down indefinitely
without user-visible impact (modulo higher cold-start latency on Local
during the outage).

To force the circuit open without waiting 30 s:

```bash
# Scale to 0 (test only — do not do this in prod):
kubectl scale deploy wasm-pool-service --replicas=0
```

### 4. Bad release detection

The following alerts trigger within 5 minutes of a regression:

- `WasmPoolHighP99` — external p99 > 100 ms above baseline
- `WasmPoolHighColdStartRate` — cold-starts > 2× baseline
- `WasmPoolCircuitBreakerOpen` — breaker open for > 1 minute

Investigate: check the pool service logs for the affected replica:

```bash
kubectl logs -l app=wasm-pool-service --tail=500 --since=10m | grep -i error
```

If the regression is service-side, rollback the pool service deploy:

```bash
kubectl rollout undo deploy/wasm-pool-service
```

If the regression is orchestrator-side, rollback the orchestrator deploy
(see below).

### 5. Full rollback (deploy)

Revert the orchestrator deployment that introduced the SDK. The Local
path is unchanged across SDK versions, so a `git revert` + redeploy
restores current behavior:

```bash
# Identify the SDK-introducing commit:
git log --oneline -- internal/wasmpool/
# Revert it (and any follow-up commits that depend on it):
git revert <commit-sha>
git push
# Redeploy:
kubectl rollout restart deploy/orchestrator-api
```

## Operational queries

### What's the current routing mix?

```promql
sum by (decision) (rate(wasm_pool_router_routing_decisions_total[5m]))
```

### Is the breaker tripping?

```promql
max_over_time(wasm_pool_breaker_state{state="open"}[15m]) == 1
```

### Cold-start rate (the key benefit)

```promql
rate(wasm_execution_cold_starts_total[5m])
```

Compare against the pre-externalization baseline (typically 1–2/sec
under 100 RPS for a 3-replica orchestrator). With external pool, this
should drop 2–5× because warm instances are shared across orchestrator
replicas.

### Memory per pool replica

```promql
container_memory_working_set_bytes{pod=~"wasm-pool-service-.*"}
```

Alert if any replica exceeds 28 GB (we set the limit to 32 GB; 28 GB gives
us 4 GB of headroom for OOM-kill investigation).

### Dry-run divergence rate (Phase 3 day 0–3 only)

```promql
sum by (field) (rate(wasm_pool_dry_run_divergences_total[5m]))
```

If this is non-zero, investigate the affected fields (`output`,
`error`, `latency_ms`, `cold_started`) and the affected tenants.

## Synthetic load test (staging)

Run from any host with network access to the staging wasm-pool-service:

```bash
go run ./tests/load/wasm-pool \
  -addr staging-wasm-pool-service:8084 \
  -rps 100 \
  -tenants 50 \
  -duration 10m
```

The test:
- Spins up N goroutines, each driving `rps/N` requests/sec.
- Hashes tenant IDs across the 50-tenant set.
- Reports p50 / p95 / p99 latency and cold-start count every 30s.
- Fails (exit 1) if error rate > 0.1% or p99 > 50 ms above the local
  baseline.

## Decommissioning checklist (T+30 days)

After 1–2 weeks at 100% with zero SEV-2 incidents:

1. **Shrink the local fallback pool**

   ```bash
   kubectl set env deploy/orchestrator-api WASM_POOL_LOCAL_FALLBACK_SIZE=1
   ```

   Effect: orchestrator pods drop from ~10 instances/tenant × 3–10 replicas
   to 1 instance/tenant × 3–10 replicas. Memory per pod drops ~30–40%.

2. **Mark in-process pools deprecated**

   Add `@deprecated` doc comments to `wasm.NewInstancePool`,
   `wasm.NewPythonRuntimePool`, and the `cpythonWasmPool` helper in
   `engines.go`. Do not delete yet — Local is still the breaker fallback.

3. **Lower HPA floor (optional)**

   If steady-state load is low, the orchestrator HPA can be relaxed from
   3–10 to 3–5 replicas. The external pool's HPA (3–20) absorbs the load.

4. **Remove dry-run mode**

   ```bash
   kubectl set env deploy/orchestrator-api WASM_POOL_EXTERNAL_DRY_RUN=false
   ```

   The `wasm_pool_dry_run_divergences_total` metric stops incrementing.
   Save the alert for 30 more days, then remove the alert rule.

5. **Follow-up release: delete in-process pool**

   Once `WASM_POOL_LOCAL_FALLBACK_SIZE=0` is safe (no breaker fallbacks in
   30 days), remove the local pool instantiation from `BuildRuntimeRouter`.
   This is a code change, not just an env var — see the plan's "Mark the
   in-process `InstancePool` / `PythonRuntimePool` deprecated in code;
   remove in a follow-up release."

## On-call escalation

| Symptom | First responder | Escalation |
|---------|-----------------|------------|
| Pool service pods crash-looping | Check `kubectl describe pod` for OOMKilled | Platform team |
| Circuit-breaker open for > 1 min | Check pool service health, then page on-call | Backend team |
| p99 latency spike on external cohort | Check `wasm_pool_execute_latency_seconds` per replica | Backend team |
| Dry-run divergence | Check the affected field and tenant | Backend team |
| Orchestrator pods OOM after local pool shrink | Revert `WASM_POOL_LOCAL_FALLBACK_SIZE` to 2 | Platform team |

## References

- Plan: `.kilo/plans/externalize-wasm-pool-service.md`
- SDK: `internal/wasmpool/client/README.md`
- Service: [`wasm-pool-service` repo](../wasm-pool-service/README.md)
- Module: [`wasm` repo](../wasm/README.md)
- Metrics: `deploy/monitoring/alert_rules.yml` (search for `wasm_pool_`)

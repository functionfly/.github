# Health Monitoring and Failover (MVP1)

## Probe strategy

- Ping each backend every 3–5 seconds.
- Two probe types:
  - `GET /healthz` for correctness
  - `GET /ping` for latency

Store results in Postgres.

## Health state

Computed per backend:

- last_ok_ts
- last_error_ts
- consecutive_failures
- ewma_latency_ms
- error_rate_window

## Circuit breaker

States:

- CLOSED: normal routing
- OPEN: removed from routing
- HALF_OPEN: limited test traffic

Transitions:

- CLOSED -> OPEN: N consecutive failures or high windowed error rate
- OPEN -> HALF_OPEN: cooldown elapsed
- HALF_OPEN -> CLOSED: M successes
- HALF_OPEN -> OPEN: any failure

## Failover behavior

Default:

- Retry only idempotent methods.
- Timeout budget per hop is bounded so a retry still fits overall SLA.

## Safety valves

- If all backends are OPEN, allow fallback to the least-bad backend in HALF_OPEN with strict rate.


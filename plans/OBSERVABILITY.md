# Observability (MVP1)

## Logging

Structured logs for:

- route decisions
- proxy outcomes
- health probe outcomes
- circuit breaker transitions

Required fields:

- request_id
- tenant_id
- app_id
- backend_id
- provider
- region
- latency_ms
- outcome

## Metrics

- probe_success_rate{backend}
- probe_latency_ms{backend}
- request_latency_ms{app,backend}
- request_error_rate{app,backend}
- circuit_state{backend}

## Trace propagation

- Generate `X-Request-Id` at Caddy if missing.
- Propagate through orchestrator and to edge targets.


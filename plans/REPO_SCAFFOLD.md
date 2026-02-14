# Repository scaffold (MVP1)

This document describes the target repository layout and the minimal set of files to implement first.

## Proposed mono-repo layout

```
.
├─ cmd/
│  ├─ orchestrator-api/
│  └─ health-monitor/
├─ internal/
│  ├─ api/
│  ├─ routing/
│  ├─ adapters/
│  ├─ storage/
│  ├─ auth/
│  └─ observability/
├─ migrations/
├─ deploy/
│  ├─ caddy/
│  ├─ systemd/
│  └─ docker/
├─ edge-targets/
│  ├─ cloudflare-workers/
│  ├─ vercel/
│  ├─ fly/
│  └─ deno-deploy/
├─ cli/
│  └─ ffly/
├─ web/
│  └─ dashboard/
└─ plans/
```

## Services

### Go: orchestrator-api

Responsibilities:

- CRUD for tenants/apps/backends
- route decision endpoint used by the edge entry
- audit event recording

### Go: health-monitor

Responsibilities:

- probe scheduling
- latency EWMA + error windows
- circuit breaker state transitions

## Edge entry

### Caddy

- TLS termination
- rate limiting and basic request shaping
- proxy to orchestrator for route decision
- proxy to selected backend

Note: In MVP1, keep routing decision in Go; Caddy remains configuration-light.

## Local development story

- Postgres via docker compose
- Go services run locally
- One sample edge target runs locally to validate routing

## CI

- Go format + lint + unit tests
- Web lint
- basic build artifacts


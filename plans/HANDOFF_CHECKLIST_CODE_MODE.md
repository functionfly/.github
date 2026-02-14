# Code mode handoff checklist (MVP1)

This is the execution order for implementing the scaffold.

## 1) Repo bootstrap

- Initialize Go module
- Add minimal Makefile or task runner
- Add docker compose for Postgres

## 2) Storage

- Create Postgres schema + migrations for:
  - tenants, users, apps, backends
  - health_checks, circuit_state
- Add storage package with repository interfaces

## 3) Orchestrator API

- Implement JWT auth (email+password for MVP1)
- Implement endpoints:
  - create app
  - add backend
  - list backends
  - route debug endpoint
- Implement routing core with scoring + failover list

## 4) Health monitor

- Implement probe loop (3–5s)
- Persist probe results
- Implement circuit breaker transitions

## 5) Caddy config

- Configure `/{appSlug}/*` routing endpoint
- Ensure `X-Request-Id` propagation
- Add basic per-app rate limiting

## 6) Edge targets templates

- Implement a minimal edge target for each provider:
  - `GET /healthz`
  - `GET /ping`
  - verify HMAC
  - proxy handler

## 7) CLI

- Minimal commands:
  - `ffly init`
  - `ffly backend add`
  - `ffly status`

## 8) Smoke tests

- Local: run Postgres + API + health monitor + one edge target
- Register 2 backends (one intentionally unhealthy)
- Verify routing picks healthy backend and fails over


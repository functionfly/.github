# MVP1 Scope and Success Criteria

## In-scope for MVP1

### Platform baseline

- Go control-plane API behind Caddy
- Postgres as system of record
- Cloudflare DNS for public routing

### Routing

- Health-aware selection
- Latency-informed scoring (EWMA)
- Circuit breaker (closed open half open)
- Fast failover for idempotent methods

### Provider coverage

- BYO edge targets with uniform contract for:
  - Cloudflare Workers
  - Vercel
  - Fly.io
  - Deno Deploy

### Security

- JWT for dashboard
- App keys for programmatic access
- HMAC signing between orchestrator and edge targets
- Rate limiting

### Observability

- Structured logs with request id
- Basic metrics emitted from services

## Explicitly out-of-scope for MVP1

- Storing and managing customer provider tokens by default
- One click deployment into customer provider accounts
- Paid SLA enforcement beyond best effort routing
- Arbitrary user code execution on your infrastructure
- Full dashboard UX polish (focus on core functionality)
- Global cache product with invalidation across providers
- Advanced traffic shaping (A B tests, canary, weighted routing)
- Multi region control plane (single VPS is acceptable)

## Success criteria

- A user can register an app and add 2 to 4 backends.
- Health monitor detects an unhealthy backend within one probe interval and stops routing to it.
- Router chooses the lowest score healthy backend and fails over on error.
- End to end request id is present across Caddy, orchestrator, and edge targets.
- Local smoke test demonstrates failover and recovery.

## MVP1 demo script

1. Create an app.
2. Add 2 backends.
3. Make one backend return 500.
4. Show routing to healthy backend.
5. Restore backend and show circuit moves back to closed.


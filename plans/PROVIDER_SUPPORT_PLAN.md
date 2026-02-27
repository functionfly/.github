# Provider Support Plan (Cloudflare Workers, Vercel, Fly.io)

Status: planning complete, ready for handoff to implementation.

## Decisions (v1 defaults)

### Cloudflare Workers deployment mechanism

- Use Cloudflare REST API upload + route binding.
- No `wrangler` dependency in the orchestrator.

Rationale:

- Removes the need to shell out / manage `wrangler` versions.
- Enables clean token-scope enforcement and deterministic API calls.

### Domains + routing

- Support **zone routes for custom domains**.
- Support **optional DNS switching**.

### DNS switching approach (default)

- Manage a **single CNAME-style record per hostname** and update its value to point at the active target.
- Avoid full A/AAAA management in v1.

Rationale:

- Keeps blast radius smaller than managing multiple record types.
- Compatible with blue/green style flips by changing one record value.

## Current code reality

Provider adapters today only implement runtime behaviors:

- [`common.ProviderAdapter`](internal/adapters/common/interface.go:11)
  - `GetName`, `ValidateConfig`, `GetRegions`
  - `HealthCheck`, `SignRequest`, `GetRequestTimeout`

Examples:

- Cloudflare health check uses `/healthz` in [`CloudflareAdapter.HealthCheck()`](internal/adapters/cloudflare/adapter.go:91)
- Vercel health check uses `/healthz` in [`VercelAdapter.HealthCheck()`](internal/adapters/vercel/adapter.go:103)
- Fly health check uses `/healthz` in [`FlyAdapter.HealthCheck()`](internal/adapters/fly/adapter.go:108)

Implication:

- Deployment, env var management, rollback require **new interfaces + orchestration paths**.

## Target capabilities (what we will add)

### Cloudflare Workers

- Deploy Worker via API (script/module upload)
- Bind route(s) to zone/custom domain
- Manage env vars/secrets
- Health check endpoint contract: `/healthz`
- Rollback support (optional early): redeploy previous artifact
- Optional DNS switching (record update)

### Vercel

- Deploy API routes / edge functions via Vercel API
- Project linking
- Preview vs production handling
- Environment variables

### Fly.io

- Deploy app via Machines API
- Multi-region scaling awareness
- Health check endpoints
- Restart control
- Secrets management

## Proposed abstractions

### 1) Expand the adapter contract

Keep the existing runtime interface, and add a deployment-lifecycle interface that providers can implement.

Proposed new interface (names illustrative):

- `Deploy(ctx, spec) -> deploymentID, metadata`
- `SetEnv(ctx, deploymentID, vars, secrets) -> result`
- `BindRoutes(ctx, deploymentID, routes) -> result`
- `GetDeploymentStatus(ctx, deploymentID) -> status`
- `Rollback(ctx, toDeploymentID) -> result`

Design notes:

- Split runtime traffic signing/health-check from provider deployment concerns.
- Make rollback optional for providers that cannot do it cleanly.

### 2) Common deployment spec

Define a provider-agnostic structure that contains:

- artifact reference (bundle bytes or object store key)
- desired endpoints (routes)
- env vars vs secrets
- provider identifiers (account/project/app)
- health check expectations

## Cloudflare Workers: detailed design (v1)

### Auth + token handling

Requirements:

- Token stored encrypted-at-rest.
- Least privilege scopes.
- Redact token from logs.
- Rate limit strategy: backoff + retry on 429.

Key hidden complexity:

- Account-level vs zone-level permissions.
- Different customers may grant different scope sets.

### Upload API

Plan:

- Build Worker bundle (module format preferred).
- Upload to Cloudflare Workers API.
- Stamp version metadata in:
  - deployment record in DB
  - optional Worker metadata header/kv value for healthz reporting

### Route binding

Plan:

- Support mapping one or more route patterns to the Worker.
- Provide route binding config in provider config.

### Env vars + secrets

Plan:

- Distinguish vars (non-secret) and secrets.
- Apply changes as part of deploy pipeline (or in a dedicated call).

### Health check

- Require `/healthz` to return 200 when healthy.
- Optionally return headers like `X-Workers-Version` for observability.

### Rollback (v1)

Default approach:

- Keep last N deployment artifacts in FunctionFly storage.
- Rollback = redeploy a prior artifact (idempotent API upload + route rebind).

Limitations:

- Not an instant revert unless Cloudflare supports version pinning for our chosen API path.

### Optional DNS switching

Default approach:

- For a hostname, manage a single DNS record and update its value to point to the active edge target.
- Requires additional DNS scopes.

## Execution plan (handoff to code mode)

1. Update adapter interfaces and plumbing to support deployment lifecycle.
2. Implement Cloudflare deployment client:
   - token auth
   - upload
   - env vars/secrets
   - zone route binding
   - DNS switching
3. Persist deployment state:
   - deployment IDs
   - artifacts
   - status transitions
4. Add rollback path:
   - list historical deployments
   - redeploy prior artifact
5. Add acceptance tests + mock API tests.
6. Extend Vercel + Fly adapters using the same contract.

## Tooling note

Workspace semantic search tool currently fails with a vector dimension mismatch; planning used regex search + direct file reads instead.


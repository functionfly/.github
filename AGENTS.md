# FunctionFly Development Guide

See `README.md` for project overview and `CONTRIBUTING.md` for Git workflow. This file is the main reference for **Cursor (and other AI agents)** working in this repo.

---

## Quick Start

One-liner to verify the dev stack:

```bash
sudo pg_ctlcluster 17 main start && redis-server --daemonize yes
source .env && export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=functionfly DB_SSLMODE=disable REDIS_ADDR=localhost:6379 DEVELOPMENT=true SKIP_MIGRATION_VALIDATION=true VERIFICATION_ENABLED=false
./bin/orchestrator-api --skip-migrations
```

Then in another terminal: `curl http://localhost:8080/api/health`

---

## Codebase map (for agents)

| Area | Location | Notes |
|------|----------|--------|
| **API routes** | `internal/api/routes.go` | Route registration and middleware |
| **API handlers** | `internal/api/handlers/<domain>/` | One package per domain (e.g. `vault`, `registry`, `admin`, `trustapi`) |
| **Trust API** | `internal/api/handlers/trustapi/`, `internal/storage/trustapi/` | Revocation and trust verification APIs |
| **Storage / DB** | `internal/storage/`, `internal/storage/sql/` | Repositories, migrations; Postgres + optional Redis |
| **Auth** | `internal/auth/`, `internal/api/middleware/auth.go` | Sessions, GBA plugins (MFA, SAML, WebAuthn) |
| **Dashboard (React)** | `web/dashboard/src/` | Vite SPA; `pages/`, `components/` (includes `frg/` visual workflow editor), `hooks/`, `lib/` |
| **Marketing site (Astro)** | `web/site/` | Static landing, trust, legal, blog shell; `bun run dev` → port **4321**; `bun run build` → `dist/` |
| **Public docs (Astro)** | `web/docs/` | User-facing guides (synced Markdown); separate from dashboard; `bun run dev` → port **4322**; dashboard `/docs` redirects here |
| **Deploy / edge** | `deploy/`, `deploy/edge/` | Caddy, DNS, VPS/edge scripts |
| **Cloudflare** | `docs/CLOUDFLARE.md`, `deploy/cloudflare/`, `deploy/dns/` | DNS, CDN, R2, Workers, Tunnel, Pages |
| **Repo docs (Markdown)** | `docs/` | Design, ops, internal guides (not the public docs site) |
| **Local PG 17 + pgvector** | `docs/LOCAL_POSTGRES_17.md` | PostgreSQL 17 with pgvector and extensions for local dev |
| **Go modules** | `go.work`, `go.mod` | Workspace with main module + `cmd/flypy-go/` + `cmd/delete-functions/` for incremental builds |

When adding API surface: add handler in `internal/api/handlers/`, register in `internal/api/routes.go`, and use existing storage/auth patterns.

---

## Services overview

| Service | How to run | Port |
|---------|-----------|------|
| **Orchestrator API** (Go) | `./bin/orchestrator-api --skip-migrations` or `make dev` | 8080 |
| **SAR Runtime** (Rust) | `cd runtimes/sar && cargo run` (needs NATS on 4222) | 8082 |
| **Dashboard** (Vite/React) | `cd web/dashboard && npx vite --host 0.0.0.0` | 3000 |
| **Docs site** (Astro) | `cd web/docs && bun run dev` | 4322 |
| **Marketing site** (Astro) | `cd web/site && bun run dev` | 4321 |
| **PostgreSQL** | `sudo pg_ctlcluster 17 main start` (see `docs/LOCAL_POSTGRES_17.md` to replace PG 16 with 17) | 5432 |
| **Redis** | `redis-server --daemonize yes` | 6379 |
| **AI Service (FlyMind)** | `cd ai-service && uv sync && uv run start` | 8081 |

---

## Starting the backend

1. **Start dependencies** (required first):

   ```bash
   sudo pg_ctlcluster 17 main start   # or 16 if that's the cluster you have
   redis-server --daemonize yes
   ```

2. **Environment:** Ensure `.env` exists, then:

   ```bash
   source .env
   export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=functionfly DB_SSLMODE=disable
   export REDIS_ADDR=localhost:6379 DEVELOPMENT=true SKIP_MIGRATION_VALIDATION=true VERIFICATION_ENABLED=false
   ./bin/orchestrator-api --skip-migrations
   ```

   If you still have PostgreSQL 16 and want a single cluster, see `docs/LOCAL_POSTGRES_17.md` to replace it with PG 17 on port 5432.

The `--skip-migrations` flag is required because the `migrations/` directory has duplicate sequence numbers that break golang-migrate. Schema is applied via direct SQL during initial setup.

### Verify it's working

```bash
curl http://localhost:8080/api/health
```

---

## Nx (repo root)

| Script | What runs |
|--------|-----------|
| `bun dev` | **dashboard** (3000), **marketing** `web/site` (4321), **docs** `web/docs` (4322) — the usual local trio for redirects between apps. |
| `bun run dev:all` | Every project that defines `dev` (includes **admin-dashboard**). |

`nx run <project> -t dev` works per app; `nx show projects` lists project names.

---

## Starting the dashboard

From repo root (matches Nx `dashboard` dev target):

```bash
nx run dashboard -t dev
```

Or directly:

```bash
cd web/dashboard
VITE_API_URL=http://localhost:8080 npx vite --host 0.0.0.0 --port 3000
```

Vite proxies `/api/*` to the Go backend (port 8080). Use `VITE_API_URL` when the API is on a different host/port. The package script `bun run dev` in `web/dashboard` may use Infisical; **Nx uses plain Vite** so local dev works without the Infisical CLI.

---

## Running tests and lint

- **Go tests:** `go test ./internal/...` — storage tests expect Postgres on port **5432** (may fail if using Docker on 5434).
- **Go lint:** `golangci-lint run` (binary at `~/go/bin/golangci-lint`).
- **Go build:** `go build -o bin/orchestrator-api ./cmd/orchestrator-api`
- **Dashboard:** `cd web/dashboard && npx vitest run` for tests. ESLint has a known broken import (`eslint-import-resolver-typescript` default export); treat as pre-existing.

### Optimized Build & Test Commands

The project uses Go workspaces (`go.work`) and optimized build flags for faster development:

| Command | Purpose |
|---------|---------|
| `make build` | Standard build with `-trimpath` |
| `make build-fast` | Fast dev build (allows multiple errors, smaller binaries) |
| `make build-ci` | CI-optimized (CGO disabled, stripped binaries) |
| `make build-all-modules` | Build all workspace modules |

**Test shortcuts (fast feedback):**

| Command | Purpose |
|---------|---------|
| `make test-short` | Skip heavy integration tests (uses `-short` flag) |
| `make test-parallel` | Use all CPU cores (`-parallel=NCPU`) |
| `make test-fast` | Cached, parallel, short mode — best for local dev |
| `make test-changed` | Only changed packages (requires `gotestsum`) |
| `make test-watch` | Watch mode for continuous testing |
| `make test-ci` | CI-optimized with rerun on failure |

**Build cache management:**

| Command | Purpose |
|---------|---------|
| `make cache-info` | Show cache paths |
| `make cache-stats` | Show disk usage |
| `make cache-clean` | Clean build cache |

**IDE Performance (VS Code):**

`.vscode/settings.json` includes gopls tuning:
- Excludes `vendor/`, `testdata/`, `node_modules/`, build dirs from indexing
- Disables heavy analyses (`unusedparams`, `shadow`)
- Enables `experimentalWorkspaceModule` for monorepo support
- Disables auto-updates for faster startup

**Environment variables for build tuning:**

```bash
# Use RAM disk for build cache (optional speedup)
mount -t tmpfs -o size=8G tmpfs /mnt/go-cache
GOCACHE=/mnt/go-cache make build

# Custom module cache location
GOMODCACHE=/fast-ssd/go-mod make deps
```

---

## Database Migrations

### Migration Naming Convention

**All new migrations MUST use the timestamp format:** `YYYYMMDDHHMMSS_description.sql`

Use the helper script:
```bash
./scripts/create-migration.sh "add_user_preferences_table"
```

Validate before committing:
```bash
./scripts/validate-migrations.sh
```

See `MIGRATIONS.md` for full policy details.

### Creating Migrations (Idempotent SQL)

When writing migrations, use idempotent operations:
- `CREATE TABLE IF NOT EXISTS` instead of `CREATE TABLE`
- `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`
- `DROP INDEX IF EXISTS`, `DROP TABLE IF EXISTS` in down migrations

### Migration History

The project previously had duplicate sequence numbers that have been resolved:
- `000250` duplicates renamed to `20260412180000` and `20260412180100`
- `20260412175400` duplicates resolved

### Database Indexes for Performance

Recent billing performance indexes (see `migrations/20260419131000_billing_performance_indexes.up.sql`):
- `idx_payment_retries_status_next_retry` - For dunning scheduler queries
- `idx_payment_retries_grace_period_status` - For service suspension monitoring
- `idx_invoices_period_tenant` - For period-based billing reports
- `idx_cost_allocation_entries_timestamp` - For data retention cleanup queries

### Data Retention Policy (Billing)

Compliance-based data retention is implemented for cost allocation entries:

| Retention Tier | Period | Data Affected | Implementation |
|---------------|--------|---------------|----------------|
| Detailed execution logs | 90 days | High-volume cost allocation entries | `CleanupCostAllocationByRetention()` in billing repository |
| Financial aggregates | 7 years | Invoice-level data (SOX compliance) | `CleanupFinancialAggregatesAfterRetention()` |
| Audit logs | Configurable | `retention_audit_log` table tracks all deletions | Automatic |

**Configuration (environment variables):**

| Variable | Purpose | Default |
|----------|---------|---------|
| `DATA_RETENTION_ENABLED` | Enable scheduled cleanup | `true` |
| `DATA_RETENTION_CRON` | When to run cleanup | `0 3 * * *` (3 AM daily) |
| `DATA_RETENTION_DETAILED_DAYS` | Days to keep execution logs | `90` |
| `DATA_RETENTION_FINANCIAL_YEARS` | Years to keep financial data | `7` |
| `DATA_RETENTION_DRY_RUN` | Log only, don't delete | `false` |
| `DATA_RETENTION_SKIP_IF_LEGAL_HOLD` | Skip cleanup if legal holds active | `true` |

**Legal Holds:** Use the `legal_holds` table to block deletion for litigation/audit. Check `is_under_legal_hold()` function before any bulk deletion.

---

## Known gotchas

1. ~~**Migrations:** Duplicate sequence numbers have been resolved.~~ Use timestamp format (`YYYYMMDDHHMMSS`) for all new migrations. Use `--skip-migrations` only for local dev if needed.
2. **resend-go:** Upgraded from v2.0.0 to v2.28.0. `ReplyTo` is now `string` (was `*string`); `client.Keys.Get()` was removed.
3. **Stub / beta surfaces:** `internal/adapters/functionfly` still returns "not implemented" for some paths. State Fabric routes are registered (admin/platform); verify behavior before promising them in contracts. Grep for `not implemented` in `internal/api/handlers` before launch-critical features.
4. **Postgres audit trigger:** `audit_trigger_function()` expects `ip_address` cast to `::inet` (fixed in DB setup).
5. **Admin test account (local dev only):** `admin@functionfly.local` / `admin123` (bcrypt). For production, use `go run ./cmd/create-admin -production` and `ADMIN_CREATE_PASSWORD` — see `docs/ADMIN_SETUP_README.md`.
6. **Dashboard:** `index.html` was added for Vite; `bun` is the JS package manager — run `bun install` from repo root for workspace deps.
7. **Wallet low-balance alerts:** `AGENT_WALLET_LOW_BALANCE_USD` controls when low-balance notifications are sent (default: `5.00` USD).

---

## Vault security (Secrets)

The Secrets Vault is **zero-knowledge**: the server never sees plaintext or the decryption passphrase. Encryption and decryption are done **client-side** (dashboard: `web/dashboard/src/utils/vault-crypto.ts`). The API stores only AES-256-GCM ciphertext + IV/salt/tag. There is **no server-side decrypt endpoint** by design. For audit retention and token cleanup, see `docs/VAULT_OPERATIONS.md`.

---

## Common patterns

### Adding a new API endpoint

1. Add handler in `internal/api/handlers/<domain>/`
2. Register route in `internal/api/routes.go`
3. Add storage/repository methods in `internal/storage/sql/` if needed
4. Follow existing auth patterns (see `internal/api/middleware/auth.go`)

### Database changes (manual mode)

With `--skip-migrations`, apply schema changes via direct SQL:

```bash
psql -h localhost -p 5432 -U postgres -d functionfly -f your_changes.sql
```

---

## Environment quick reference

| Variable | Purpose | Default |
|----------|---------|---------|
| `DEVELOPMENT` | Enable dev mode features | `false` |
| `SKIP_MIGRATION_VALIDATION` | Skip migration checks (required) | `false` |
| `VERIFICATION_ENABLED` | Enable trust verification | `true` |
| `AGENT_WALLET_LOW_BALANCE_USD` | Low balance alert threshold | `5.00` |
| `REDIS_ADDR` | Redis connection | `localhost:6379` |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` / `DB_SSLMODE` | Postgres connection | varies |

---

## Troubleshooting

- **API won’t start:** Ensure Postgres and Redis are running and env vars are set (especially `DB_*`, `REDIS_ADDR`).
- **Dashboard can’t reach API:** Check `VITE_API_URL` and that the orchestrator is on 8080; Vite proxy only applies to `/api/*`.
- **Storage tests fail:** Point DB at port 5432 (local Postgres) or adjust test config to match your DB port.

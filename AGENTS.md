# FunctionFly Development Guide

See `README.md` for project overview and `CONTRIBUTING.md` for Git workflow. This file is the main reference for **Cursor (and other AI agents)** working in this repo.

---

## Codebase map (for agents)

| Area | Location | Notes |
|------|----------|--------|
| **API routes** | `internal/api/routes.go` | Route registration and middleware |
| **API handlers** | `internal/api/handlers/<domain>/` | One package per domain (e.g. `vault`, `registry`, `admin`) |
| **Storage / DB** | `internal/storage/`, `internal/storage/sql/` | Repositories, migrations; Postgres + optional Redis |
| **Auth** | `internal/auth/`, `internal/api/middleware/auth.go` | Sessions, GBA plugins (MFA, SAML, WebAuthn) |
| **Dashboard (React)** | `web/dashboard/src/` | Vite SPA; `pages/`, `components/`, `hooks/`, `lib/` |
| **Deploy / edge** | `deploy/`, `deploy/edge/` | Caddy, DNS, VPS/edge scripts |
| **Cloudflare** | `docs/CLOUDFLARE.md`, `deploy/cloudflare/`, `deploy/dns/` | DNS, CDN, R2, Workers, Tunnel, Pages |
| **Docs** | `docs/` | Design and operational docs |
| **Neon Postgres** | `docs/NEON.md` | Optional: use Neon for Postgres; set `DATABASE_URL` or `DB_*` |
| **Local PG 17 + pgvector** | `docs/LOCAL_POSTGRES_17.md` | Migrate local Debian DB to PostgreSQL 17 with pgvector and extensions |

When adding API surface: add handler in `internal/api/handlers/`, register in `internal/api/routes.go`, and use existing storage/auth patterns.

---

## Services overview

| Service | How to run | Port |
|---------|-----------|------|
| **Orchestrator API** (Go) | `./bin/orchestrator-api --skip-migrations` or `make dev` | 8080 |
| **Dashboard** (Vite/React) | `cd web/dashboard && npx vite --host 0.0.0.0` | 3000 |
| **PostgreSQL** | `sudo pg_ctlcluster 17 main start` (see `docs/LOCAL_POSTGRES_17.md` to replace PG 16 with 17) | 5432 |
| **Redis** | `redis-server --daemonize yes` | 6379 |

---

## Starting the backend

1. **Start dependencies** (required first):

   ```bash
   sudo pg_ctlcluster 17 main start
   redis-server --daemonize yes
   ```

2. **Environment:** Ensure `.env` exists (copy from `.env.example` if present). Then:

   ```bash
   source .env
   export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=functionfly DB_SSLMODE=disable
   export REDIS_ADDR=localhost:6379 DEVELOPMENT=true SKIP_MIGRATION_VALIDATION=true VERIFICATION_ENABLED=false
   ./bin/orchestrator-api --skip-migrations
   ```

   If you still have PostgreSQL 16 and want a single cluster, see `docs/LOCAL_POSTGRES_17.md` to replace it with PG 17 on port 5432.

The `--skip-migrations` flag is required because the `migrations/` directory has duplicate sequence numbers that break golang-migrate. Schema is applied via direct SQL during initial setup.

---

## Starting the dashboard

```bash
cd web/dashboard
VITE_API_URL=http://localhost:8080 npx vite --host 0.0.0.0 --port 3000
```

Vite proxies `/api/*` to the Go backend (port 8080). Use `VITE_API_URL` when the API is on a different host/port.

---

## Running tests and lint

- **Go tests:** `go test ./internal/...` — storage tests expect Postgres on port **5432** (may fail if using Docker on 5434).
- **Go lint:** `golangci-lint run` (binary at `~/go/bin/golangci-lint`).
- **Go build:** `go build -o bin/orchestrator-api ./cmd/orchestrator-api`
- **Dashboard:** `cd web/dashboard && npx vitest run` for tests. ESLint has a known broken import (`eslint-import-resolver-typescript` default export); treat as pre-existing.

---

## Known gotchas

1. **Migrations:** Duplicate sequence numbers in `migrations/` (e.g. two `000001_*.sql`). Use `--skip-migrations` when starting the API; apply schema changes via direct SQL if needed.
2. **resend-go:** Upgraded from v2.0.0 to v2.28.0. `ReplyTo` is now `string` (was `*string`); `client.Keys.Get()` was removed.
3. **Stub packages:** `internal/adapters/functionfly`, `internal/api/handlers/statefabric`, `internal/storage/statefabric` — return "not implemented"; do not remove without replacing.
4. **Postgres audit trigger:** `audit_trigger_function()` expects `ip_address` cast to `::inet` (fixed in DB setup).
5. **Admin test account:** `admin@functionfly.local` / `admin123` (bcrypt).
6. **Dashboard:** `index.html` was added for Vite; `bun` is the JS package manager — run `bun install` from repo root for workspace deps.

---

## Vault security (Secrets)

The Secrets Vault is **zero-knowledge**: the server never sees plaintext or the decryption passphrase. Encryption and decryption are done **client-side** (dashboard: `web/dashboard/src/utils/vault-crypto.ts`). The API stores only AES-256-GCM ciphertext + IV/salt/tag. There is **no server-side decrypt endpoint** by design. For audit retention and token cleanup, see `docs/VAULT_OPERATIONS.md`.

---

## Troubleshooting

- **API won’t start:** Ensure Postgres and Redis are running and env vars are set (especially `DB_*`, `REDIS_ADDR`).
- **Dashboard can’t reach API:** Check `VITE_API_URL` and that the orchestrator is on 8080; Vite proxy only applies to `/api/*`.
- **Storage tests fail:** Point DB at port 5432 (local Postgres) or adjust test config to match your DB port.

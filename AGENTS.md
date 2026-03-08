# FunctionFly Development Guide

See `README.md` for project overview and `CONTRIBUTING.md` for Git workflow.

## Cursor Cloud specific instructions

### Services overview

| Service | How to run | Port |
|---------|-----------|------|
| **Orchestrator API** (Go) | `./bin/orchestrator-api --skip-migrations` or `make dev` | 8080 |
| **Dashboard** (Vite/React) | `cd web/dashboard && npx vite --host 0.0.0.0` | 3000 |
| **PostgreSQL** | `sudo pg_ctlcluster 16 main start` | 5432 |
| **Redis** | `redis-server --daemonize yes` | 6379 |

### Starting the backend

PostgreSQL and Redis must be running first. Start them with:
```
sudo pg_ctlcluster 16 main start
redis-server --daemonize yes
```

Then source `.env` and start the orchestrator API:
```
source .env
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=functionfly DB_SSLMODE=disable
export REDIS_ADDR=localhost:6379 DEVELOPMENT=true SKIP_MIGRATION_VALIDATION=true VERIFICATION_ENABLED=false
./bin/orchestrator-api --skip-migrations
```

The `--skip-migrations` flag is required because the `migrations/` directory has duplicate sequence numbers that break golang-migrate. The schema is applied via direct SQL during initial setup.

### Starting the dashboard

```
cd web/dashboard
VITE_API_URL=http://localhost:8080 npx vite --host 0.0.0.0 --port 3000
```

The Vite config proxies `/api/*` requests to the Go backend automatically.

### Running tests and lint

- **Go tests:** `go test ./internal/...` (storage tests may fail if targeting Docker port 5434 instead of local 5432)
- **Go lint:** `golangci-lint run` (binary at `~/go/bin/golangci-lint`)
- **Go build:** `go build -o bin/orchestrator-api ./cmd/orchestrator-api`
- **Dashboard lint:** ESLint config has a broken import (`eslint-import-resolver-typescript` default export) — this is a pre-existing issue
- **Dashboard tests:** `cd web/dashboard && npx vitest run`

### Known gotchas

1. **Migration files have duplicate sequence numbers** (e.g., two `000001_*.sql` files). The `golang-migrate` library rejects this. Use `--skip-migrations` when starting the API, and apply schema changes via direct SQL if needed.
2. **`resend-go` dependency was upgraded** from v2.0.0 (broken module path) to v2.28.0. The `ReplyTo` field changed from `*string` to `string`, and `client.Keys.Get()` was removed.
3. **Three stub packages were created** to fill missing imports: `internal/adapters/functionfly`, `internal/api/handlers/statefabric`, `internal/storage/statefabric`. These return "not implemented" responses.
4. **Audit trigger inet cast**: The `audit_trigger_function()` in PostgreSQL needed the `ip_address` value cast to `::inet`. This was fixed during DB setup.
5. **Admin test account:** `admin@functionfly.local` / `admin123` (created during setup, password set via bcrypt).
6. **No `index.html`** existed in `web/dashboard/` — it was created for Vite to serve the SPA.
7. **bun** is the package manager for JS workspaces (root `bun.lock`). Run `bun install` from the repo root to install all workspace dependencies.

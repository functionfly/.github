# Neon Postgres Setup

This guide walks through using [Neon](https://neon.tech) (serverless Postgres) to manage your FunctionFly database for development, staging, or production.

## Why Neon

- **Serverless**: Scale-to-zero when idle; autoscale under load.
- **Branching**: Create staging/preview branches from production with copy-on-write (see [docs/STAGING.md](STAGING.md)).
- **Compatible**: Standard Postgres; works with existing migrations and tooling.
- **Connection pooling**: Built-in PgBouncer for high connection counts.

## Quick setup

### 1. Create a Neon account and project

1. Sign up at [neon.tech](https://neon.tech).
2. Create a new **Project** (e.g. `functionfly`).
3. Choose region (e.g. US East 1) and Postgres version.
4. Create a database named `functionfly` (or use the default `neondb` and set `DB_NAME` accordingly).

### 2. Get connection details

In the Neon Console → your project → **Connection details**:

- **Host** (endpoint): e.g. `ep-xxx.us-east-1.aws.neon.tech`
- **Port**: `5432`
- **Database**: `functionfly` (or the name you created)
- **User** / **Password**: project role credentials

You can use either:

- **Option A – Connection string (recommended)**  
  Copy the **Connection string** (with password). Use the **pooled** variant for the app (host contains `-pooler`).

- **Option B – Individual env vars**  
  Set `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE=require`.

### 3. Configure FunctionFly

**Option A – `DATABASE_URL` (single variable)**

```bash
# Pooled connection (recommended for API/app) – use the "Connection pooling" toggle in Neon Console
export DATABASE_URL="postgresql://USER:PASSWORD@ep-xxx-pooler.us-east-1.aws.neon.tech/functionfly?sslmode=require"
```

**Option B – `DB_*` variables**

```bash
export DB_HOST=ep-xxx-pooler.us-east-1.aws.neon.tech
export DB_PORT=5432
export DB_USER=functionfly_owner
export DB_PASSWORD=your-neon-password
export DB_NAME=functionfly
export DB_SSLMODE=require
```

If both are set, `DATABASE_URL` takes precedence. Pool settings (`DB_MAX_OPEN_CONNS`, etc.) still apply when using `DATABASE_URL`.

### 4. Run migrations

The app uses `--skip-migrations` by default (see [AGENTS.md](../AGENTS.md)). To apply schema to Neon:

- **With DATABASE_URL**: ensure `DATABASE_URL` is set, then run your migration command (e.g. migrate up via your normal process).
- **With DB_***: ensure `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE=require` are set.

For migrations, use a **direct** (non-pooled) connection if your migration tool has issues with PgBouncer (e.g. session state). In Neon, direct = host **without** `-pooler`. For most cases the pooled endpoint is fine.

**Using the Neon CLI:** From the repo, run `make migrate-neon` to run migrations against the Neon **production** branch and **functionfly** database (uses direct connection). For staging use `NEON_BRANCH=staging make migrate-neon`.

**Updating Fly.io secrets:** After migrations, set Fly secrets (including `DATABASE_URL` from Neon) with:

```bash
./.fly/set-secrets-from-neon.sh production
```

This script gets the pooled connection string from `neon connection-string production --pooled --database-name functionfly`, generates `JWT_SECRET`, `API_SHARED_SECRET`, and `DB_MASTER_KEY_PASSWORD` if not set, and runs `fly secrets set` for **functionfly-control**. Set `REDIS_ADDR` (and optionally `REDIS_PASSWORD`) before running if you use Fly Redis or Upstash; otherwise create Redis with `fly redis create --name functionfly-control-redis` and set the URL.

### 5. Start the API

```bash
source .env
./bin/orchestrator-api --skip-migrations
```

## Connection pooling

Neon provides PgBouncer in front of Postgres:

- **Pooled** (host like `ep-xxx-pooler.region.aws.neon.tech`): use for the orchestrator API and any app that opens many connections. Supports thousands of client connections while using a limited number of real Postgres connections.
- **Direct** (host like `ep-xxx.region.aws.neon.tech`): use for migrations, pg_dump, or tools that need session-level features (e.g. prepared statements across transactions, `SET` persistence). Our app uses protocol-level prepared statements and is fine with the pooler.

In the Neon Console, use the “Connection pooling” toggle when copying the connection string to get the `-pooler` host.

References:

- [Neon: Connection pooling](https://neon.com/docs/connect/connection-pooling)
- [Neon: Connect from any app](https://neon.com/docs/connect/connect-from-any-app)

## Neon CLI

The [Neon CLI](https://neon.com/docs/reference/neon-cli) lets you manage projects, branches, and connection strings from the terminal. Use it to get connection strings, create read replicas, and manage branches (e.g. for staging).

### Install

If the `neon` command is not available:

- **npm** (Node.js 18+): `npm i -g neonctl`
- **Bun**: `bun install -g neonctl`
- **Homebrew** (macOS/Linux): `brew install neonctl`
- **Linux binary**: `curl -sL https://github.com/neondatabase/neonctl/releases/latest/download/neonctl-linux-x64 -o neonctl && chmod +x neonctl` (then move to your PATH)
- **Without installing**: `npx neonctl <command>` or `bunx neonctl <command>`

Check: `neon --version`

From the repo you can run `make neon-install` to install via npm if `neon` is missing.

### Authenticate

- **Web (interactive)**: run `neon auth` (or `make neon-auth`). A browser opens to log in to Neon and authorize the CLI.
- **API key (CI/scripts)**: create an API key in [Neon Console → Account → API Keys](https://console.neon.tech/app/settings/api-keys), then set `export NEON_API_KEY=your_key` or pass `--api-key` to commands.

### Set context (optional)

To default to a specific project (and org) so you don't pass `--project-id` every time:

```bash
neon set-context --project-id <project_id> --org-id <org_id>
```

Get project/org IDs from the Neon Console URL or `neon projects list`.

### Common commands

| Goal | Command |
|-----|--------|
| Connection string (default branch) | `neon connection-string` or `neon cs` |
| Pooled connection string | `neon cs --pooled` |
| Connection string for a branch | `neon cs <branch>` or `neon cs <branch> --pooled` |
| Read replica connection string | `neon cs <branch> --endpoint-type read_only` (or `--pooled`) |
| List branches | `neon branches list` |
| Add read replica to a branch | `neon branches add-compute <branch> --type read_only` |
| Open psql with connection string | `neon cs --psql` |

Use `make neon-cs` for the primary connection string and `make neon-cs-pooled` for the pooled one (requires auth and optional `NEON_PROJECT_ID` / context).

References:

- [Neon CLI: Install and connect](https://neon.com/docs/reference/cli-install)
- [Neon CLI: connection-string](https://neon.com/docs/reference/cli-connection-string)
- [Neon CLI: branches](https://neon.com/docs/reference/cli-branches)

## Neon with read replicas (hybrid)

You can run **primary (read–write) + read replicas** so the API uses the primary for writes and sends read traffic to replica(s), with fallback to the primary if no replica is healthy. Neon read replicas use the same storage as the primary (no extra data copy) and support scale-to-zero and autoscaling.

### 1. Create a read replica in Neon

- **Console**: Neon Console → **Branches** → select your branch → **Add Read Replica**. Configure compute size and optional scale-to-zero.
- **CLI**: `neon branches add-compute <branch> --type read_only`
- **API**: [Create endpoint](https://api-docs.neon.tech/reference/createprojectendpoint) with `"type": "read_only"`.

Free plan: max 3 read replica computes per project. See [Neon: Create and manage Read Replicas](https://neon.com/docs/guides/read-replica-guide).

### 2. Get the read replica connection host

In Neon Console → **Connect** → choose branch, database, and role. Under **Compute**, select a **Replica** (not the primary). Copy the host (e.g. `ep-xxx-readonly.us-east-1.aws.neon.tech`). For many connections, enable **Connection pooling** and use the `-pooler` host. Replicas use the same user/password/database as the primary.

### 3. Configure FunctionFly for hybrid (primary + replicas)

Set the **primary** as in [Quick setup](#quick-setup) (e.g. `DATABASE_URL` or `DB_*`). Then enable read replicas and add at least one replica host:

```bash
# Primary (already set)
export DATABASE_URL="postgresql://USER:PASSWORD@ep-xxx-pooler.us-east-1.aws.neon.tech/functionfly?sslmode=require"
# Or use DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE=require

# Enable read replicas and add replica endpoint(s)
export DB_READ_REPLICA_ENABLED=true
export DB_REPLICA_1_HOST=ep-xxx-readonly-pooler.us-east-1.aws.neon.tech
export DB_REPLICA_1_PORT=5432
# Optional: DB_REPLICA_1_WEIGHT=1, DB_REPLICA_1_PRIORITY=1, DB_REPLICA_1_REGION=us-east-1
# Add more replicas with DB_REPLICA_2_HOST, DB_REPLICA_3_HOST, ... (up to 5)
```

Replicas use the same **user**, **password**, **database**, and **sslmode** as the primary (from `DATABASE_URL` or `DB_*`). Only replica host/port (and optional weight/priority/region) are configured via `DB_REPLICA_*`.

### 4. Behavior

- **Writes** and **transactions** always use the primary.
- **Reads** use a healthy replica when `DB_READ_REPLICA_ENABLED=true` and at least one replica is configured and healthy; otherwise the API falls back to the primary.
- Replicas are **eventually consistent**; short replication delay is normal (see [Neon: Monitoring read replicas](https://neon.com/docs/guides/read-replica-guide#monitoring-read-replicas)).

References:

- [Neon: Read replicas overview](https://neon.com/docs/introduction/read-replicas)
- [Neon: Create and manage Read Replicas](https://neon.com/docs/guides/read-replica-guide)

## Staging and branching

You can use Neon branches for staging (see [STAGING.md](STAGING.md)):

- Create a branch from your main branch in the Neon Console (or via [Neon CLI](https://neon.com/docs/reference/neon-cli)).
- Use that branch’s connection details for staging env (different `DB_HOST` or `DATABASE_URL`).
- Run the same migrations against the staging branch.

## Connection limits and pool sizing

Neon limits Postgres connections by compute size (e.g. 0.25 CU ≈ 97 user connections). With pooling, many more clients can connect. Tune app-side pool size so you don’t queue unnecessarily:

- `DB_MAX_OPEN_CONNS`: e.g. 20–50 for a single API instance.
- `DB_MAX_IDLE_CONNS`: e.g. 5–10.

See [Neon connection pooling](https://neon.com/docs/connect/connection-pooling) and your plan’s compute size for exact limits.

## Security

- Use **SSL**: Neon requires TLS; set `DB_SSLMODE=require` or use a connection string with `?sslmode=require`.
- Store credentials in env or a secret manager (e.g. Infisical); never commit them.
- Optional: restrict access with [Neon IP allow list](https://neon.com/docs/introduction/ip-allow).

## Troubleshooting

| Issue | What to check |
|-------|----------------|
| Connection timeout | Region vs app location; Neon compute may be suspended (scale-to-zero) – first request can be slow. |
| “Too many connections” | Use the **pooled** endpoint and/or lower `DB_MAX_OPEN_CONNS`. |
| Migrations fail | Try the **direct** (non-pooler) connection string for the migration step. |
| “Relation does not exist” after SET | With pooler, session `SET` (e.g. `search_path`) is not persistent; use direct connection or set at role level. |

More: [Neon connection errors](https://neon.com/docs/connect/connection-errors), [Neon connection latency](https://neon.com/docs/connect/connection-latency).

## References

- [Neon docs index](https://neon.com/docs/llms.txt)
- [Neon: Get started](https://neon.com/docs/get-started/signing-up)
- [Neon: Connect from any app](https://neon.com/docs/connect/connect-from-any-app)
- [Neon: Connection pooling](https://neon.com/docs/connect/connection-pooling)
- [FunctionFly STAGING.md](STAGING.md) – staging env and Neon branches
- [FunctionFly PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md) – production env vars and deployment

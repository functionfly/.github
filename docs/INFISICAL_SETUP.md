# Infisical setup (canonical secrets)

This guide describes how **Infisical** fits into FunctionFly: it is the **canonical place to create and edit** secrets for `dev`, `staging`, and `prod`. Runtime on **Fly.io** still uses **Fly secrets** (synced from Infisical; Fly does not read Infisical directly). Local development uses **`infisical run`**, a **service token**, or a **gitignored `.env`** file—never commit real secrets.

## Roles: one mental model

| Layer | Role |
|--------|------|
| **Infisical** | Where humans and automation **edit** secrets per environment |
| **Fly secrets** | What Fly Machines **receive** at runtime (set via `fly secrets` after sync) |
| **`.env`** (gitignored) | Optional **local cache** (`infisical export`) or overrides; not a second source of truth |

**GitHub Actions** may keep a small set of repository secrets (e.g. `DOCKER_*`, `FLY_API_TOKEN`, `CLOUDFLARE_*`) for CI; those can mirror values stored in Infisical for operational simplicity.

## Local development

### Authentication

- **Interactive:** `infisical login` (browser).
- **CI / scripts:** create a **service token** in the Infisical dashboard (read access to the right project and environment) and set `INFISICAL_TOKEN`.

### Project linkage

The repo may contain [`.infisical.json`](../.infisical.json) with a `workspaceId`. If you **fork the repo** or use a **different Infisical organization**, run `infisical init` in the repo root and commit the updated file, or rely on `INFISICAL_PROJECT_ID` / dashboard-linked project without relying on a stale workspace id.

### Running the API and tools

Makefile targets prefer Infisical when the CLI is installed **and** `INFISICAL_TOKEN` is set; otherwise they use `DB_*` defaults or a sourced `.env`:

```bash
make dev              # Orchestrator with Infisical (dev) or local defaults
make api              # Same pattern
make health-monitor
make setup            # Sources .env if present, then Infisical or plain DB_*
```

**Without Infisical** (403, offline, or no token): copy [`.env.example`](../.env.example) to `.env`, fill in values, and use targets such as `make api-local`, `make setup-local`, or `make migrate-local`.

**`make setup`** runs [`scripts/run-setup.sh`](../scripts/run-setup.sh): it tries `infisical run` when the CLI and `INFISICAL_TOKEN` are set; on failure (e.g. **403**), it **falls back** to `DB_*` from your shell / `.env`. To **skip Infisical** entirely: `SKIP_INFISICAL=1 make setup`.

### Frontend

Dashboard and site read `VITE_*` at build time. Store those in Infisical for your environment or local `.env` under `web/dashboard` / `web/site` (see each app’s `.env.example`).

## Sync Infisical → Fly.io

After you change production (or staging) secrets in Infisical, push them to the Fly app:

```bash
# From repo root; requires infisical CLI + fly CLI + INFISICAL_TOKEN (or login)
export INFISICAL_ENV=prod          # or staging
export FLY_APP=functionfly-control
./scripts/sync-infisical-to-fly.sh

# Stage only (no immediate rollout); machines pick up on next deploy/restart
STAGE=1 ./scripts/sync-infisical-to-fly.sh
```

The script exports Infisical secrets as dotenv, keeps only keys listed in [`scripts/fly-secrets-allowlist.txt`](../scripts/fly-secrets-allowlist.txt) (plus a hard block on `VITE_*` and `INFISICAL_*`), and runs `fly secrets import`. To allow additional keys, edit the allowlist file.

Override allowlist path: `FLY_SECRETS_ALLOWLIST=/path/to/allowlist.txt`.

**EU / self-hosted Infisical:** set the API URL so the CLI hits the right host, for example:

```bash
export INFISICAL_API_URL=https://eu.infisical.com/api
# or: infisical --domain https://eu.infisical.com ...
```

See also [FLY_SECRETS_FROM_ENV.md](FLY_SECRETS_FROM_ENV.md) for manual `fly secrets` flows when you are not using Infisical.

## Production: service tokens

1. Open the [Infisical dashboard](https://app.infisical.com) (or your EU/self-hosted URL).
2. Select the FunctionFly project.
3. **Service Tokens** → create a token with read access to the target environment (`prod`, etc.).

Use that token as `INFISICAL_TOKEN` in automation, Docker ([`docker-compose.yml`](../docker-compose.yml)), or CI.

### Optional secrets (examples)

- **Open Router** (`OPENROUTER_API_KEY`) — Admin Registry “Generate with AI” for descriptions.
- **Stripe** (`STRIPE_SECRET_KEY`, etc.) — billing and portal; see [Stripe docs](https://stripe.com/docs).

## Docker

1. Copy `.env.docker` to `.env` if you use that flow, and set `INFISICAL_TOKEN=...`.
2. `docker compose up` — compose passes `INFISICAL_TOKEN` into services that need it.

## CI/CD

```yaml
env:
  INFISICAL_TOKEN: ${{ secrets.INFISICAL_TOKEN }}
```

Prefer narrow-scoped tokens and environments. For **syncing to Fly** from CI, use the manual workflow [`.github/workflows/sync-infisical-to-fly.yml`](../.github/workflows/sync-infisical-to-fly.yml) (set repository secrets `INFISICAL_TOKEN` and `FLY_API_TOKEN`), or run [`scripts/sync-infisical-to-fly.sh`](../scripts/sync-infisical-to-fly.sh) locally so logs never print secret values.

## Troubleshooting

### `403 Forbidden` / “You are not allowed to access this resource”

Often the CLI is pointed at the **wrong Infisical instance** or the token cannot see the workspace:

- Use **`INFISICAL_API_URL`** (or `--domain`) for **EU**: `https://eu.infisical.com/api` vs US default `https://app.infisical.com/api`.
- Confirm the **service token** is for the correct **project** and **environment**, and has not expired.
- If the repo’s `.infisical.json` references another team’s workspace, run **`infisical init`** or fix `workspaceId` / `INFISICAL_PROJECT_ID`.

### “You must be logged in”

Run `infisical login`, or set `INFISICAL_TOKEN`.

### “Missing credentials” / encryption key errors

Run commands from the directory that contains `.infisical.json`, or re-run `infisical init`.

### Wrong environment

Match `--env=dev` (or `prod`) to the Infisical environment name.

### Secrets not loading in Make

Check `INFISICAL_TOKEN`, environment name, and that variable names match exactly between Infisical and the app.

## References

- [Infisical docs](https://infisical.com/docs)
- [Fly secrets](https://fly.io/docs/reference/secrets/) — runtime store after sync
- [FLY_SECRETS_FROM_ENV.md](FLY_SECRETS_FROM_ENV.md) — manual import / exceptions

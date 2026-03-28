# Fly.io secrets (runtime)

Fly Machines read configuration from **Fly secrets** (and non-secret `[env]` in `fly.toml`). Your **gitignored `.env`** is for local development only—it is not automatically available on Fly.

## Default path: Infisical → Fly

**Recommended:** maintain secrets in **Infisical** per environment, then sync to Fly:

```bash
INFISICAL_ENV=prod FLY_APP=functionfly-control ./scripts/sync-infisical-to-fly.sh
```

Details, allowlist, and EU/self-hosted Infisical URLs: [INFISICAL_SETUP.md](INFISICAL_SETUP.md).

## What belongs on Fly

- **Set on Fly:** orchestrator/API variables — `DATABASE_URL` or `DB_*`, `REDIS_*`, `JWT_SECRET`, `API_SHARED_SECRET`, `BASE_URL`, `CORS_ALLOWED_ORIGINS`, OAuth client IDs/secrets, etc. (see [`.fly/secrets.example`](../.fly/secrets.example)).
- **Do not set on Fly:** frontend-only `VITE_*`, `INFISICAL_*`, local-only paths, or dev flags you do not want in production.

## Exceptions: manual `fly secrets`

Use these when you are not syncing from Infisical (e.g. one-off debugging or a minimal app).

### One-off: `fly secrets set`

```bash
fly secrets set \
  GITHUB_CLIENT_ID="your-id" \
  GITHUB_CLIENT_SECRET="your-secret" \
  BASE_URL="https://api.functionfly.com" \
  --app functionfly-control
```

Use quotes if values contain spaces or special characters.

### Bulk: `fly secrets import`

The CLI reads **`NAME=VALUE`** lines from stdin (one per line).

1. Create a file (e.g. `fly-secrets.env`) with **only** `KEY=value` lines—no `export`, no comments, or filter them out.
2. Import:

   ```bash
   fly secrets import --app functionfly-control < fly-secrets.env
   ```

3. Do not commit that file; keep it gitignored.

### Stage without immediate rollout

```bash
fly secrets set KEY=value --app functionfly-control --stage
# or
fly secrets import --app functionfly-control --stage < fly-secrets.env
```

Machines pick up staged secrets on the **next** deploy or restart.

### Verify

```bash
fly secrets list --app functionfly-control
```

Values are redacted; only names are listed.

## Repo helpers

- [`scripts/sync-infisical-to-fly.sh`](../scripts/sync-infisical-to-fly.sh) — Infisical export → allowlist → `fly secrets import`.
- [`.fly/set-secrets.sh`](../.fly/set-secrets.sh) — scripted `fly secrets set` from variables you edit in the script.
- [`.fly/set-secrets-from-neon.sh`](../.fly/set-secrets-from-neon.sh) — Neon `DATABASE_URL` plus generated auth secrets and URLs.

## Loading `.env` then setting Fly (careful)

You can `set -a && source .env && set +a` and then:

```bash
fly secrets set GITHUB_CLIENT_ID="$GITHUB_CLIENT_ID" GITHUB_CLIENT_SECRET="$GITHUB_CLIENT_SECRET" --app functionfly-control
```

Only pass keys you intend to send; avoid piping your entire laptop `.env` unless every variable is safe for that Fly app.

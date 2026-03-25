# Push local env to Fly secrets

Your **`.env`** is for local development. **Fly Machines** get configuration from **Fly secrets** (or `[env]` in `fly.toml` for non-sensitive values). They are separate; you copy or import the values you need.

## 1. What belongs on Fly

- **Set on Fly:** API/orchestrator variables — `DATABASE_URL` or `DB_*`, `REDIS_*`, `JWT_SECRET`, `API_SHARED_SECRET`, `BASE_URL`, `CORS_ALLOWED_ORIGINS`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GOOGLE_*`, etc.
- **Do not set on Fly:** Frontend-only `VITE_*`, local paths, `DEVELOPMENT=true` for production, or secrets you only use on your laptop.

## 2. One-off: copy values into `fly secrets set`

From your `.env`, copy values and run (replace with your app name if different):

```bash
fly secrets set \
  GITHUB_CLIENT_ID="your-id" \
  GITHUB_CLIENT_SECRET="your-secret" \
  BASE_URL="https://api.functionfly.com" \
  --app functionfly-control
```

Repeat or combine multiple keys in one `fly secrets set` line. Use quotes if values contain spaces or special characters.

## 3. Bulk: `fly secrets import` from a file

The CLI reads **`NAME=VALUE` pairs from stdin** (one per line).

1. Create a **dedicated file** (e.g. `fly-secrets.env`) containing **only** lines you want on Fly — **no** `export`, **no** comments, **no** blank lines unless you filter them:

   ```bash
   GITHUB_CLIENT_ID=abc123
   GITHUB_CLIENT_SECRET=secretvalue
   BASE_URL=https://api.functionfly.com
   ```

2. Import:

   ```bash
   fly secrets import --app functionfly-control < fly-secrets.env
   ```

3. **Do not commit** `fly-secrets.env`; add it to `.gitignore` if you keep it on disk.

**Note:** If your `.env` uses `export KEY=value` or quotes, fix the file first or use `grep`/`sed` to normalize. Invalid lines will cause errors.

## 4. Stage without deploying immediately

```bash
fly secrets set KEY=value --app functionfly-control --stage
# or
fly secrets import --app functionfly-control --stage < fly-secrets.env
```

Secrets are stored; machines pick them up on the **next** deploy or restart.

## 5. Verify

```bash
fly secrets list --app functionfly-control
```

Values are **not** shown (redacted); you only see names.

## 6. Repo scripts (alternative)

- **`.fly/set-secrets.sh`** — edit variables inside the script, then run `./.fly/set-secrets.sh production functionfly-control`.
- **`.fly/set-secrets-from-neon.sh`** — Neon `DATABASE_URL` + generated secrets + `CORS_*` / `BASE_URL` / `FRONTEND_URL`.

## 7. Loading `.env` in your shell then setting (careful)

You can **export** vars in your shell from `.env` (e.g. `set -a && source .env && set +a`) and then:

```bash
fly secrets set GITHUB_CLIENT_ID="$GITHUB_CLIENT_ID" GITHUB_CLIENT_SECRET="$GITHUB_CLIENT_SECRET" --app functionfly-control
```

Only do this for keys you intend to send; avoid pushing the whole environment in one command unless you know every variable is safe for production Fly.

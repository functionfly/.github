# FunctionFly public documentation site (`web/docs`)

Static site (Astro) deployed at `docs.functionfly.com` (or your `PUBLIC_SITE_URL`).

## What gets published

- **Hand-written** pages live in `src/content/docs/` as `cli.md`, `getting-started.md`, `functions.md`, and `deployment.md`. These are not removed by sync.
- **Synced** pages come only from an explicit **public allowlist** in [`scripts/sync-docs.mjs`](scripts/sync-docs.mjs). The script copies matching files from the repo [`docs/`](../docs/) folder and **deletes** any previously synced file that is no longer allowlisted.

Internal material (production/staging runbooks, disaster recovery, monitoring, vault operations, email/DNS setup, etc.) must stay in `docs/` but **must not** be added to the allowlist without security/legal review.

## Build

```bash
cd web/docs
bun run build   # runs sync-docs.mjs, then astro build
```

## Relationship to the dashboard

The React app under `src/App.tsx` / `src/legacy-vite/` is legacy; the shipped site is Astro. In-app documentation for logged-in users lives in `web/dashboard`.

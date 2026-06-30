# FunctionFly public documentation site (`web/docs`)

Static site (Astro + Starlight) deployed at `docs.functionfly.com` (or your `PUBLIC_SITE_URL`).

## What gets published

- **Hand-written** pages live in `src/content/docs/` as `cli.md`, `getting-started.md`, `functions.md`, and `deployment.md`. These are not removed by sync.
- **Synced** pages come only from an explicit **public allowlist** in [`scripts/sync-docs.mjs`](scripts/sync-docs.mjs). The script copies matching files from the repo [`docs/`](../docs/) folder and **deletes** any previously synced file that is no longer allowlisted.

Internal material (production/staging runbooks, disaster recovery, monitoring, vault operations, email/DNS setup, etc.) must stay in `docs/` but **must not** be added to the allowlist without security/legal review.

## Build

```bash
cd web/docs
bun install
bun run build   # runs sync-docs.mjs, then astro build
```

## Development

```bash
bun run dev     # port 4322
```

## Architecture

- **Framework:** Astro v6 + Starlight (static site generation)
- **API Reference:** Auto-generated from `src/content/api/openapi.yaml` via starlight-openapi
- **Custom CSS:** `src/styles/custom.css` (Starlight theme override, self-contained)
- **Security:** `src/middleware.ts` adds security headers (X-Frame-Options, HSTS, etc.)
- **Analytics:** Google Analytics and Mixpanel via env vars (`PUBLIC_GOOGLE_ANALYTICS_ID`, `PUBLIC_MIXPANEL_TOKEN`)
- **Error Tracking:** Sentry (conditional on `SENTRY_DSN` env var)

## Environment Variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `PUBLIC_GOOGLE_ANALYTICS_ID` | Google Analytics tracking ID | No |
| `PUBLIC_MIXPANEL_TOKEN` | Mixpanel project token | No |
| `SENTRY_DSN` | Sentry error tracking DSN | No |

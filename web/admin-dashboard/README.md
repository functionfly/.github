# FunctionFly Admin Dashboard

Standalone React/Vite SPA that powers `admin.functionfly.com`. It provides
platform-level administration (tenants, users, billing, factory, trust & safety,
content, infra) and is hardened for direct internet exposure:

- Server-issued HMAC-SHA256 request signing (no shared secret in the bundle)
- Session-bound JWT in `sessionStorage` (no cross-tab `localStorage` bootstrap)
- MFA challenge + 4-hour re-verification
- Cloudflare Zero Trust header detection (configurable)
- IP allowlist, SIEM, fraud detection, audit log surfaces
- Content Security Policy, HSTS, COOP/COEP/CORP, MIME-sniff protection, secure
  iframe headers (see `nginx.conf` for the production web-server config)

## Quickstart

```bash
cd web/admin-dashboard
bun install                    # or npm install

# 1. Pull env vars (one-time, requires Vercel link):
# vercel env pull

# 2. Start the dev server (default port 3002):
bun run dev
# Server: http://localhost:3002

# 3. Type-check + build (CI gate):
bun run type-check              # tsc --noEmit
bun run build                   # tsc && vite build

# 4. Tests
bun run test                    # vitest (unit)
bun run e2e                     # playwright (e2e)
bun run lint                    # eslint
```

The Vite dev server proxies `/api/*` to the URL in `VITE_API_BASE_URL`. If that
variable is unset, the dev server throws at startup so we fail loud rather than
silently pointing at a stale default.

## Environment

Copy `.env.example` to `.env.production` (or use `vercel env pull`):

| Variable | Purpose | Default |
|----------|---------|---------|
| `VITE_API_BASE_URL` | Backend origin (proxied in dev) | `https://api.functionfly.com` |
| `VITE_ADMIN_API_BASE_URL` | Admin API base | `${VITE_API_BASE_URL}/v1/admin` |
| `VITE_SESSION_TIMEOUT` | Absolute session TTL (ms) | `1800000` (30 min) |
| `VITE_IDLE_TIMEOUT` | Idle timeout (ms) | `900000` (15 min) |
| `VITE_MFA_REVERIFY_INTERVAL` | How often MFA must be re-confirmed (ms) | `14400000` (4 h) |
| `VITE_ENABLE_IP_WHITELIST` | Show IP allowlist UI / enforce | `true` (prod) / `false` (dev) |
| `VITE_ENABLE_DEVICE_FINGERPRINT` | Emit device fingerprint events | `true` (prod) / `false` (dev) |
| `VITE_ENABLE_AUDIT_LOGGING` | Emit client audit events | `true` |
| `VITE_ENABLE_SESSION_RECORDING` | Post-hoc session replay | `false` |
| `VITE_EXPECT_ZT_HEADERS` | Require Cloudflare ZT headers | `true` (prod) / `false` (dev) |
| `VITE_DEVELOPMENT` | Dev-mode flag for client logic | `false` (prod) / `true` (dev) |

> **Security:** Never set `VITE_ADMIN_SHARED_SECRET` (or any other secret) as a
> `VITE_*` variable — those values are inlined into the browser bundle and
> would leak immediately. The HMAC secret lives only on the server as
> `API_SHARED_SECRET`; the admin SPA obtains short-lived signatures via
> `POST /v1/admin/auth/sign-request`.

`.env.development` provides the local defaults so `bun run dev` works without
copying any file. CI sets `VITE_API_BASE_URL=https://example.invalid` as a
build-time placeholder; the runtime config is the only thing that matters
there.

## Routes

All admin pages live inside `src/pages/`. The single source of truth for
routing is `src/routes/adminRoutes.tsx`. `src/App.tsx` only wires up
providers, the auth-restore gate, the layout, and the public auth pages.

Each route is permission-gated via `<AdminPage permission="..." />` — the
component redirects to `/access-denied` if the signed-in admin lacks the
permission. Page-level code is still expected to verify destructive actions
are HMAC-signed by the client interceptor (see
`src/lib/api/hmacSigner.ts`).

## Architecture notes

| Concern | Where |
|--------|-------|
| Auth state | `src/stores/adminAuthStore.ts` (Zustand) |
| Session restore | `src/components/auth/AdminAuthRestore.tsx` |
| Permission checks | `src/hooks/useAccessControl.ts` + `src/components/auth/AdminPage.tsx` |
| CSRF token | `src/lib/security/csrf.ts` |
| HMAC request signing | `src/lib/api/hmacSigner.ts` |
| Zero-Trust gate | `src/lib/security/zeroTrust.ts` |
| API client + interceptors | `src/lib/api/adminClient.ts` |
| React Query setup | `src/App.tsx` (`QueryClient` defaults) |
| Centralized logger | `src/lib/monitoring/logger.ts` (silenced in prod) |
| Constants | `src/lib/constants.ts` (routes + API paths) |

## Deployment

### Vercel (production target)

```bash
cd web/admin-dashboard
bun run vercel:deploy          # production
bun run vercel:deploy:preview  # preview
```

**Hostname:** `admin.functionfly.com`

Custom-domain steps are in `docs/STANDALONE_ADMIN_DASHBOARD.md`. The Vercel
project's `vercel.json` (or `vercel.ts`) and the `Dockerfile`/`nginx.conf`
are alternative deployment targets — see below.

### Docker / nginx

```bash
docker build -t functionfly-admin-dashboard .
docker run -p 80:80 functionfly-admin-dashboard
```

The multi-stage `Dockerfile` builds with Node 20-alpine and ships via the
unprivileged `nginx` user. `nginx.conf` is hardened: HSTS, CSP, COOP/COEP,
distinct caching for hashed assets vs. HTML, blocked access to `.git`/editor
swap files, and SPA fallback to `/index.html`.

### CI

`.github/workflows/admin-dashboard-ci.yml` runs on PR/push when files under
`web/admin-dashboard/` change: `typecheck`, `build`, `unit-tests`, and
`lint` (lint is currently `continue-on-error: true` while we work through
pre-existing rule issues).

## Testing

- **Unit** — `bun run test` (Vitest). Tests live next to source as
  `*.test.ts` / `*.test.tsx`. Use `// @vitest-environment happy-dom` for any
  test that touches `window`/`sessionStorage`/`document`.
- **E2E** — `bun run e2e` (Playwright). Specs live in `e2e/`. They mock the
  backend with `page.route` so they don't need a live API.
- **Smoke** — `bun run build && bun run preview`, then open
  `http://localhost:4173/auth/login`.

## Security model (TL;DR)

- **Auth**: JWT in `sessionStorage`, set on successful login or after a valid
  MFA challenge. A page refresh in the same tab restores the session via
  `POST /v1/admin/auth/session`. A new tab requires a fresh admin login
  (this is intentional — see `AdminAuthRestore` for the rationale).
- **CSRF**: Double-submit token. The client fetches `/csrf` once and
  includes the token in `X-CSRF-Token` on mutating requests.
- **HMAC**: For mutating requests that need elevated trust (e.g. PATCH on
  `/security/measures`), the client calls `/auth/sign-request` with a hash
  of `(method, path, body)`, then sends `X-Signature` / `X-Signature-Ts`
  / `X-Signature-Nonce` headers. Signatures are cached for 5 min and bound
  to the active session.
- **Zero Trust**: When `VITE_EXPECT_ZT_HEADERS=true`, the client checks for
  the `CF_Authorization` cookie before mounting protected pages. The check
  is purely a UX redirect — the API also re-verifies these headers server-side.
- **Headers**: All page HTML is rendered with the strict CSP/COOP/COEP
  headers from `nginx.conf`. Third-party script loading is disabled.

## Logging

Use `logger` from `@/lib/monitoring/logger` instead of `console.*`. The
logger is silenced in production builds (`import.meta.env.PROD === true`)
and proxies to `console.*` in dev. Future Sentry/Datadog sinks can be
registered via `registerLogSink()`.

## Project structure

```
src/
├── pages/            # One file per route (lazy-loaded)
├── components/       # Layout, ui, auth, factory, security
├── stores/           # Zustand stores
├── hooks/            # Custom React hooks
├── lib/
│   ├── api/          # adminClient, hmacSigner, factory api
│   ├── security/     # csrf, zeroTrust
│   ├── monitoring/   # logger, security events
│   ├── routing/      # admin route table
│   └── constants.ts  # ROUTES, API_ROUTES, AUDIT_EVENT_TYPES, MFA policies
├── routes/           # adminRoutes.tsx (single source of truth for routes)
├── types/            # TypeScript types
└── App.tsx           # Providers + auth gate + layout
e2e/                  # Playwright specs
.github/workflows/    # admin-dashboard-ci.yml
Dockerfile, nginx.conf, vercel.json
```

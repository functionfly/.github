# Cloudflare integration across FunctionFly

This document is the central reference for **all Cloudflare usage** on the platform: DNS, CDN, R2 storage, Workers deployments, optional Tunnel/Pages, and security (WAF). Use it to configure production and staging with Cloudflare end-to-end.

---

## Overview

| Area | Purpose | Docs / config |
|------|--------|----------------|
| **DNS** | Hostnames, GeoDNS, health checks | `deploy/dns/cloudflare-geo-dns.json`, `scripts/update-cloudflare-dns.sh`, `.fly/setup-dns.sh` |
| **CDN** | Static assets (SDK, docs) at cdn.* | `docs/CDN_SETUP.md`, `CACHE_CDN_*` env |
| **R2** | Object storage (uploads, optional CDN origin) | `docs/OBJECT_STORAGE.md`, `STORAGE_BACKEND=r2` |
| **Workers** | Deploy user functions to Cloudflare Workers | `internal/adapters/cloudflare/`, backend provider `workers` |
| **Tunnel** (optional) | Expose API/dashboard without opening ports | `deploy/cloudflare-tunnel.example.sh` |
| **Pages** (optional) | Host dashboard/docs | This doc |
| **WAF / proxy** | DDoS, bot protection, SSL | Cloudflare dashboard; enable proxy (orange cloud) |

---

## Environment variables (Cloudflare)

Set these where the orchestrator or CI runs. Prefer a secrets manager (e.g. Infisical) in production.

### DNS / API automation

| Variable | Required for | Description |
|----------|----------------|-------------|
| `CLOUDFLARE_ACCOUNT_ID` | Workers, API | Account ID from Cloudflare dashboard URL |
| `CLOUDFLARE_ZONE_ID` | DNS script, CI | Zone ID for your domain (e.g. functionfly.com) |
| `CLOUDFLARE_API_TOKEN` | DNS script, Workers, CI | API token with needed permissions (Zone:DNS:Edit, Account:Workers:Edit, etc.) |

Used by: `scripts/update-cloudflare-dns.sh`, CI/CD (e.g. `.github/workflows/ci-cd.yml`), and the Cloudflare Workers adapter when deploying functions.

### CDN (static assets)

| Variable | Description |
|----------|-------------|
| `CACHE_CDN_ENABLED` | Set to `true` to use CDN base URL for SDK/docs links |
| `CACHE_CDN_PROVIDER` | `cloudflare` (default), or `cloudfront` / `fastly` |
| `CACHE_CDN_BASE_URL` | e.g. `https://cdn.functionfly.com` or `https://cdn.staging.functionfly.com` |
| `CACHE_CDN_MAX_AGE` | Cache TTL in seconds (e.g. `86400`) |

See **CDN setup** below and `docs/CDN_SETUP.md`.

### R2 (object storage)

| Variable | Required | Description |
|----------|----------|-------------|
| `STORAGE_BACKEND` | Yes | Set to `r2` |
| `STORAGE_BUCKET` | Yes | R2 bucket name |
| `R2_ACCOUNT_ID` | Yes | Cloudflare account ID |
| `R2_ACCESS_KEY_ID` | Yes* | R2 API token access key (or use `AWS_ACCESS_KEY_ID`) |
| `R2_SECRET_ACCESS_KEY` | Yes* | R2 API token secret (or use `AWS_SECRET_ACCESS_KEY`) |
| `R2_ENDPOINT` | No | Override; usually auto from `R2_ACCOUNT_ID` |
| `R2_PUBLIC_URL` | No | Custom domain for public links (e.g. `https://cdn.functionfly.com`) |

\* For R2, you can use either `R2_ACCESS_KEY_ID`/`R2_SECRET_ACCESS_KEY` or `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` (same values from the R2 API token). Prefer R2-prefixed vars when using only Cloudflare.

See `docs/OBJECT_STORAGE.md`.

### Workers (deploying functions)

Configured per-backend in the app/registry. The adapter expects in `provider_config`:

- `account_id` – Cloudflare account ID  
- `api_token` – API token with Workers Scripts + Routes permissions  
- `script_name` – Worker script name  
- `zone_id` – (for route binding) Zone ID for custom domains  
- `workers_subdomain` – (for blue/green) Your Workers subdomain (e.g. `mycompany` for `mycompany.workers.dev`); required for correct CNAME target when using workers.dev with blue/green deployments  

No global env vars are required in the orchestrator for Workers; credentials are stored with the backend.

---

## DNS

- **Config:** `deploy/dns/cloudflare-geo-dns.json` describes records and GeoDNS. Replace placeholders `${CLOUDFLARE_ACCOUNT_ID}`, `${CLOUDFLARE_ZONE_ID}` (or inject when applying).
- **Script:** From repo root, run:
  ```bash
  export CLOUDFLARE_API_TOKEN=… CLOUDFLARE_ZONE_ID=…
  ./scripts/update-cloudflare-dns.sh apply   # basic records
  ./scripts/update-cloudflare-dns.sh geodns  # GeoDNS (if using Fly regions)
  ./scripts/update-cloudflare-dns.sh verify # list records
  ```
  
  Or use the production DNS setup script in `.fly/`:
  ```bash
  .fly/setup-dns.sh --apply --zone ZONE_ID --token TOKEN
  .fly/setup-dns.sh --env staging --apply --zone ZONE_ID --token TOKEN
  ```
- **Proxy:** Use Cloudflare proxy (orange cloud) for public hostnames to get DDoS protection, caching, and WAF. Use DNS-only (grey cloud) for internal or health-check targets if needed.

---

## CDN setup

1. **Origin:** Either Caddy (reverse proxy to orchestrator) or **R2** as CDN origin.  
   - Caddy: point `cdn.*` CNAME to the host running Caddy; Caddy serves from orchestrator (see `deploy/caddy/Caddyfile`).  
   - R2: create bucket, add custom domain `cdn.functionfly.com`, set CNAME to the R2 target; sync SDK/docs/static into the bucket (see `docs/CDN_SETUP.md`).
2. **Orchestrator:** Set `CACHE_CDN_ENABLED=true`, `CACHE_CDN_PROVIDER=cloudflare`, `CACHE_CDN_BASE_URL=https://cdn.functionfly.com` (or staging URL).
3. **Cloudflare:** Enable proxy for the CDN hostname for caching and protection.

---

## R2 as object storage

1. In Cloudflare dashboard: R2 → Create bucket (e.g. `functionfly-uploads`).  
2. R2 → Manage R2 API Tokens: create a token with Object Read & Write; copy Access Key ID and Secret.  
3. Set `STORAGE_BACKEND=r2`, `STORAGE_BUCKET=…`, `R2_ACCOUNT_ID=…`, and either `R2_ACCESS_KEY_ID`/`R2_SECRET_ACCESS_KEY` or `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY`.  
4. Optional: attach a custom domain to the bucket for public URLs and set `R2_PUBLIC_URL`.

---

## Cloudflare Workers (function deployments)

- The platform deploys user functions to Workers via `internal/adapters/cloudflare/` (provider name `workers`).  
- Backends are configured with Cloudflare `account_id`, `api_token`, `script_name`; route binding uses `zone_id`.  
- Health checks, deploy, rollback, and blue/green (with DNS switch) are supported.  
- **workers.dev:** When deploying via the upload API, the Worker is not reachable at `<script>.<subdomain>.workers.dev` until the workers.dev subdomain is enabled. The client exposes `EnableWorkersDev(ctx, scriptName)`; call it after deploy if you need that URL (e.g. for health checks or E2E). Custom domains (zone routes) are unaffected.  
- See `docs/VERSIONING_SYSTEM.md` and `plans/VERSIONING_SYSTEM.md` for versioning behavior.

---

## Cloudflare Tunnel (optional)

To expose the orchestrator and dashboard without opening server ports:

1. Create a Tunnel in Cloudflare Zero Trust (https://one.dash.cloudflare.com → Networks → Tunnels → Create Tunnel → select "Cloudflared").  
2. Note the tunnel token.  
3. Run `deploy/cloudflare-tunnel.example.sh --token <YOUR_TOKEN>`.  
4. In the tunnel's Public Hostname configuration, add:
   - `api.localhost` → `http://localhost:8080`
   - `dashboard.localhost` → `http://localhost:3000`

Benefits: no inbound firewall rules, DDoS and WAF at the edge, automatic HTTPS.

---

## Cloudflare Pages (optional)

To host the dashboard (or docs) on Pages:

1. Connect the repo to Cloudflare Pages; set build command and output directory (e.g. `web/dashboard`: build `npm run build` or `bun run build`, output `dist`).
2. Set env vars in Pages (e.g. `VITE_API_URL=https://api.functionfly.com`).
3. In DNS, add CNAME `app` or `dashboard` → `<project>.pages.dev` with proxy on.
4. For staging, use a separate Pages project or branch and e.g. `cdn.staging.functionfly.com` / `app.staging.functionfly.com`.

`deploy/dns/cloudflare-geo-dns.json` already references `functionfly-dashboard.pages.dev` and `functionfly-docs.pages.dev`; point your Pages projects to those subdomains or update the JSON to match your project names.

### Cloudflare Pages Configuration

The dashboard is pre-configured for Cloudflare Pages deployment. Key files:

| File | Purpose |
|------|---------|
| `web/dashboard/_headers` | Security headers, CSP, and cache policies |
| `web/dashboard/_redirects` | SPA routing fallback (all routes → index.html) |
| `web/dashboard/vite.config.ts` | Build config with base path and Cloudflare plugin |

#### Environment Variables (Cloudflare Pages)

Set these in the Cloudflare Pages project settings:

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `VITE_API_URL` | Yes | Production API endpoint | `https://api.functionfly.com` |
| `VITE_BLOG_API_URL` | No | Blog API endpoint | `https://blog-api.functionfly.com` |
| `VITE_GOOGLE_ANALYTICS_ID` | No | GA4 Measurement ID | `G-XXXXXXXXXX` |
| `VITE_HOTJAR_SITE_ID` | No | Hotjar Site ID | `1234567` |

#### Build Settings (Cloudflare Pages)

| Setting | Value |
|---------|-------|
| Build command | `npm run build` or `bun run build` |
| Build output directory | `web/dashboard/dist` |
| Node.js version | `18` or `20` |

#### Deployment Steps

```bash
# 1. Build locally to verify
cd web/dashboard
npm install
npm run build

# 2. Verify _headers and _redirects are in dist/
ls -la dist/_headers dist/_redirects

# 3. Deploy via Cloudflare Pages
# Option A: Git integration (recommended)
# Connect repo in Cloudflare Dashboard → Pages → Create project
# Set build settings as above

# Option B: wrangler CLI
wrangler pages project create functionfly-dashboard
wrangler pages deploy dist --project-name=functionfly-dashboard
```

#### DNS Configuration

After deployment, add DNS records in Cloudflare:

| Type | Name | Value | Proxy |
|------|------|-------|-------|
| CNAME | app | `<project>.pages.dev` | Proxied |
| CNAME | dashboard | `<project>.pages.dev` | Proxied |

For production, use your custom domain (e.g., `app.functionfly.com`).

---

## WAF and security

- **Proxy:** Enable “Proxied” (orange cloud) for public hostnames to send traffic through Cloudflare (DDoS, caching, SSL).  
- **SSL/TLS:** Use “Full (strict)” and either Cloudflare-origin certs or your own on the origin.  
- **Security level:** In Security → Settings, set Security Level (e.g. Medium) and optionally configure WAF rules.  
- **Bot Fight Mode / Under Attack:** Enable when under abuse; adjust as needed to avoid blocking legitimate traffic.

---

## Quick checklist (production)

- [ ] Domain on Cloudflare; DNS managed (Zone ID and API token set).  
- [ ] `api.*` and `cdn.*` (and app/dashboard if used) have DNS records; proxy enabled where appropriate.  
- [ ] SSL/TLS Full (strict).  
- [ ] `CACHE_CDN_ENABLED=true`, `CACHE_CDN_PROVIDER=cloudflare`, `CACHE_CDN_BASE_URL` set.  
- [ ] R2 bucket and API token if using `STORAGE_BACKEND=r2`; `R2_ACCOUNT_ID` and credentials set.  
- [ ] Workers backend configured (account_id, api_token, script_name, zone_id) for function deployments.  
- [ ] Optional: Pages for dashboard/docs.  
- [ ] CI secrets: `CLOUDFLARE_ZONE_ID`, `CLOUDFLARE_API_TOKEN` (and `CLOUDFLARE_ACCOUNT_ID` if used in CI).

---

## E2E testing and production readiness

Use this checklist to verify the Cloudflare Workers integration end-to-end with real credentials. Run against a staging or test Cloudflare account first.

### Prerequisites

- Cloudflare account with Workers enabled.
- **API token** with permissions: Account / Workers Scripts (Edit), Account / Workers Routes (Edit), Zone / DNS (Edit). Create under My Profile → API Tokens.
- **Account ID**: from the Cloudflare dashboard URL or Overview.
- **Zone ID**: for the domain you will use (custom domain or test zone).
- **Workers subdomain**: set in Workers → Overview (e.g. `mycompany` → `mycompany.workers.dev`). Required for blue/green CNAME target.
- Orchestrator API running and authenticated (session cookie or token for `POST /api/...`).
- An **app** created in the dashboard (note `app_id`).

### 1. Create a Workers backend

**Request:** `POST /api/apps/{appId}/backends`

- `provider`: `"workers"`
- `region`: `"us-east-1"` (or any supported Workers region; see adapter `GetRegions()`)
- `url`: Worker URL after deploy, e.g. `https://<script_name>.<workers_subdomain>.workers.dev` (you can use the expected URL; it must match `*.workers.dev` or a custom domain)
- `shared_secret`: optional; if omitted the API generates one (needed for health check signing)

**Example body:**

```json
{
  "provider": "workers",
  "region": "us-east-1",
  "url": "https://e2e-test.mycompany.workers.dev"
}
```

**Verify:** Response returns backend `id` and `url`. List backends with `GET /api/apps/{appId}/backends`.

### 2. Deploy a minimal Worker script

**Request:** `POST /api/apps/{appId}/deploy`  
**Note:** This route may require HMAC signature (see `RequireHMACSignature` in `internal/api/routes.go`). Configure the CLI or use the dashboard deploy flow that signs requests.

- `provider`: `"workers"`
- `region`: `"us-east-1"`
- `artifact`: Base64-encoded Worker script (JavaScript).
- `provider_config`: `account_id`, `api_token`, `script_name`; optionally `zone_id` for route binding.

**Minimal script (save as `worker.js`, then base64-encode):**

```js
addEventListener('fetch', (e) => {
  e.respondWith(new Response('OK', { status: 200 }));
});
```

Add a `/healthz` responder for health checks:

```js
addEventListener('fetch', (e) => {
  const u = new URL(e.request.url);
  if (u.pathname === '/healthz') {
    return e.respondWith(new Response('ok', { status: 200 }));
  }
  e.respondWith(new Response('OK', { status: 200 }));
});
```

**Example body:**

```json
{
  "provider": "workers",
  "region": "us-east-1",
  "artifact": "<base64 of worker.js>",
  "routes": [],
  "provider_config": {
    "account_id": "<CLOUDFLARE_ACCOUNT_ID>",
    "api_token": "<API_TOKEN>",
    "script_name": "e2e-test"
  }
}
```

**Verify:** Response has `status: "success"` and a `deployment_id`. In Cloudflare dashboard, Workers → Scripts, the script `e2e-test` should exist.

### 3. Health check

- **Manual:** `GET https://<script_name>.<workers_subdomain>.workers.dev/healthz` → expect `200 OK`.
- **Via platform:** If the orchestrator runs health checks, ensure the backend URL is exactly the Worker URL (including `https://` and no trailing slash). Health checks use `/healthz` and request signing (see `internal/adapters/cloudflare/adapter.go`).

### 4. Blue/green deployment (optional)

**Request:** `POST /api/apps/{appId}/deploy/blue-green`  
Requires HMAC signature. Include in body:

- `artifact`: Base64-encoded new script.
- `provider`: `"workers"`
- `provider_config`: `account_id`, `api_token`, `script_name`, and **`workers_subdomain`** (e.g. `"mycompany"`).
- `zone_id`: Cloudflare zone ID for the domain.
- `domain`: Full record name to update (e.g. `fn.example.com`).
- `enable_proxied`: `true` or `false` for Cloudflare proxy.

**Verify:** DNS for `domain` points to the new Worker (e.g. `<script>-green.<workers_subdomain>.workers.dev`). Switch again to confirm blue/green toggle.

### 5. Rollback

**Request:** `POST /api/deployments/{deploymentId}/rollback`  
Body may include previous artifact or reference; see `HandleRollback` in `internal/api/handlers/deployments/`. Requires HMAC if so configured.

**Verify:** Worker script content reverts to the previous version in Cloudflare.

### 6. Cleanup

- Delete the Worker script in Cloudflare (Workers → Scripts → Delete), or leave it for future E2E runs.
- Optionally delete the backend via the API/dashboard.

### Automated E2E tests (Cloudflare API only)

The repo includes Go E2E tests that run against the real Cloudflare API when credentials are set. They are skipped when env vars are missing.

**Setup:** Copy `.env.cloudflare.e2e.example` to `.env.cloudflare.e2e`, set `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_API_TOKEN`. Optionally set `CLOUDFLARE_WORKERS_SUBDOMAIN` (e.g. `mycompany`) to test HTTP GET to the Worker URL and `/healthz`.

**Run:**

```bash
source .env.cloudflare.e2e && go test ./internal/adapters/cloudflare -run E2E -v -timeout 120s -count=1
```

Tests: deploy a minimal Worker, get deployment status, optionally hit the Worker URL and `/healthz`, set env vars, rollback, and delete the script (cleanup).

### Production readiness checklist (Workers)

- [ ] E2E: Run automated E2E (`source .env.cloudflare.e2e && go test ./internal/adapters/cloudflare -run E2E -v -count=1`) or manual: create backend, deploy script, confirm script appears in Cloudflare.
- [ ] E2E: Call Worker URL (and `/healthz` if implemented); expect 200.
- [ ] E2E: Blue/green with `workers_subdomain` set; confirm DNS switches to new script hostname.
- [ ] E2E: Rollback succeeds and Worker content reverts.
- [ ] API token has minimum required scopes; stored only in deploy request (not persisted in backend).
- [ ] Worker script name is unique per app/environment to avoid collisions.
- [ ] Health check and request signing work with the backend’s `shared_secret`.

---

## References

- **CDN:** `docs/CDN_SETUP.md`  
- **Object storage:** `docs/OBJECT_STORAGE.md`  
- **Production:** `docs/PRODUCTION_DEPLOYMENT.md`  
- **Staging:** `docs/STAGING_DEPLOYMENT_GUIDE.md`, `plans/STAGING_DEPLOYMENT_ARCHITECTURE.md`  
- **Deploy config:** `deploy/dns/cloudflare-geo-dns.json`, `scripts/update-cloudflare-dns.sh`, `.fly/setup-dns.sh`  
- **Workers adapter:** `internal/adapters/cloudflare/`

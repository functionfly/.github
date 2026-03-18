# CDN setup (cdn.functionfly.com)

This doc describes how to put **cdn.functionfly.com** in front of static assets (SDK, docs, static files) so links and responses use the CDN URL and benefit from edge caching.

**Cloudflare overview:** For DNS, R2, Workers, Tunnel, and Pages in one place, see [CLOUDFLARE.md](CLOUDFLARE.md).

---

## Best option for production (Fly.io or any host)

**Use Cloudflare R2 as the CDN origin.** No origin server is needed for static traffic: R2 serves SDK, docs, and static files at the edge. Your orchestrator (e.g. on Fly.io) stays free of CDN load, and you get unlimited egress from R2 through the Cloudflare cache.

| Approach | When to use |
|--------|----------------|
| **R2 (recommended)** | Production on Fly.io or whenever you want zero origin load, best scale, and no server to run. |
| Caddy | You already run a server (VPS) with Caddy in front of the API. |
| Fly + Cloudflare proxy | Quick path: point `cdn` at the same Fly app and put Cloudflare in front; origin stays Fly (see [Fly.io note](#when-using-flyio-no-caddy) below). |

Full manual steps for the **R2 production setup** are in [Production: Cloudflare R2 as CDN origin](#production-cloudflare-r2-as-cdn-origin) below.

---

## Production: Cloudflare R2 as CDN origin

Manual steps to serve `cdn.functionfly.com` from R2. No server required; works with Fly.io or any orchestrator host.

### 1. Create the R2 bucket

1. In [Cloudflare Dashboard](https://dash.cloudflare.com) go to **R2** → **Overview** → **Create bucket**.
2. Name it (e.g. `functionfly-cdn`), choose a region, create.
3. In the bucket, ensure **Settings** → **Public access** is configured for the custom domain (next step); R2 will show the CNAME target after you add the domain.

### 2. Attach custom domain cdn.functionfly.com

1. In the bucket, open **Settings** → **Custom Domains** → **Connect Domain**.
2. Enter `cdn.functionfly.com`. Cloudflare will show the target hostname (e.g. `functionfly-cdn.<account>.r2.cloudflarestorage.com`). Leave this open for the next step.

### 3. Add DNS record

1. Go to **Websites** → your zone **functionfly.com** → **DNS** → **Records**.
2. **Add record:**
   - **Type:** `CNAME`
   - **Name:** `cdn`
   - **Target:** the R2 custom domain target from step 2 (e.g. `functionfly-cdn.<account>.r2.cloudflarestorage.com`)
   - **Proxy status:** **Proxied** (orange cloud) so traffic and caching go through Cloudflare.
3. Save. Cloudflare will issue a certificate for `cdn.functionfly.com` automatically.

### 4. Upload assets to R2

Use the same path layout the app expects:

- `sdk/{sdk}/{version}/{filename}` — e.g. `sdk/js/1.0.0/functionfly.js`
- `docs/{type}/{version}/{path}` — e.g. `docs/api/1.0.0/readme.md`
- `static/{category}/{path}` — e.g. `static/images/logo.png`

**Options:**

- **Wrangler CLI:** From a directory that mirrors that layout, use `wrangler r2 object put functionfly-cdn/sdk/... --file=...` (or a script loop).
- **Dashboard:** R2 → bucket → **Upload** and create the folder structure.
- **CI:** After building SDK/docs, upload from GitHub Actions (or similar) using Wrangler or the S3-compatible API.

Ensure objects are **publicly readable** (bucket public access or per-object ACL as needed for your R2 setup).

### 5. Orchestrator (Fly.io) environment

Set these in Fly secrets (or your env) so the app uses the CDN URL for links and headers:

```bash
fly secrets set CACHE_CDN_ENABLED=true
fly secrets set CACHE_CDN_PROVIDER=cloudflare
fly secrets set CACHE_CDN_BASE_URL=https://cdn.functionfly.com
fly secrets set CACHE_CDN_MAX_AGE=86400
```

Redeploy or restart the app so the new env is picked up.

### 6. Verification

```bash
dig cdn.functionfly.com +short
curl -sI https://cdn.functionfly.com/sdk/js/1.0.0/functionfly.js
```

Expect **200** and caching headers. With `CACHE_CDN_ENABLED=true`, registry/cache stats and SDK/doc links in the app should point to `https://cdn.functionfly.com`.

---

## When using Fly.io (no Caddy)

If you don’t run a server (e.g. API only on Fly.io):

- **Recommended:** Use [Production: Cloudflare R2](#production-cloudflare-r2-as-cdn-origin) above. No origin server; R2 serves static assets at the edge.
- **Alternative:** Point `cdn.functionfly.com` at the **same Fly app** (add `cdn.functionfly.com` as a custom hostname in Fly, then CNAME `cdn` to your Fly app). Put Cloudflare in front (proxy on) so the edge caches responses. The API must answer `https://cdn.functionfly.com/sdk/...` (without `/v1`); the app currently serves these under `/v1/sdk/...`, so you’d need either a Cloudflare Worker that rewrites `/sdk` → `/v1/sdk` and proxies to the API, or root-level routes in the Go app for `/sdk`, `/docs`, `/static`. R2 avoids that and is the better production choice.

---

## Manual setup (Option A: Caddy as CDN origin)

Follow these steps in order to bring up the CDN with Caddy proxying to the orchestrator.

### 1. DNS record for cdn.functionfly.com

Add a record so `cdn.functionfly.com` points at the server that runs Caddy.

**If using Cloudflare DNS (recommended):**

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com) → select the zone **functionfly.com**.
2. Go to **DNS** → **Records** → **Add record**.
3. Set:
   - **Type:** `CNAME`
   - **Name:** `cdn` (creates `cdn.functionfly.com`)
   - **Target:** hostname of the server running Caddy, e.g.:
     - Same box as API: use the same target as `api.functionfly.com` (e.g. `api` or your edge hostname).
     - Or the server’s FQDN, e.g. `edge.functionfly.com` or your VPS hostname.
   - **Proxy status:** **Proxied** (orange cloud) for DDoS/caching, or **DNS only** if you want traffic to hit Caddy directly.
4. Save.

**If not using Cloudflare:** Create a CNAME (or A/AAAA) for `cdn` pointing to your Caddy server’s hostname or IP.

**Verify:**

```bash
dig cdn.functionfly.com +short
# or
nslookup cdn.functionfly.com
```

You should see the Caddy host’s IP (or a Cloudflare IP if proxied).

---

### 2. Ensure Caddy serves cdn.functionfly.com

The repo already defines the CDN block in `deploy/caddy/Caddyfile`. Confirm it’s present and that Caddy is loading that file.

1. **Check the block exists:**

```bash
grep -A 20 "cdn.functionfly.com" deploy/caddy/Caddyfile
```

You should see the `cdn.functionfly.com { ... }` block that rewrites `/sdk`, `/docs`, `/static` to `/v1/...` and reverse-proxies to `orchestrator-api:8080`.

2. **If Caddy runs on the host (not Docker):**
   - Copy or symlink the Caddyfile to Caddy’s config (e.g. `/etc/caddy/Caddyfile`).
   - Reload Caddy:

```bash
sudo systemctl reload caddy
# or
caddy reload --config /path/to/Caddyfile
```

3. **If Caddy runs in Docker (e.g. with docker-compose):**
   - Ensure `deploy/caddy/Caddyfile` is mounted into the container and the container uses it.
   - Reload:

```bash
docker exec <caddy-container-name> caddy reload --config /etc/caddy/Caddyfile
# or recreate the stack so the new Caddyfile is picked up
docker compose -f deploy/docker-compose.yml up -d --force-recreate caddy
```

4. **Important:** The CDN block uses `reverse_proxy orchestrator-api:8080`. So the Caddy container must resolve `orchestrator-api` (same Docker network as the orchestrator) or you must use the host/port where the orchestrator is reachable (e.g. `host.docker.internal:8080` or the actual hostname).

---

### 3. Orchestrator environment variables

Set these where the orchestrator runs (e.g. `.env`, systemd unit, or Docker env).

```bash
CACHE_CDN_ENABLED=true
CACHE_CDN_PROVIDER=cloudflare
CACHE_CDN_BASE_URL=https://cdn.functionfly.com
CACHE_CDN_MAX_AGE=86400
```

- **CACHE_CDN_ENABLED=true** — turns on CDN URLs and Cache-Control for SDK/docs/static.
- **CACHE_CDN_BASE_URL** — base URL used in links and headers (must match the host you configured in step 1 and 2).
- **CACHE_CDN_MAX_AGE=86400** — 24 hours; can be reduced for staging.

Restart the orchestrator after changing env (e.g. `systemctl restart orchestrator-api` or `docker compose restart orchestrator-api`).

---

### 4. Verification

- **DNS:** Already checked in step 1.
- **HTTP (CDN host):**

```bash
curl -sI https://cdn.functionfly.com/sdk/js/1.0.0/functionfly.js
```

Expect **200** (or **404** if that path isn’t deployed yet). Response should include caching headers; with Caddy you may see `Cache-Control` and optionally `X-CDN-*` style headers.

- **App:** With `CACHE_CDN_ENABLED=true`, registry/cache stats and SDK/doc links in the app should use `https://cdn.functionfly.com`.

---

## What the CDN serves

| Path pattern   | Purpose                    | Example |
|----------------|----------------------------|--------|
| `/sdk/*`       | SDK bundles by version     | `/sdk/js/1.0.0/functionfly.js` |
| `/docs/*`      | Documentation by type/ver | `/docs/api/1.0.0/readme.md` |
| `/static/*`    | Other static assets        | `/static/images/logo.png` |

The orchestrator API exposes these under `/v1/...` (see `internal/api/routes.go`). The CDN host (R2 or Caddy) serves the same path layout without the `/v1` prefix.

## Verification

- **DNS:** `dig cdn.functionfly.com` (or `nslookup cdn.functionfly.com`) should resolve to your Caddy host or R2.
- **HTTP:**  
  `curl -sI https://cdn.functionfly.com/sdk/js/1.0.0/functionfly.js`  
  should return 200 (or 404 if that path is not yet deployed) with `Cache-Control` and, if using Caddy/origin, `X-CDN-Provider` / `X-CDN-Cache` headers.
- **App:** With `CACHE_CDN_ENABLED=true`, registry/cache stats and SDK/doc links should reference `https://cdn.functionfly.com`.

## References

- Production env: `docs/PRODUCTION_DEPLOYMENT.md` (CDN Configuration section)
- Staging CDN: `CACHE_CDN_BASE_URL=https://cdn.staging.functionfly.com` in `.env.staging.example` / `docker-compose.staging.yml`
- Architecture: `plans/STAGING_DEPLOYMENT_ARCHITECTURE.md` (DNS table and R2 CNAME examples)

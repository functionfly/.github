# CDN setup (cdn.functionfly.com)

This doc describes how to put **cdn.functionfly.com** in front of static assets (SDK, docs, static files) so links and responses use the CDN URL and benefit from edge caching.

## What the CDN serves

| Path pattern   | Purpose                    | Example |
|----------------|----------------------------|--------|
| `/sdk/*`       | SDK bundles by version     | `/sdk/js/1.0.0/functionfly.js` |
| `/docs/*`      | Documentation by type/ver | `/docs/api/1.0.0/readme.md` |
| `/static/*`    | Other static assets        | `/static/images/logo.png` |

The orchestrator API exposes these under `/v1/...` (see `internal/api/routes.go`). The CDN host should proxy the same paths to the orchestrator (or to R2 once you migrate assets there).

## Option A: Caddy as CDN origin (recommended first step)

1. **DNS**  
   Add a CNAME (or A/AAAA if Caddy is on a fixed IP):
   - **Name:** `cdn`
   - **Target:** hostname of the server running Caddy (e.g. your main edge or the same host as `api.functionfly.com`).
   - **Proxy:** Cloudflare proxy (orange cloud) optional; turn on for DDoS and caching.

2. **Caddy**  
   The repo already includes a `cdn.functionfly.com` block in `deploy/caddy/Caddyfile` that:
   - Reverse-proxies to `orchestrator-api:8080`
   - Caches responses for `/sdk/*`, `/docs/*`, `/static/*` at the edge (e.g. 24h)

   Reload Caddy after changing the Caddyfile so the CDN host is active.

3. **Orchestrator env**  
   Point the app at the CDN URL (default is already `https://cdn.functionfly.com`):

   ```bash
   CACHE_CDN_ENABLED=true
   CACHE_CDN_PROVIDER=cloudflare
   CACHE_CDN_BASE_URL=https://cdn.functionfly.com
   CACHE_CDN_MAX_AGE=86400
   ```

   If you use another provider, set `CACHE_CDN_PROVIDER` to `cloudfront` or `fastly` and keep `CACHE_CDN_BASE_URL` as above.

After this, the API will use `https://cdn.functionfly.com` for CDN URLs and Cache-Control headers; clients and Cloudflare (if proxied) can cache at the edge.

## Option B: Cloudflare R2 as CDN origin

For higher scale or to offload static traffic from the orchestrator:

1. **R2 bucket**  
   Create a bucket (e.g. `functionfly-cdn`) in Cloudflare R2 and enable public access or a custom domain.

2. **Custom domain**  
   In R2 / Cloudflare Dashboard, attach the custom hostname `cdn.functionfly.com` to that bucket. Cloudflare will show the required CNAME target (e.g. `functionfly-cdn.r2.cloudflarestorage.com`).

3. **DNS**  
   - **Name:** `cdn`
   - **Target:** R2 custom domain target (e.g. `functionfly-cdn.r2.cloudflarestorage.com`)
   - **Proxy:** Typically enabled (orange cloud) so traffic goes through Cloudflare.

4. **Uploads**  
   Sync or upload SDK/docs/static assets to the bucket under the same path layout (`/sdk/...`, `/docs/...`, `/static/...`) so `https://cdn.functionfly.com/sdk/...` resolves to R2.

5. **Orchestrator**  
   Keep the same env as in Option A so the app still uses `CACHE_CDN_BASE_URL=https://cdn.functionfly.com` for links and headers.

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

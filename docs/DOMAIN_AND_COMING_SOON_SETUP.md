# Domain & Coming-Soon Setup (functionfly.com)

This guide covers:

1. **Fly.io backend** – deploy the API and attach `api.functionfly.com`
2. **Frontend (Vercel or Cloudflare Pages)** – deploy the dashboard/coming-soon page and attach `functionfly.com`
3. **Coming-soon mode** – so visitors to functionfly.com only see the launch/waitlist page

**If you don’t see the coming-soon page at functionfly.com:** The API at api.functionfly.com is separate from the **website** at functionfly.com. You must deploy the **frontend** (dashboard app) to Vercel or Cloudflare Pages and point functionfly.com and <www.functionfly.com> at that deployment. Until then, functionfly.com will show whatever is there now (e.g. registrar parking, or nothing). Follow the **Frontend** steps below.

**If you see example worker code** (e.g. "Hello from FunctionFly with Hot Reload!" or raw code from `examples/fixtures/index.js`): functionfly.com (or www) is currently pointed at a **Worker or wrong project** that serves that example, not the coming-soon site. To fix: point **functionfly.com** and **<www.functionfly.com>** at a **separate** deployment that builds and serves the **static coming-soon page** (build: `make build-coming-soon`, output: `web/dashboard/dist`). Use either a new Vercel/Cloudflare Pages project and add functionfly.com + www as its custom domains, or change the existing project so it builds/serves the coming-soon output instead of the Worker/example. Do not use a Worker or the `examples/` folder for the main site.

### Marketing site (`web/site`) vs coming-soon

The repo has **two** static options for the apex domain:

| Artifact | Source | When to use |
|----------|--------|-------------|
| **Full marketing site** | `web/site` (Astro). Build: `cd web/site && bun run build` → output `web/site/dist/` | Public launch: home, trust, blog, legal pages, etc. |
| **Waitlist / coming-soon** | `web/coming-soon` copied by `make build-coming-soon` → `web/dashboard/dist/` | Pre-launch or maintenance page with minimal surface |

Point **only one** deployment at `functionfly.com` / `www`. Do not flip DNS between them without aligning analytics, SEO, and forms. Set `PUBLIC_SITE_URL` (and optionally `PUBLIC_DOCS_URL`) when building `web/site` so canonical URLs, Open Graph, and `robots.txt` / sitemap match your domain.

---

## What you need to do

- **API (Fly.io)**  
  Deploy the app, set secrets (including `CORS_ALLOWED_ORIGINS` with marketing, app, auth, and admin origins — e.g. `https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://auth.functionfly.com,https://admin.functionfly.com` — plus `BASE_URL`, `FRONTEND_URL`).  
  Run: `flyctl certs add api.functionfly.com --app functionfly-control`.  
  In DNS: **CNAME** `api` → `functionfly-control.fly.dev` (and any TXT for Fly ownership if required).

- **Frontend (Vercel or Cloudflare Pages)**  
  Connect the repo.  
  **Build command:** `make build-coming-soon` (from repo root).  
  **Output directory:** `web/dashboard/dist`.  
  Optional: set **`API_URL`** to point the waitlist form at a different API (default `https://api.functionfly.com`).  
  Add domains `functionfly.com` and `www.functionfly.com` and set the DNS records they show (A/CNAME for root and www).

- **Verify**  
  Open <https://functionfly.com> and confirm you see the coming-soon page and that the waitlist form submits to <https://api.functionfly.com/v1/feedback>.

Full step-by-step (including optional Cloudflare Pages) is below, with a **Quick start** and then detailed sections.

---

## Quick start (get the coming-soon page live)

1. **API (Fly.io)**  
   Deploy backend, set secrets (DB, Redis, JWT, CORS, BASE_URL, FRONTEND_URL), then:

   ```bash
   flyctl certs add api.functionfly.com --app functionfly-control
   ```

   In DNS: **CNAME** `api` → `functionfly-control.fly.dev`. Verify: `curl -s https://api.functionfly.com/healthz`

2. **Frontend (Vercel or Cloudflare Pages)**  
   The coming-soon page is the **static** site in `web/coming-soon/index.html` (no React build).  
   - Connect this repo.  
   - **Build:** `make build-coming-soon` (from repo root). **Output:** `web/dashboard/dist`.  
   - Optional: set `API_URL` when building (e.g. `API_URL=https://api.staging.functionfly.com make build-coming-soon`) to point the waitlist form at a different API; default is `https://api.functionfly.com`.  
   - Add domains: `functionfly.com`, `www.functionfly.com`.  
   - In DNS: point root and `www` to the host’s targets (e.g. CNAME to Vercel/Pages).

3. **Verify**  
   Open <https://functionfly.com> and confirm you see the coming-soon page. Submit the waitlist form and confirm it posts to <https://api.functionfly.com/v1/feedback> (e.g. check Network tab or that you see the success state).

**Local build (optional):** From repo root, run `make build-coming-soon` to produce `web/dashboard/dist` from `web/coming-soon/` (e.g. for a manual deploy or to preview).

---

## 1. Fly.io backend (API)

### 1.1 Create the app and deploy

From the repo root:

```bash
flyctl auth login
flyctl apps create functionfly-control   # if not already created
flyctl deploy --config deploy/fly/functionfly-control/fly.toml --remote-only
```

See [FLY_DEPLOYMENT.md](FLY_DEPLOYMENT.md) for GitHub Actions, path filters, and manual deploy.

### 1.2 Set secrets

In **Fly Dashboard → functionfly-control → Secrets**, set at least:

- `DATABASE_URL` (or `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`)
- `REDIS_ADDR`, `REDIS_PASSWORD` (if Redis uses auth)
- `JWT_SECRET`, `API_SHARED_SECRET`, `DB_MASTER_KEY_PASSWORD`
- **Browser clients:** `CORS_ALLOWED_ORIGINS` must list every HTTPS origin that calls the API (e.g. marketing, `https://app.functionfly.com`, `https://auth.functionfly.com`, `https://admin.functionfly.com`). Example: `https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://auth.functionfly.com,https://admin.functionfly.com`. Also set `BASE_URL=https://api.functionfly.com`, `FRONTEND_URL=https://functionfly.com` — without CORS + `FRONTEND_URL`, the waitlist form on functionfly.com cannot POST to `/v1/feedback`.
- `LOG_LEVEL=info`, `DEVELOPMENT=false`

There is no separate “coming soon page secret”; the API uses CORS and the above URLs to allow the coming-soon site to submit to the feedback endpoint.

Full list and examples: [FLY_DEPLOYMENT.md#required-secrets](FLY_DEPLOYMENT.md#required-secrets-flyio). To set all from Neon + generated secrets (including coming-soon vars): `./.fly/set-secrets-from-neon.sh production`.

### 1.3 Custom domain for the API

**Option A — Let Fly issue the cert (default)**  
Add the hostname; Fly will use ACME (Let’s Encrypt) to issue a cert:

```bash
flyctl certs add api.functionfly.com --app functionfly-control
flyctl certs show api.functionfly.com --app functionfly-control
```

**Option B — Use your purchased wildcard SSL**  
If you have a wildcard certificate for `*.functionfly.com` (e.g. from your CA zip), you can import it so Fly uses that instead of ACME.

**One-shot (recommended):**

1. Extract your CA zip (e.g. the zip from your provider) into **`deploy/edge/certs-in/`** (on Windows: copy the extracted folder into `deploy/edge/certs-in/` in your repo). Add your **private key** there as `privkey.pem` (or any `.key` file) if the zip didn’t include it.
2. Ensure DNS is set (CNAME `api` → `functionfly-control.fly.dev`, and any TXT for Fly ownership).
3. From the repo root (with [flyctl](https://fly.io/docs/hands-on/install-flyctl/) logged in):

   ```bash
   deploy/edge/import-fly-cert.sh
   ```

   That script runs [prepare-certs.sh](deploy/edge/prepare-certs.sh) to build `fullchain.pem` and `privkey.pem`, then runs `flyctl certs import` for `api.functionfly.com` on app `functionfly-control`. Override with `FLY_APP=myapp FLY_HOSTNAME=api.mydomain.com deploy/edge/import-fly-cert.sh` if needed.

4. **Verify domain ownership (required after custom cert import)**  
   Fly will print a TXT record you must add in DNS, for example:
   - **Name/host:** `_fly-ownership.api` (or `_fly-ownership.api.functionfly.com` depending on your DNS provider)
   - **Type:** TXT  
   - **Value:** the token Fly showed (e.g. `app-dm1qpjj`)  
   Add that record at your DNS provider (e.g. Cloudflare), wait for propagation, then run:

   ```bash
   flyctl certs check api.functionfly.com --app functionfly-control
   ```

5. Confirm the cert is active and the API responds:

   ```bash
   flyctl certs show api.functionfly.com --app functionfly-control
   curl -sI https://api.functionfly.com/healthz
   ```

When both a custom cert and an ACME cert exist for a hostname, Fly serves the custom cert as primary. For renewals, re-import the new chain and key before the old cert expires.

**DNS (either option)**  
In your DNS (e.g. Cloudflare):

- **CNAME** `api` → `functionfly-control.fly.dev.`
- **Cloudflare:** set the `api` record to **DNS only** (grey cloud). If it’s Proxied (orange cloud), `api.functionfly.com` may not resolve or Fly’s TLS may not work; turning proxy off fixes this.
- **TXT** for custom cert: add the `_fly-ownership.api` (or full name) TXT record with the value from `flyctl certs show` / the import output so the custom certificate becomes active.

Then verify:

```bash
flyctl certs check api.functionfly.com --app functionfly-control
curl -s https://api.functionfly.com/healthz
```

**If curl fails with "Could not resolve host" or HTTP 000** — Check that the name resolves: `nslookup api.functionfly.com` or `host api.functionfly.com`. If it doesn’t, wait a few minutes for DNS propagation or flush DNS (e.g. on Windows: `ipconfig /flushdns`; WSL often uses Windows DNS). Ensure the CNAME target is exactly `functionfly-control.fly.dev` (trailing dot optional in Cloudflare).

**Edge VPS (e.g. edge.functionfly.com)**  
To use the same wildcard on your own edge VPS (Caddy in front of the edge service), see [deploy/edge/README-CERTS.md](deploy/edge/README-CERTS.md) for uploading and configuring the cert on the VPS nodes.

---

## 2. Frontend on your domain (functionfly.com)

You can host the frontend on **Vercel**, **Cloudflare Pages**, or any static host. The steps below use Vercel; adapt for others.

### 2.1 Connect the repo (Vercel)

1. [Vercel](https://vercel.com) → Add New → Project → Import your Git repo.
2. **Root Directory:** repo root.
3. **Build command:** `make build-coming-soon` (produces the static coming-soon page from `web/coming-soon/` into `web/dashboard/dist`).
4. **Output directory:** `web/dashboard/dist`.

Optional: to point the waitlist form at a different API (e.g. staging), set env **`API_URL`** (e.g. `https://api.staging.functionfly.com`) so the build injects it into the page.

### 2.2 Environment variables (Production)

The static coming-soon page does not require Vite env vars. Optional:

| Name | Value |
|------|--------|
| `API_URL` | (optional) API base for waitlist form; default `https://api.functionfly.com`. Set for staging or a different backend. |

When you’re ready to go live with the full site (dashboard, landing, etc.), switch the project to build the dashboard (e.g. `nx build dashboard`) and point the domain at that deployment instead.

### 2.3 Add domain functionfly.com

1. **Project → Settings → Domains**
2. Add `functionfly.com` and `www.functionfly.com`.
3. Vercel will show the required DNS records.

### 2.4 DNS (functionfly.com and www)

At your DNS provider (e.g. Cloudflare):

- **A** `@` → Vercel’s IP (or **CNAME** `@` → `cname.vercel-dns.com` if your provider allows CNAME on root).
- **CNAME** `www` → `cname.vercel-dns.com`.

(Use the exact targets Vercel shows.)

After propagation, `https://functionfly.com` and `https://www.functionfly.com` will serve the static coming-soon page; the waitlist form will POST to `https://api.functionfly.com/v1/feedback` (launch_waitlist).

---

## 3. Summary

| What | Where | URL |
|------|--------|-----|
| API (backend) | Fly.io | `https://api.functionfly.com` |
| Frontend (coming-soon only) | Vercel (or Pages) | `https://functionfly.com`, `https://www.functionfly.com` |

**After launch:** Switch the frontend project to build the full dashboard (e.g. `nx build dashboard`) and redeploy so the full site (landing, dashboard, etc.) is available.

---

## 4. Optional: Cloudflare Pages instead of Vercel

1. **Pages** → Create project → Connect Git → select repo.
2. **Build:** Framework preset = None. Build command: `make build-coming-soon`. Output: `web/dashboard/dist`.
3. **Environment variables (optional):** `API_URL=https://api.functionfly.com` (or your API base) to point the waitlist form; default is production API.
4. **Custom domains:** Add `functionfly.com` and `www.functionfly.com`; add the CNAME/A records Cloudflare shows.

---

## 5. Deploy frontend via CLI (Cloudflare Pages or Vercel)

Use the CLI to build and deploy the coming-soon page without the dashboard UI.

### Cloudflare Pages (wrangler)

From the repo root. **Requires** **`CLOUDFLARE_API_TOKEN`** and **`CLOUDFLARE_ACCOUNT_ID`** (account ID from dashboard URL or Overview; without it you may get "Unable to authenticate request [code: 10001]"). Create the token at [Cloudflare API Tokens](https://dash.cloudflare.com/profile/api-tokens). **Requires `CLOUDFLARE_API_TOKEN`** (create at [Cloudflare API Tokens](https://dash.cloudflare.com/profile/api-tokens) with “Edit Cloudflare Workers” / Pages permissions).

```bash
# One-shot: build and deploy (creates project if needed)
make deploy-coming-soon
```

Or step by step:

```bash
# 1. Build the dashboard with coming-soon mode
make build-coming-soon

# 2. Set auth (or add to .env)
export CLOUDFLARE_API_TOKEN=your_token_here
export CLOUDFLARE_ACCOUNT_ID=your_account_id

# 3. Create the Pages project once (if it doesn't exist)
npx wrangler pages project create functionfly-dashboard

# 4. Deploy the built output
npx wrangler pages deploy web/dashboard/dist --project-name=functionfly-dashboard
```

Then **point functionfly.com at Pages** (otherwise you’ll still see the Worker example):

1. **Add custom domains to Pages**  
   **Cloudflare Dashboard** → **Workers & Pages** → **functionfly-dashboard** → **Custom domains** → **Set up a custom domain**. Add `functionfly.com` and `www.functionfly.com`. Cloudflare will show the DNS records to create (or offer to add them automatically if the zone is on Cloudflare).

2. **Stop the Worker from handling functionfly.com**  
   **Workers & Pages** → **Workers** → select the Worker that currently serves the “Hello from FunctionFly with Hot Reload!” example → **Settings** → **Triggers** → **Routes**. Remove (or change) the routes for `functionfly.com/*` and `www.functionfly.com/*` so that traffic for those hostnames goes to **Pages** (functionfly-dashboard), not the Worker.  
   If you use **Workers Routes** in the **functionfly.com** zone, remove the route that sends `*functionfly.com/*` to that Worker.

3. **DNS**  
   In the **functionfly.com** zone, ensure `@` (root) and `www` point to the Pages project (Cloudflare usually gives a CNAME target like `functionfly-dashboard.pages.dev`; for root some providers use a CNAME flattening or A record). After propagation, `https://functionfly.com` and `https://www.functionfly.com` will serve the coming-soon page.

**SSL on \*.pages.dev:** If `https://<deployment>.functionfly-dashboard.pages.dev` shows ERR_SSL_VERSION_OR_CIPHER_MISMATCH, use the custom domain (`https://functionfly.com`) instead; Cloudflare will terminate TLS for the custom domain. The \*.pages.dev URL can be ignored for production.

### Vercel

From the repo root:

```bash
# 1. Build with coming-soon mode
make build-coming-soon

# 2. Deploy (requires Vercel CLI: npm i -g vercel)
vercel --prod
# Or link first: vercel link, then set env in dashboard (optional API_URL) and run vercel --prod
```

Add **functionfly.com** and **<www.functionfly.com>** in **Vercel** → Project → Settings → Domains, then update DNS as Vercel instructs.

---

## 6. Troubleshooting

### "Blocked by CORS policy: No 'Access-Control-Allow-Origin' header"

The browser blocks requests from `https://www.functionfly.com` (or `https://functionfly.com`) to `https://api.functionfly.com` because the API did not send a CORS allow-origin header.

**Fix:** On the **API** (Fly.io app), set the allowed frontend origins. From repo root:

```bash
flyctl secrets set CORS_ALLOWED_ORIGINS="https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://auth.functionfly.com,https://admin.functionfly.com" --app functionfly-control
```

Redeploy or restart the API if needed. Ensure there are no spaces inside the comma-separated list. After this, the preflight `OPTIONS` and the actual `GET`/`POST` (e.g. `/api/auth/get-session`, `/v1/feedback`) will include `Access-Control-Allow-Origin` and the browser will allow the request.

### "Refused to execute script from .../_vercel/insights/script.js" (MIME type 'text/html')

This usually happens when the **frontend is deployed to Cloudflare Pages** (or another host) but the app still loads Vercel Web Analytics. The script URL `/_vercel/insights/script.js` only exists on Vercel; elsewhere the server returns the SPA fallback (index.html), so the browser reports a wrong MIME type.

**Fix:** If you deploy to Cloudflare Pages, do **not** enable Vercel Web Analytics for that deployment. The dashboard only injects the Vercel Analytics script when `VITE_VERCEL_ANALYTICS=true` at build time; leave it unset for Cloudflare builds. If you use Vercel for the same app, set `VITE_VERCEL_ANALYTICS=true` in the Vercel project environment so the script is loaded only there.

---

## 7. Testing the waitlist (email) API

The coming-soon form submits to `POST /v1/feedback` with `feedbackType=launch_waitlist`. The API stores the email and interests in the database; it does not send a confirmation email to the user.

### "Failed to create feedback" — allow `launch_waitlist` in the database

If the form returns **"Failed to create feedback"** (or the API returns 400 with a message about launch waitlist), the API’s database likely does not yet allow `launch_waitlist` as a `feedback_type`. Run this migration on the **same database the API uses** (production, staging, or local):

```sql
-- migrations/20260328000000_feedback_launch_waitlist.up.sql
ALTER TABLE feedback DROP CONSTRAINT IF EXISTS feedback_feedback_type_check;
ALTER TABLE feedback ADD CONSTRAINT feedback_feedback_type_check CHECK (
  feedback_type IN ('bug', 'feature', 'improvement', 'general', 'launch_waitlist')
);
```

- **Local:** `psql $DATABASE_URL -f migrations/20260328000000_feedback_launch_waitlist.up.sql` (or run the SQL in your DB client).
- **Fly.io / Neon:** Connect to the DB (e.g. `flyctl postgres connect` or Neon SQL Editor) and run the two `ALTER TABLE` statements above.

### Using Neon (or a specific DB) for the local API

To have your **local** API (e.g. `./bin/orchestrator-api --skip-migrations`) use **Neon** instead of local Postgres:

1. Set `DATABASE_URL` to your Neon pooled connection string (Neon Console → Connection details → **Connection string** with **Pooled**), e.g.:

   ```bash
   export DATABASE_URL="postgresql://USER:PASSWORD@ep-xxx-pooler.region.aws.neon.tech/functionfly?sslmode=require"
   ```

   Or source a `.env` that contains `DATABASE_URL` (see [docs/NEON.md](NEON.md)).

2. Apply the schema (the app uses `--skip-migrations` by default). Ensure the `feedback` table exists and allows `launch_waitlist` (run the migration that adds it, or apply `migrations/20260328000000_feedback_launch_waitlist.up.sql` manually if you manage schema yourself).

3. Start the API (e.g. `make dev` or `./bin/orchestrator-api --skip-migrations`). It will connect to the DB given by `DATABASE_URL` (or `DB_*` if `DATABASE_URL` is not set).

### Option A: curl (local or production API)

**Local API (default port 8080):**

```bash
curl -X POST http://localhost:8080/v1/feedback \
  -F "feedbackType=launch_waitlist" \
  -F "subject=Launch signup" \
  -F "message=Serverless functions, Early access / beta" \
  -F "email=you@example.com"
```

**Production API:**

```bash
curl -X POST https://api.functionfly.com/v1/feedback \
  -F "feedbackType=launch_waitlist" \
  -F "subject=Launch signup" \
  -F "message=Serverless functions, Early access / beta" \
  -F "email=you@example.com"
```

Success: HTTP 200 and a JSON body with the created feedback (e.g. `id`, `feedback_type`, `user_email`, `message`, `created_at`). Failure: 4xx/5xx and often `{"error":"..."}`.

### Option B: Browser (form + Network tab)

1. Open the coming-soon page (e.g. `https://www.functionfly.com` or `http://localhost:3000` with the dashboard running).
2. Open DevTools → **Network**; filter by **Fetch/XHR**.
3. Enter an email and optional interests, then click **Notify me**.
4. Find the request to `feedback` (or `v1/feedback`). Check:
   - **Status** 200 → API accepted and stored the signup.
   - **Response** tab → JSON with `id` and `user_email` (or `user_email: null` if the backend didn’t receive email).

If the request is **blocked (CORS)** or fails, fix CORS on the API (see §6) and ensure the frontend’s `VITE_API_URL` points at the same API you’re testing.

### Option C: Confirm in the database or admin

- **Database:** Query the `feedback` table for `feedback_type = 'launch_waitlist'` and recent `created_at`; the `user_email` column holds the submitted email.
- **Admin UI:** If you have an admin feedback list (e.g. `/admin` → feedback), open it and confirm the new waitlist entry and email.

### View the list of waitlist signups

**Admin API (recommended)**  
As an admin, call the feedback list endpoint with `feedback_type=launch_waitlist`:

```bash
# Replace YOUR_JWT with a valid admin JWT (e.g. from logging in as admin@functionfly.local)
curl -s -H "Authorization: Bearer YOUR_JWT" \
  "https://api.functionfly.com/v1/admin/feedback?feedback_type=launch_waitlist&limit=100"
```

The response is JSON with a `feedback` array; each item has `user_email`, `message` (interests), `created_at`, etc.

**Database**  
To list signups directly in SQL (e.g. in Neon SQL Editor or `psql $DATABASE_URL`):

```sql
SELECT id, user_email, message, created_at
FROM feedback
WHERE feedback_type = 'launch_waitlist'
ORDER BY created_at DESC;
```

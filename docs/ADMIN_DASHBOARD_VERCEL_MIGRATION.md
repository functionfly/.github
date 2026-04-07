# Admin Dashboard: Cloudflare Pages → Vercel Migration

This guide covers migrating `admin.functionfly.com` from Cloudflare Pages to Vercel.

## Overview

| Item | Before (Cloudflare Pages) | After (Vercel) |
|------|-------------------------|-----------------|
| **Hosting** | Cloudflare Pages (`functionfly-admin-dashboard`) | Vercel (auto-detected) |
| **Config file** | `public/_redirects`, `public/_headers` | `vercel.json` |
| **Build command** | `bun run build:vite` | `bun run build:vite` |
| **Output** | `dist/` | `dist/` |
| **SPA routing** | `_redirects: /* → /index.html 200` | `vercel.json` rewrites |

## One-Time Setup

### 1. Install Vercel CLI

```bash
npm i -g vercel
```

### 2. Login and Link Project

```bash
cd web/admin-dashboard
vercel login
vercel link
```

### 3. Pull Environment Variables

```bash
vercel env pull
```

This creates a `.env.vercel` file with values from the Vercel dashboard.

### 4. Add Custom Domain

In the Vercel dashboard:

1. Go to **Settings → Domains**
2. Click **Add** → enter `admin.functionfly.com`
3. Follow the instructions to verify domain ownership

### 5. Update DNS

In Cloudflare DNS for `functionfly.com`, update the existing `admin` CNAME record:

| Type | Name | Value | Proxy |
|------|------|-------|-------|
| CNAME | admin | `cname.vercel-dns.com` | Proxied |

**Current value:** `functionfly-admin-dashboard.pages.dev` (Cloudflare Pages)
**New value:** `cname.vercel-dns.com` (Vercel)

This is the same pattern already used for `app` and `docs` subdomains pointing to Vercel.

### 6. Update CORS

Add `https://admin.functionfly.com` to `CORS_ALLOWED_ORIGINS` in your orchestrator API:

```bash
# Example: update in Fly.io secrets
fly secrets set CORS_ALLOWED_ORIGINS="https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://auth.functionfly.com,https://admin.functionfly.com,https://dashboard.functionfly.com" -a functionfly-control
```

### 7. Add Environment Variables in Vercel

Set these in **Vercel Dashboard → Settings → Environment Variables** or via CLI:

| Name | Example Value | Notes |
|------|--------------|-------|
| `VITE_API_BASE_URL` | `https://api.functionfly.com` | Production API |
| `VITE_ADMIN_API_BASE_URL` | `https://api.functionfly.com/v1/admin` | Admin API |
| `VITE_ADMIN_SHARED_SECRET` | `(secret)` | HMAC signing secret |
| `VITE_SESSION_TIMEOUT` | `1800000` | 30 minutes |
| `VITE_IDLE_TIMEOUT` | `900000` | 15 minutes |
| `VITE_MFA_REVERIFY_INTERVAL` | `14400000` | 4 hours |
| `VITE_ENABLE_IP_WHITELIST` | `true` | Security |
| `VITE_ENABLE_DEVICE_FINGERPRINT` | `true` | Security |
| `VITE_ENABLE_AUDIT_LOGGING` | `true` | Audit |
| `VITE_SENTRY_DSN` | `(optional)` | Error tracking |
| `VITE_EXPECT_ZT_HEADERS` | `true` | Zero Trust |

## Deploy

### Local Deploy

```bash
cd web/admin-dashboard

# Production
bun run vercel:deploy

# Preview (for testing)
bun run vercel:deploy:preview
```

### Using Nx

```bash
# Production
nx run admin-dashboard:deploy

# Preview
nx run admin-dashboard:deploy:preview
```

### Via Git (Automatic)

Once linked, push to the configured branch and Vercel auto-deploys.

## Rollback (if needed)

```bash
# Find previous deployment
vercel ls

# Promote previous deployment to production
vercel alias set <deployment-url> admin.functionfly.com
```

## Cleanup (Cloudflare)

After verifying Vercel works:

1. **Remove Cloudflare Pages project** (optional): Dashboard → Workers & Pages → Delete
2. **Keep DNS CNAME** pointing to Vercel, or switch to Vercel DNS
3. **Remove** `bun run pages:deploy` and `bun run pages:deploy:preview` from `package.json`

## Files Changed

| File | Change |
|------|--------|
| `web/admin-dashboard/vercel.json` | Created - SPA routing + security headers |
| `web/admin-dashboard/package.json` | Added `vercel:deploy` scripts |
| `web/admin-dashboard/project.json` | Added `deploy` and `deploy:preview` targets |
| `web/admin-dashboard/README.md` | Updated deployment instructions |

## Troubleshooting

### 404 on refresh (SPA routes)

Verify `vercel.json` has the rewrite:

```json
{
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ]
}
```

### CORS errors

1. Check `CORS_ALLOWED_ORIGINS` includes `https://admin.functionfly.com`
2. Verify the API is restarted to pick up new CORS settings
3. Check browser console for specific blocked origin

### Environment variables not loading

1. Run `vercel env pull` to fetch latest from dashboard
2. Check variable names match exactly (case-sensitive)
3. For preview deploys, set variables for "Production" and "Preview" environments

### Domain not working

1. Verify DNS is pointing to Vercel: `dig +short admin.functionfly.com`
2. Check Vercel domain status in dashboard
3. Wait for SSL certificate to provision (can take a few minutes)

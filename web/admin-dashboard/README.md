# FunctionFly Admin Dashboard

Standalone admin dashboard SPA for FunctionFly serverless platform with enhanced security.

## Features

- ✅ Standalone SPA at `admin.functionfly.com`
- ✅ Enhanced session management (30 min timeout, 15 min idle)
- ✅ MFA re-verification every 4 hours
- ✅ IP whitelisting support
- ✅ Comprehensive audit logging
- ✅ HMAC request signing for sensitive operations
- ✅ Rate limiting and security headers

## Getting Started

### Prerequisites

- Node.js 20+
- npm or yarn

### Installation

```bash
cd web/admin-dashboard
npm install
```

### Development

```bash
npm run dev
```

Server runs on `http://localhost:3002`

### Build

```bash
npm run build
npm run preview
```

## Environment Configuration

Copy `.env.example` to `.env.production` and configure:

```bash
cp .env.example .env.production
```

Key variables:

- `VITE_API_BASE_URL` - API base URL
- `VITE_ADMIN_SHARED_SECRET` - HMAC signing secret
- `VITE_SESSION_TIMEOUT` - Session timeout in ms
- `VITE_IDLE_TIMEOUT` - Idle timeout in ms
- `VITE_MFA_REVERIFY_INTERVAL` - MFA re-verification interval

## Testing

```bash
# Unit tests
npm test

# E2E tests
npm run e2e

# Lint
npm run lint
```

## Build for Production

```bash
npm run build
```

Output is in `dist/` directory.

### Cloudflare Pages (Wrangler CLI)

Static assets and SPA redirects are configured via `public/_redirects` and `public/_headers`. The project name defaults to `functionfly-admin-dashboard` (override with `--project-name`).

**Production hostname: `admin.functionfly.com`**

Wrangler’s success URL (`https://<hash>.functionfly-admin-dashboard.pages.dev`) is only the default Pages preview host. To serve the admin app at **`https://admin.functionfly.com`**:

1. **DNS** — In the `functionfly.com` zone, **CNAME** `admin` → `functionfly-admin-dashboard.pages.dev` (proxied / orange cloud). The repo template for this is `deploy/dns/cloudflare-geo-dns.json` (record name `admin`).
2. **Custom domain on the Pages project** — Cloudflare Dashboard → **Workers & Pages** → **functionfly-admin-dashboard** → **Custom domains** → **Set up a domain** → `admin.functionfly.com`. Finish the flow so Pages issues TLS and routes that hostname to this project (same account as the zone).
3. **API** — Add `https://admin.functionfly.com` to `CORS_ALLOWED_ORIGINS` on the orchestrator so the browser can call `https://api.functionfly.com`.

**One-time**

1. Create an API token with **Account: Cloudflare Pages: Edit** (and **Account: Read** if prompted).
2. Export it: `export CLOUDFLARE_API_TOKEN=...` (required in CI/non-interactive shells), or run `wrangler login` locally.
3. Create the Pages project once (optional; Wrangler may prompt on first deploy):

   ```bash
   bunx wrangler pages project create functionfly-admin-dashboard --production-branch=master
   ```

**Deploy**

```bash
cd web/admin-dashboard
bun run pages:deploy
```

This runs `vite build` then `wrangler pages deploy dist`. Set `VITE_*` variables in the Cloudflare Pages project for production builds, or use `.env.production` when building locally.

**Note:** `npm run build` runs `tsc && vite build`. Until `tsc` is clean, the Pages script uses `build:vite` only. Fix TypeScript errors to restore strict `tsc` in CI.

## Docker Deployment

```bash
docker build -t functionfly-admin-dashboard .
docker run -p 80:80 functionfly-admin-dashboard
```

## Project Structure

```
src/
├── pages/          # Admin pages
├── components/     # Reusable components
├── stores/         # Zustand state management
├── hooks/          # Custom React hooks
├── lib/
│   ├── api/        # API client & HMAC signing
│   ├── security/   # Security utilities
│   └── monitoring/ # Monitoring & analytics
├── types/          # TypeScript types
└── App.tsx         # Main app component
```

## Security

- All admin operations require JWT authentication
- Sensitive operations signed with HMAC-SHA256
- Session stored in memory only (no localStorage)
- Automatic session expiry and idle timeout
- MFA re-verification on interval
- Comprehensive audit logging
- IP whitelist support (optional)

## Support

For issues or questions, contact the FunctionFly team.

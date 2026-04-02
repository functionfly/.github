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

### Vercel Deployment

The admin dashboard is deployed to Vercel with SPA routing and security headers configured in `vercel.json`.

**Production hostname: `admin.functionfly.com`**

Wrangler’s success URL (`https://<hash>.functionfly-admin-dashboard.pages.dev`) is only the default Pages preview host. To serve the admin app at **`https://admin.functionfly.com`**:

1. Install Vercel CLI: `npm i -g vercel`
2. Login to Vercel: `vercel login`
3. Link the project (from `web/admin-dashboard` dir): `vercel link`
4. Pull environment variables: `vercel env pull`
5. Add custom domain in Vercel Dashboard: **Settings → Domains → Add** → `admin.functionfly.com`
6. Configure DNS in Cloudflare: **CNAME** `admin` → `cname.vercel-dns.com` (or use Vercel DNS if migrating fully)
7. Add `https://admin.functionfly.com` to `CORS_ALLOWED_ORIGINS` on the orchestrator API

**Deploy:**

```bash
cd web/admin-dashboard

# Production deploy
bun run vercel:deploy

# Or using Nx
nx run admin-dashboard:deploy

# Preview deploy
bun run vercel:deploy:preview
```

**Environment variables** (set in Vercel Dashboard or via `vercel env add`):

- `VITE_API_BASE_URL` - API base URL (e.g., `https://api.functionfly.com`)
- `VITE_ADMIN_API_BASE_URL` - Admin API URL (e.g., `https://api.functionfly.com/v1/admin`)
- `VITE_ADMIN_SHARED_SECRET` - HMAC signing secret
- `VITE_SESSION_TIMEOUT` - Session timeout in ms
- `VITE_IDLE_TIMEOUT` - Idle timeout in ms
- `VITE_MFA_REVERIFY_INTERVAL` - MFA re-verification interval
- `VITE_ENABLE_IP_WHITELIST` - Enable IP whitelisting
- `VITE_ENABLE_DEVICE_FINGERPRINT` - Enable device fingerprinting
- `VITE_ENABLE_AUDIT_LOGGING` - Enable audit logging
- `VITE_SENTRY_DSN` - Sentry DSN (optional)
- `VITE_EXPECT_ZT_HEADERS` - Expect Zero Trust headers

**Note:** `npm run build` runs `tsc && vite build`. Until `tsc` is clean, the deploy script uses `build:vite` only. Fix TypeScript errors to restore strict `tsc` in CI.

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

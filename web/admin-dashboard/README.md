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

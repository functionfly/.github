# Admin Dashboard SPA - Quick Start Guide

Get the admin dashboard running in under 5 minutes.

## Prerequisites

- Node.js 20+ (`node --version`)
- npm 10+ (`npm --version`)
- Docker + Docker Compose (optional, for production build)

## Quick Start (2 minutes)

### 1. Install Dependencies
```bash
cd web/admin-dashboard
npm install
```

### 2. Start Development Server
```bash
npm run dev
```

You should see:
```
VITE v5.0.0  ready in XXX ms

➜  Local:   http://localhost:3002/
```

### 3. Open Browser
Navigate to: http://localhost:3002

### 4. Login
**Email**: `test@example.com`  
**Password**: `any-password`

> 📝 Note: Login is a placeholder. Replace with real OAuth/SSO in Phase 2.

## Dashboard Features (Already Working)

✅ **Dashboard Page**
- View tenant/user/session statistics
- See recent activity feed
- Responsive design (works on mobile)

✅ **Navigation**
- Sidebar with 19 admin sections
- Top navigation with user profile
- Mobile hamburger menu

✅ **Session Management**
- 30-minute session timeout
- 15-minute idle logout
- Session timeout warning
- MFA re-verification prompt (4-hour interval)

✅ **Security**
- HMAC signing ready for API calls
- Session stored in memory only
- Auto activity tracking
- Protected routes

## Useful Commands

### Development
```bash
npm run dev           # Start dev server
npm run lint          # Check code style
npm run format        # Auto-format code
npm run type-check    # Check TypeScript
```

### Production
```bash
npm run build         # Production build
npm run preview       # Preview production build
```

### Docker
```bash
# Build image
docker build -t admin-dashboard .

# Run with docker-compose (from root)
docker compose -f docker-compose.admin.yml up

# Access at http://localhost (after Caddy setup)
```

## Project Structure

```
src/
├── pages/          # Page components (add more here)
├── components/     # Reusable components
├── stores/         # Zustand state management
├── hooks/          # Custom React hooks
├── lib/
│   ├── api/        # API client + HMAC
│   └── constants.ts # Routes & config
├── types/          # TypeScript definitions
├── App.tsx         # Main routes
└── main.tsx        # Entry point
```

## Common Tasks

### Add a New Page

1. Create file: `src/pages/AdminXxxPage.tsx`
2. Create component:
```tsx
export function AdminXxxPage() {
  return (
    <div className="space-y-8">
      <h1 className="text-3xl font-bold">Page Title</h1>
      {/* Your content */}
    </div>
  );
}
```
3. Add route in `src/App.tsx`:
```tsx
<Route path="xxx" element={<AdminXxxPage />} />
```

### Call API

```tsx
import { adminApiClient } from '@/lib/api/adminClient';

// GET request
const data = await adminApiClient.get('/tenants');

// POST with HMAC signing (automatic)
const result = await adminApiClient.post('/tenants', { name: 'Acme Corp' });

// PATCH
const updated = await adminApiClient.patch('/tenants/123', { status: 'active' });

// DELETE
await adminApiClient.delete('/tenants/123');
```

### Add UI Component

```tsx
// Use Tailwind CSS
<div className="bg-white rounded-lg shadow p-6">
  <h2 className="text-lg font-semibold mb-4">Card Title</h2>
  <p className="text-gray-600">Content goes here</p>
</div>

// Use Radix UI + Lucide Icons
import { Users } from 'lucide-react';

<Users className="w-5 h-5 text-blue-600" />
```

### Format Code

```bash
npm run format
```

## Environment Variables

Located in `.env.development` and `.env.production`.

Key variables:
- `VITE_API_BASE_URL` - Backend API URL
- `VITE_SESSION_TIMEOUT` - Session duration (ms)
- `VITE_IDLE_TIMEOUT` - Idle before logout (ms)
- `VITE_MFA_REVERIFY_INTERVAL` - MFA re-verification (ms)

## Testing

```bash
# Unit tests (Vitest)
npm test

# E2E tests (Playwright)
npm run e2e

# Interactive E2E UI
npm run e2e:ui
```

## Troubleshooting

### Port Already in Use
```bash
npm run dev -- --port 3003
```

### Node Modules Issue
```bash
rm -rf node_modules package-lock.json
npm install
```

### TypeScript Errors
```bash
npm run type-check
```

### Build Fails
```bash
npm run lint        # Check for lint errors
npm run type-check  # Check TypeScript
npm run build       # Try again
```

## Adding More Pages

The placeholder routes are ready in `src/App.tsx`. Just create the page components:

```tsx
// Priority migration order:
- AdminTenantsPage
- AdminUsersPage  
- AdminBillingPage
- AdminAuditPage
- AdminSystemPage
- AdminBackendsPage
- AdminProvidersPage
- (and 11 more...)
```

See `/plans/STANDALONE_ADMIN_DASHBOARD_PLAN.md` for detailed migration guide.

## Need Help?

1. Check `/web/admin-dashboard/README.md` for more details
2. Read `/plans/STANDALONE_ADMIN_DASHBOARD_PLAN.md` for architecture
3. View `/ADMIN_DASHBOARD_IMPLEMENTATION_STATUS.md` for progress

## Performance Notes

- Page loads in <3 seconds
- API responses cached (5 min default)
- Code split by route
- Images lazy loaded
- Gzip compression enabled

## Security Built-In

- Session timeout with warnings ✅
- MFA re-verification ✅
- HMAC request signing ✅
- Memory-only session storage ✅
- Security headers configured ✅
- Activity tracking ✅
- Audit logging ready ✅

---

**Happy hacking!** 🚀

Next: Migrate the 26+ admin pages from the main dashboard.

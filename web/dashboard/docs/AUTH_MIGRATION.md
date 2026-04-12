# Dashboard Auth Migration to @web/auth

This document describes the migration of the dashboard authentication from an embedded auth system to a standalone auth microsite (`@web/auth` / `functionfly-auth`).

## Overview

The dashboard now delegates all authentication (login, signup, password reset, MFA) to the standalone auth site running at `auth.functionfly.com` (or `localhost:4324` in dev). This provides:

- Centralized auth logic across all FunctionFly properties
- Consistent branding and UX for auth flows
- Better security isolation
- Easier maintenance of auth code

## Architecture

```
┌─────────────────────────┐     ┌─────────────────────────┐     ┌─────────────────────────┐
│       Dashboard         │     │      Auth Site          │     │    Orchestrator API     │
│   (app.functionfly.com) │     │ (auth.functionfly.com)  │     │  (api.functionfly.com)  │
│                         │     │                         │     │                         │
│ ┌─────────────────────┐ │     │ ┌─────────────────────┐ │     │ ┌─────────────────────┐ │
│ │   ProtectedRoute    │ │     │ │   Login/Signup      │ │     │ │   /v1/auth/*        │ │
│ │                     │ │     │ │   Pages             │ │     │ │                     │ │
│ └─────────────────────┘ │     │ └─────────────────────┘ │     │ └─────────────────────┘ │
│           │             │     │           │             │     │           │             │
│           │ redirect    │     │           │ API calls   │     │           │             │
│           │─────────────┼─────┼──────────▶│             │     │           │             │
│           │             │     │           │             │     │           │             │
│           │             │     │           └─────────────┼─────┼──────────▶│             │
│           │             │     │                       │   │     │           │             │
│           │             │     │           │             │   │     │           │             │
│           │             │     │           │◀────────────┘   │     │           │             │
│           │             │     │           │   returns tokens │     │           │             │
│           │             │     │           │                  │     │           │             │
│           │             │     │           ▼                  │     │           │             │
│           │             │     │ ┌─────────────────────┐        │     │           │             │
│           │             │     │ │  sessionStorage:    │        │     │           │             │
│           │             │     │ │  ff_token           │        │     │           │             │
│           │             │     │ │  ff_refresh_token   │        │     │           │             │
│           │             │     │ └─────────────────────┘        │     │           │             │
│           │             │     │           │                    │     │           │             │
│           │             │     │           │ redirect to        │     │           │             │
│           │             │     │           │ dashboard          │     │           │             │
│           │             │     │           ▼                    │     │           │             │
│ ┌─────────────────────┐ │     │                                 │     │           │             │
│ │   /auth/callback    │ │◀────┘                                 │     │           │             │
│ │   (migrates tokens  │ │         ┌─────────────────────┐      │     │           │             │
│ │   to localStorage)  │ │◀────────│  localStorage:        │      │     │           │             │
│ └─────────────────────┘ │         │  ff-access-token      │      │     │           │             │
│           │             │         │  ff-refresh-token     │      │     │           │             │
│           │ redirect    │         └─────────────────────┘      │     │           │             │
│           │ to intended │                                       │     │           │             │
│           ▼ destination │                                       │     │           │             │
│ ┌─────────────────────┐ │                                       │     │           │             │
│ │   /overview, etc.   │ │                                       │     │           │             │
│ └─────────────────────┘ │                                       │     │           │             │
└─────────────────────────┘                                       │     │           │             │
                                                                  └─────────────────────────┘
```

### Development URLs

| Environment | Dashboard | Auth Site | API |
|-------------|-----------|-----------|-----|
| **Local Dev** | http://localhost:3000 | http://localhost:4324 | http://localhost:8080 |
| **Production** | https://app.functionfly.com | https://auth.functionfly.com | https://api.functionfly.com |

## Flow

1. **Unauthenticated user hits ProtectedRoute** → Redirects to auth site with `redirect_uri` parameter
2. **User authenticates on auth site** → Auth site calls orchestrator API
3. **Auth site receives tokens** → Stores in `sessionStorage` (temporary)
4. **Auth site redirects back** → To dashboard `/auth/callback?redirect=/intended/path`
5. **Dashboard callback handler** → Migrates tokens from `sessionStorage` to `localStorage`
6. **Dashboard initializes session** → Calls `initialize()` to validate token and load user
7. **User redirected to intended destination** → Or `/onboarding` for new users

## Key Files

| File | Purpose |
|------|---------|
| `src/lib/auth-integration.ts` | URL builders, token migration, redirect helpers |
| `src/pages/AuthCallbackPage.tsx` | Receives tokens from auth site, initializes session |
| `src/stores/authStore.ts` | Modified to work with external auth; login/signup deprecated |
| `src/App.tsx` | Updated route protection; legacy routes redirect to auth site |

## Configuration

Add to your `.env`:

```env
# Optional: standalone auth site (@web/auth). Default shown below.
# - Dev: http://localhost:4324 (auth site dev server)
# - Prod: https://auth.functionfly.com
VITE_AUTH_SITE_URL=http://localhost:4324

# Optional: dashboard/app origin for auth redirects.
# This is where users return after authenticating on the auth site.
# - Dev: http://localhost:3000
# - Prod: https://app.functionfly.com
# If unset, defaults to window.location.origin (current URL)
VITE_APP_URL=http://localhost:3000
```

### Production Environment Variables

For production deployment to `app.functionfly.com`:

```env
VITE_API_URL=https://api.functionfly.com
VITE_AUTH_SITE_URL=https://auth.functionfly.com
VITE_APP_URL=https://app.functionfly.com
```

## API Changes

### Deprecated (redirects to auth site)

```typescript
// OLD - No longer works locally
authStore.login({ email, password });
authStore.signup({ email, password, ... });
authStore.verifyMFA(code);
```

### New Pattern

```typescript
// Redirect to auth site for login
import { redirectToAuthSite } from '@/lib/auth-integration';
redirectToAuthSite('/intended/path');

// Or just use ProtectedRoute which handles this automatically
```

## Token Storage

| Storage | Key | Purpose | Set By |
|---------|-----|---------|--------|
| `sessionStorage` | `ff_token` | Temporary token from auth site | Auth site callback |
| `sessionStorage` | `ff_refresh_token` | Temporary refresh token | Auth site callback |
| `localStorage` | `ff-access-token` | Primary token storage | Dashboard (migrated) |
| `localStorage` | `ff-refresh-token` | Primary refresh storage | Dashboard (migrated) |

Tokens are migrated from `sessionStorage` → `localStorage` by `AuthCallbackPage` and `migrateTokensFromSessionStorage()`.

## Route Changes

| Old Route | New Behavior |
|-----------|--------------|
| `/login` | Redirects to auth site |
| `/signup` | Redirects to auth site |
| `/auth/oauth/callback` | Redirects to auth site |
| `/auth/verify-email` | Handled by auth site |
| `/auth/reset-password` | Handled by auth site |
| `/auth/callback` | **NEW** - Receives tokens from auth site |

## Backward Compatibility

- Old `localStorage` tokens (`ff-access-token`, `ff-refresh-token`) continue to work
- The `initialize()` function still validates and refreshes tokens as before
- API client (`api/client.ts`) unchanged - still reads from `localStorage`
- If auth site is unreachable, the redirect will fail visibly (no silent failure)

## Migration Checklist

- [ ] Auth site deployed and accessible (`auth.functionfly.com` or `localhost:4324`)
- [ ] Orchestrator API has auth endpoints (`/v1/auth/*`)
- [ ] Dashboard `.env` has `VITE_AUTH_SITE_URL` set correctly
- [ ] Test login flow end-to-end
- [ ] Test signup flow with invite codes
- [ ] Test password reset flow
- [ ] Test OAuth flows (GitHub, Google)
- [ ] Test MFA/TOTP flows
- [ ] Verify old `/login` bookmarks redirect correctly
- [ ] Verify token refresh still works
- [ ] Test logout clears both storage types

## Troubleshooting

### "No authentication token found" error

- Check that auth site is running (`nx run functionfly-auth -t dev`)
- Check browser console for CORS errors
- Verify `VITE_AUTH_SITE_URL` matches the actual auth site origin

### Infinite redirect loop

- Check that auth site callback is redirecting to `/auth/callback` (not `/login`)
- Verify `AuthCallbackPage` is mounted in `App.tsx`
- Check that `initialize()` isn't throwing errors

### Tokens not persisting

- Check that `localStorage` is not disabled/blocked
- Verify no errors in `migrateTokensFromSessionStorage()`
- Check that `initialize()` validates the token successfully

## Dev Commands

```bash
# Start all services (from repo root)
bun dev

# Or start individually:
nx run functionfly-auth -t dev  # Auth site (port 4324)
nx run dashboard -t dev         # Dashboard (port 3000)
./bin/orchestrator-api --skip-migrations  # API (port 8080)
```

## Auth Site Repository

The standalone auth site is located at `web/auth/` in the monorepo:

```
web/auth/
├── src/
│   ├── pages/
│   │   ├── login.astro       # Login form
│   │   ├── signup.astro      # Registration form
│   │   ├── auth/callback.astro # Handles API tokens, redirects to dashboard
│   │   └── ...
│   └── config.ts             # SITE_ORIGIN, API_ORIGIN, APP_ORIGIN
```

See `web/auth/README.md` for auth site specific documentation.

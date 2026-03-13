# E2E Tests (Playwright)

## API Keys suite

The API Keys tests (`api-keys.spec.ts`) cover:

- **Security**: Unauthenticated users are redirected to login when visiting `/dashboard/api-keys` or `/settings`.
- **Create**: Form validation (name required), full create flow, success screen with one-time key and countdown.
- **List**: New key appears in the table after creation.
- **Copy**: Copy button shows "Copied!" feedback.
- **Delete**: Confirmation dialog, key removed from list, and deleted key does not reappear after refresh (active-only default).

## Prerequisites

1. **Backend** (orchestrator API) running on port **8080** with Postgres and Redis (see repo root `AGENTS.md`).
2. **Dashboard** started with API reachable (e.g. `VITE_API_URL=http://localhost:8080` or `/api` with Vite proxy).
3. **Test user** in the database. Default: `admin@functionfly.local` / `admin123`. Override with:
   - `E2E_LOGIN_EMAIL`
   - `E2E_LOGIN_PASSWORD`

## Run

```bash
cd web/dashboard
npx playwright test e2e/api-keys.spec.ts
```

With UI:

```bash
npx playwright test e2e/api-keys.spec.ts --ui
```

All e2e tests (including `example.spec.ts`):

```bash
npx playwright test
```

## CI

In CI, set `CI=1` so Playwright uses one worker and retries. Ensure the backend is running and the dashboard `webServer` (or `PLAYWRIGHT_BASE_URL`) points at it.

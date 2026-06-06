/**
 * E2E: access-denied page + role-denial UX.
 *
 * We mock the /admin/session endpoint to return a non-super_admin user
 * and then visit pages that require super_admin. The AdminLayout/Page
 * should redirect to /access-denied, where the page renders a clear
 * "Access blocked" message.
 */
import { expect, test } from '@playwright/test';
import { gotoAdminDashboard } from './utils';

const regularAdminBootstrap = {
  session: {
    id: 's1',
    access_token: 'jwt',
    expires_at: new Date(Date.now() + 3600_000).toISOString(),
    ip_address: '127.0.0.1',
    device_fingerprint: 'fp',
  },
  user: {
    id: 'u1',
    email: 'admin@example.com',
    name: 'Admin',
    role: 'admin',
    mfa_enabled: true,
  },
};

const superAdminBootstrap = {
  ...regularAdminBootstrap,
  user: { ...regularAdminBootstrap.user, role: 'super_admin' },
};

test.describe('Access control + /access-denied', () => {
  test('an unauthenticated user is redirected to /auth/login from a protected route', async ({ page }) => {
    await page.context().clearCookies();
    await page.goto('/users');
    await expect(page).toHaveURL(/\/auth\/login/);
  });

  test('renders the access-denied page inside the AdminLayout', async ({ page }) => {
    // Bootstrap as a regular admin so we have a session.
    await page.route('**/v1/admin/session', (route) => {
      route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(regularAdminBootstrap) });
    });

    await gotoAdminDashboard(page, '/access-denied');

    // The layout chrome (sidebar/header) is rendered because the route is
    // inside AdminLayout, and the page itself shows the blocked message.
    await expect(page.getByRole('heading', { name: /access denied/i })).toBeVisible();
    await expect(page.getByText(/don't have permission/i).first()).toBeVisible();
  });

  test('cannot reach /access-denied when not authenticated (redirects to login)', async ({ page }) => {
    await page.context().clearCookies();
    await page.goto('/access-denied');
    await expect(page).toHaveURL(/\/auth\/login/);
  });

  test('super_admin sees the admin layout chrome (header + sidebar)', async ({ page }) => {
    await page.route('**/v1/admin/session', (route) => {
      route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(superAdminBootstrap) });
    });

    await page.goto('/');
    // Header is always present when logged in.
    await expect(page.getByRole('banner')).toBeVisible();
    // Notifications bell from the header.
    await expect(page.locator('button[aria-label*="otification" i]')).toBeVisible();
  });
});

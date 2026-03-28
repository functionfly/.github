import { Page } from '@playwright/test';

/**
 * Test credentials for admin dashboard
 * SECURITY: Default credentials are only available in DEVELOPMENT mode
 */
function getTestCredentials() {
  if (process.env.NODE_ENV !== 'development' && process.env.DEVELOPMENT !== 'true') {
    throw new Error(
      'Test credentials should not be used outside DEVELOPMENT mode. ' +
        'Set DEVELOPMENT=true or provide credentials explicitly via environment variables.'
    );
  }
  return {
    email: process.env.TEST_ADMIN_EMAIL || 'admin@functionfly.local',
    password: process.env.TEST_ADMIN_PASSWORD || 'admin123',
  };
}

export const TEST_CREDENTIALS = getTestCredentials();

/**
 * Navigate to the admin dashboard and wait for it to load
 */
export async function gotoAdminDashboard(page: Page, path: string = '/') {
  await page.goto(path);
  await page.waitForLoadState('networkidle');
}

/**
 * Login to the admin dashboard
 */
export async function loginToAdmin(page: Page, email?: string, password?: string) {
  const { email: defaultEmail, password: defaultPassword } = TEST_CREDENTIALS;

  await gotoAdminDashboard(page, '/auth/login');
  await page.fill('input[name="email"]', email || defaultEmail);
  await page.fill('input[name="password"]', password || defaultPassword);
  await page.click('button[type="submit"]');
  await page.waitForURL(/\/dashboard/);
}

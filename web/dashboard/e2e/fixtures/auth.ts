import { Page } from '@playwright/test';

/**
 * Test credentials (AGENTS.md: admin test account).
 * Requires backend running with this user seeded.
 */
export const TEST_LOGIN = {
  email: process.env.E2E_LOGIN_EMAIL ?? 'admin@functionfly.local',
  password: process.env.E2E_LOGIN_PASSWORD ?? 'admin123',
};

/**
 * Log in via the login page and wait until the user is in the app (no longer on /login).
 * Use before API key (or other protected) e2e tests.
 */
export async function loginAsTestUser(page: Page): Promise<void> {
  await page.goto('/login');
  await page.getByLabel(/email/i).fill(TEST_LOGIN.email);
  await page.locator('#password').fill(TEST_LOGIN.password);
  await page.getByRole('main').getByRole('button', { name: 'Sign In' }).click();
  // Wait until we leave the login page (dashboard, onboarding, or MFA)
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 15000 });
}

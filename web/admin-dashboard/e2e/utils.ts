import { Page } from '@playwright/test';

/**
 * Test credentials for admin dashboard
 */
export const TEST_CREDENTIALS = {
  email: 'admin@functionfly.local',
  password: 'admin123',
};

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

  // Fill in the login form
  const emailInput = page.locator('input[type="email"], input[name="email"], input[id="email"]');
  const passwordInput = page.locator(
    'input[type="password"], input[name="password"], input[id="password"]'
  );
  const submitButton = page.locator(
    'button[type="submit"], button:has-text("Sign in"), button:has-text("Login")'
  );

  if (await emailInput.isVisible()) {
    await emailInput.fill(email || defaultEmail);
  }

  if (await passwordInput.isVisible()) {
    await passwordInput.fill(password || defaultPassword);
  }

  if (await submitButton.isVisible()) {
    await submitButton.click();
  }

  // Wait for navigation after login
  await page.waitForLoadState('networkidle');
}

/**
 * Check if user is logged in (not on login page)
 */
export async function isLoggedIn(page: Page): Promise<boolean> {
  const currentUrl = page.url();
  return !currentUrl.includes('/auth/login');
}

/**
 * Wait for the dashboard to be fully loaded
 */
export async function waitForDashboardLoad(page: Page) {
  await page.waitForLoadState('networkidle');

  // Wait for any loading spinners to disappear
  await page
    .waitForFunction(
      () => {
        const spinners = document.querySelectorAll('[class*="spinner"], [class*="loading"]');
        return spinners.length === 0;
      },
      { timeout: 10000 }
    )
    .catch(() => {
      // Ignore timeout - spinners might not exist
    });
}

/**
 * Take a screenshot with a descriptive name
 */
export async function takeScreenshot(page: Page, name: string) {
  await page.screenshot({
    path: `e2e/screenshots/${name}-${Date.now()}.png`,
    fullPage: true,
  });
}

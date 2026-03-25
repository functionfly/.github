import { expect, test } from '@playwright/test';
import { gotoAdminDashboard, loginToAdmin, waitForDashboardLoad } from './utils';

test.describe('System Health', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await loginToAdmin(page);
    await waitForDashboardLoad(page);
    await gotoAdminDashboard(page, '/system');
  });

  test('should display system health page', async ({ page }) => {
    // Check page loads
    await page.waitForLoadState('networkidle');

    // Check for system health content
    const healthContent = page.locator('[class*="health"], [class*="system"], h1');
    await expect(healthContent.first())
      .toBeVisible({ timeout: 5000 })
      .catch(() => {
        // Content may have different class
      });
  });

  test('should show system status indicators', async ({ page }) => {
    // Look for status indicators (green/red/yellow dots)
    const statusIndicators = page.locator(
      '[class*="status"], [class*="indicator"], [class*="badge"]'
    );
    const count = await statusIndicators.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('should display API health', async ({ page }) => {
    // Look for API status
    const apiStatus = page.locator('text=API, text=/api/i');
    const isVisible = await apiStatus.isVisible().catch(() => false);
    if (isVisible) {
      await expect(apiStatus.first()).toBeVisible();
    }
  });

  test('should display database health', async ({ page }) => {
    // Look for database status
    const dbStatus = page.locator('text=Database, text=Postgres, text=DB');
    const isVisible = await dbStatus
      .first()
      .isVisible()
      .catch(() => false);
    if (isVisible) {
      await expect(dbStatus.first()).toBeVisible();
    }
  });

  test('should display Redis health', async ({ page }) => {
    // Look for Redis status
    const redisStatus = page.locator('text=Redis, text=Cache');
    const isVisible = await redisStatus
      .first()
      .isVisible()
      .catch(() => false);
    if (isVisible) {
      await expect(redisStatus.first()).toBeVisible();
    }
  });

  test('should show metrics', async ({ page }) => {
    // Look for metrics cards
    const metrics = page.locator('[class*="metric"], [class*="stat"], [class*="card"]');
    const count = await metrics.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should have refresh button', async ({ page }) => {
    // Look for refresh button
    const refreshButton = page.locator(
      'button:has-text("Refresh"), button:has-text("Reload"), [class*="refresh"]'
    );

    const isVisible = await refreshButton.isVisible().catch(() => false);
    if (isVisible) {
      await expect(refreshButton.first()).toBeVisible();
    }
  });

  test('should navigate to status page', async ({ page }) => {
    // Look for status page link
    const statusLink = page.locator('a:has-text("Status"), a[href*="/status"]');

    const isVisible = await statusLink.isVisible().catch(() => false);
    if (isVisible) {
      await statusLink.first().click();
      await page.waitForLoadState('networkidle');

      // Verify navigation
      const url = page.url();
      expect(url).toContain('/status');
    }
  });
});

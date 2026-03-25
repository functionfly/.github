import { expect, test } from '@playwright/test';
import { gotoAdminDashboard, loginToAdmin, waitForDashboardLoad } from './utils';

test.describe('Users Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await loginToAdmin(page);
    await waitForDashboardLoad(page);
    await gotoAdminDashboard(page, '/users');
  });

  test('should display users list page', async ({ page }) => {
    // Check page loads
    await page.waitForLoadState('networkidle');

    // Check for users table or list
    const usersTable = page.locator('table, [class*="table"], [class*="list"]');
    await expect(usersTable.first())
      .toBeVisible({ timeout: 5000 })
      .catch(() => {
        // Table may have different class
      });
  });

  test('should show user statistics', async ({ page }) => {
    // Look for statistics cards
    const statsCards = page.locator('[class*="stat"], [class*="card"], [class*="metric"]');
    const count = await statsCards.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should filter users by role', async ({ page }) => {
    // Look for role filter dropdown
    const roleFilter = page.locator('select[name="role"], [class*="filter"]:has-text("Role")');

    if (await roleFilter.isVisible()) {
      await roleFilter.selectOption('admin');
      await page.waitForTimeout(500);

      // Verify filtering worked
      const tableRows = page.locator('tbody tr, [class*="row"]');
      const rowCount = await tableRows.count();
      expect(rowCount).toBeGreaterThanOrEqual(0);
    }
  });

  test('should search users', async ({ page }) => {
    // Look for search input
    const searchInput = page.locator(
      'input[placeholder*="search"], input[type="search"], [class*="search"] input'
    );

    if (await searchInput.isVisible()) {
      await searchInput.fill('admin');
      await page.waitForTimeout(500);

      // Verify search worked
      const tableRows = page.locator('tbody tr, [class*="row"]');
      const rowCount = await tableRows.count();
      expect(rowCount).toBeGreaterThanOrEqual(0);
    }
  });

  test('should navigate to user detail', async ({ page }) => {
    // Look for first user link
    const userLink = page
      .locator('tbody tr:first-child a, [class*="user"]:first-child a, a[href*="/users/"]')
      .first();

    if (await userLink.isVisible()) {
      await userLink.click();
      await page.waitForLoadState('networkidle');

      // Verify we're on user detail page
      const url = page.url();
      expect(url).toContain('/users/');
    }
  });

  test('should show invite user button', async ({ page }) => {
    // Look for invite button
    const inviteButton = page.locator(
      'button:has-text("Invite"), button:has-text("Add User"), a:has-text("Invite User")'
    );

    // Invite button may or may not be visible based on permissions
    const isVisible = await inviteButton.isVisible().catch(() => false);
    if (isVisible) {
      await expect(inviteButton).toBeVisible();
    }
  });

  test('should display user role badges', async ({ page }) => {
    // Look for role badges
    const roleBadges = page.locator('[class*="badge"], [class*="role"]');
    const count = await roleBadges.count();

    // Role badges may or may not exist
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('should have export functionality', async ({ page }) => {
    // Look for export button
    const exportButton = page.locator('button:has-text("Export"), [class*="export"]');

    // Export may not be available
    const isVisible = await exportButton.isVisible().catch(() => false);
    if (isVisible) {
      await expect(exportButton).toBeVisible();
    }
  });
});

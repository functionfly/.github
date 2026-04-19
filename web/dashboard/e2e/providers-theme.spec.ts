import { test, expect } from '@playwright/test';
import { loginAsTestUser } from './fixtures/auth';

test.describe('Providers Page Themes', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
  });

  test('Providers page loads and shows heading', async ({ page }) => {
    await page.goto('/providers');
    await expect(page).toHaveURL(/\/providers/);
    await expect(page.getByRole('heading', { name: 'Providers' }).first()).toBeVisible();
  });

  test('Density toggle buttons are visible', async ({ page }) => {
    await page.goto('/providers');

    // Check for density buttons
    await expect(page.getByRole('button', { name: /Compact/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /Comfort/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /Dashboard/i })).toBeVisible();
  });

  test('Glass morphism toggle is visible', async ({ page }) => {
    await page.goto('/providers');
    await expect(page.getByRole('button', { name: /Glass/i })).toBeVisible();
  });

  test('Status glow toggle is visible', async ({ page }) => {
    await page.goto('/providers');
    await expect(page.getByRole('button', { name: /Glow/i })).toBeVisible();
  });

  test('Provider cards are visible', async ({ page }) => {
    await page.goto('/providers');

    // Wait for cards to load
    await page.waitForTimeout(2000);

    // Check for provider cards (should have provider-card class or similar)
    const cards = page.locator('.provider-card, [class*="card"]').first();
    await expect(cards).toBeVisible({ timeout: 10000 });
  });

  test('Can toggle density modes', async ({ page }) => {
    await page.goto('/providers');
    await page.waitForTimeout(1000);

    // Click Dashboard density
    await page.getByRole('button', { name: /Dashboard/i }).click();
    await page.waitForTimeout(500);

    // Click Compact density
    await page.getByRole('button', { name: /Compact/i }).click();
    await page.waitForTimeout(500);

    // Click back to Comfortable
    await page.getByRole('button', { name: /Comfort/i }).click();
    await page.waitForTimeout(500);

    // Page should still be functional
    await expect(page.getByRole('heading', { name: 'Providers' })).toBeVisible();
  });

  test('Screenshot test - Providers page renders correctly', async ({ page }) => {
    await page.goto('/providers');
    await page.waitForTimeout(3000);

    // Take screenshot for visual verification
    await page.screenshot({ path: '/tmp/providers_test.png', fullPage: true });

    // Verify page is still functional
    await expect(page.getByRole('heading', { name: 'Providers' })).toBeVisible();
  });
});

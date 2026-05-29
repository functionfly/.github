import { test, expect } from '@playwright/test';

test.describe('FRG gallery', () => {
  test('loads the graphs landing page', async ({ page }) => {
    await page.goto('/frg');
    await expect(page.getByRole('heading', { name: /Function Runtime Graphs/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /Create Graph/i })).toBeVisible();
  });

  test('opens the new graph editor', async ({ page }) => {
    await page.goto('/frg/new');
    await expect(page.getByPlaceholder(/Graph name/i)).toBeVisible({ timeout: 15000 });
  });
});

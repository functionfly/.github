/**
 * E2E tests for the API Keys system.
 *
 * Prerequisites:
 * - Backend (orchestrator-api) running on port 8080 with Postgres + Redis.
 * - Dashboard started with VITE_API_URL pointing at backend (e.g. /api or http://localhost:8080).
 * - Test user seeded (default: admin@functionfly.local / admin123). Override with E2E_LOGIN_EMAIL, E2E_LOGIN_PASSWORD.
 *
 * Run: cd web/dashboard && npx playwright test e2e/api-keys.spec.ts
 */

import { test, expect } from '@playwright/test';
import { loginAsTestUser } from './fixtures/auth';

test.describe('API Keys', () => {
  test.describe('Security', () => {
    test('unauthenticated user is redirected to login when visiting API keys page', async ({
      page,
    }) => {
      await page.goto('/dashboard/api-keys');
      await expect(page).toHaveURL(/\/login/);
    });

    test('unauthenticated user cannot access API keys list via settings path', async ({
      page,
    }) => {
      await page.goto('/settings');
      await expect(page).toHaveURL(/\/login/);
    });
  });

  test.describe('Authenticated flows', () => {
    test.beforeEach(async ({ page }) => {
      await loginAsTestUser(page);
    });

    test('authenticated user can open API Keys page and see the section', async ({
      page,
    }) => {
      await page.goto('/dashboard/api-keys');
      await expect(page).toHaveURL(/\/api-keys/);
      await expect(page.getByRole('heading', { name: /API Keys/i }).first()).toBeVisible();
      await expect(page.getByRole('heading', { name: 'Your API Keys' })).toBeVisible();
    });

    test('create API key: form validation requires name', async ({ page }) => {
      await page.goto('/dashboard/api-keys');
      await page.getByRole('button', { name: 'Create API Key' }).first().click();
      await expect(page.getByRole('dialog', { name: /Create API Key/i })).toBeVisible();
      // Submit without name (button inside dialog)
      await page.getByRole('dialog').getByRole('button', { name: /^Create API Key$/i }).click();
      await expect(page.getByText(/name is required|Name is required/i)).toBeVisible();
    });

    test('create API key: full flow shows success and key once', async ({ page }) => {
      await page.goto('/dashboard/api-keys');
      const uniqueName = `e2e-key-${Date.now()}`;
      await page.getByRole('button', { name: 'Create API Key' }).first().click();
      await expect(page.getByRole('dialog', { name: /Create API Key/i })).toBeVisible();

      await page.getByRole('dialog').getByLabel(/^Name/i).fill(uniqueName);
      await page.getByRole('dialog').getByRole('button', { name: /^Create API Key$/i }).click();

      // Success view: modal shows "API Key Created" and the key value (prefix ffp_ for platform)
      await expect(page.getByRole('dialog', { name: /API Key Created/i })).toBeVisible({
        timeout: 10000,
      });
      await expect(page.getByText(/Copy it now as it will not be shown again/i)).toBeVisible();
      await expect(page.locator('code').filter({ hasText: /^ffp_/ })).toBeVisible();
    });

    test('create API key: list updates after creation', async ({ page }) => {
      await page.goto('/dashboard/api-keys');
      const uniqueName = `e2e-list-${Date.now()}`;
      await page.getByRole('button', { name: 'Create API Key' }).first().click();
      await page.getByRole('dialog').getByLabel(/^Name/i).fill(uniqueName);
      await page.getByRole('dialog').getByRole('button', { name: /^Create API Key$/i }).click();

      await expect(page.getByRole('dialog', { name: /API Key Created/i })).toBeVisible({
        timeout: 10000,
      });
      await page.getByRole('button', { name: /Done/i }).click();

      await expect(page.getByRole('dialog')).not.toBeVisible();
      await expect(page.getByRole('cell', { name: uniqueName })).toBeVisible({ timeout: 5000 });
    });

    test('copy button shows feedback after click', async ({ page }) => {
      await page.goto('/dashboard/api-keys');
      const uniqueName = `e2e-copy-${Date.now()}`;
      await page.getByRole('button', { name: 'Create API Key' }).first().click();
      await page.getByRole('dialog').getByLabel(/^Name/i).fill(uniqueName);
      await page.getByRole('dialog').getByRole('button', { name: /^Create API Key$/i }).click();

      await expect(page.getByRole('dialog', { name: /API Key Created/i })).toBeVisible({
        timeout: 10000,
      });
      await page.getByRole('dialog').getByRole('button', { name: /^Copy$/i }).click();
      await expect(
        page.getByRole('dialog').getByRole('button', { name: 'Copied!' }),
      ).toBeVisible();
    });

    test('delete key: confirmation required and key disappears from list', async ({
      page,
    }) => {
      await page.goto('/dashboard/api-keys');
      const uniqueName = `e2e-delete-${Date.now()}`;
      await page.getByRole('button', { name: 'Create API Key' }).first().click();
      await page.getByRole('dialog').getByLabel(/^Name/i).fill(uniqueName);
      await page.getByRole('dialog').getByRole('button', { name: /^Create API Key$/i }).click();
      await expect(page.getByRole('dialog', { name: /API Key Created/i })).toBeVisible({
        timeout: 10000,
      });
      await page.getByRole('button', { name: /Done/i }).click();

      await expect(page.getByRole('cell', { name: uniqueName })).toBeVisible({ timeout: 5000 });

      const row = page.getByRole('row').filter({ hasText: uniqueName });
      await row.getByRole('button', { name: /Delete/i }).click();
      await expect(page.getByRole('alertdialog', { name: /Delete API Key/i })).toBeVisible();
      await expect(page.getByRole('alertdialog').getByText(uniqueName)).toBeVisible();
      await page.getByRole('button', { name: /^Delete$/i }).click();

      await expect(page.getByRole('cell', { name: uniqueName })).not.toBeVisible({ timeout: 5000 });
    });

    test('deleted key does not reappear after page refresh', async ({ page }) => {
      await page.goto('/dashboard/api-keys');
      const uniqueName = `e2e-refresh-${Date.now()}`;
      await page.getByRole('button', { name: 'Create API Key' }).first().click();
      await page.getByRole('dialog').getByLabel(/^Name/i).fill(uniqueName);
      await page.getByRole('dialog').getByRole('button', { name: /^Create API Key$/i }).click();
      await expect(page.getByRole('dialog', { name: /API Key Created/i })).toBeVisible({
        timeout: 10000,
      });
      await page.getByRole('button', { name: /Done/i }).click();
      await expect(page.getByRole('cell', { name: uniqueName })).toBeVisible({ timeout: 5000 });

      const row = page.getByRole('row').filter({ hasText: uniqueName });
      await row.getByRole('button', { name: /Delete/i }).click();
      await page.getByRole('button', { name: /^Delete$/i }).click();
      await expect(page.getByRole('cell', { name: uniqueName })).not.toBeVisible({ timeout: 5000 });

      await page.reload();
      await expect(page.getByRole('heading', { name: /API Keys/i }).first()).toBeVisible({
        timeout: 5000,
      });
      await expect(page.getByRole('cell', { name: uniqueName })).not.toBeVisible();
    });

    test('countdown is visible on success screen', async ({ page }) => {
      await page.goto('/dashboard/api-keys');
      const uniqueName = `e2e-countdown-${Date.now()}`;
      await page.getByRole('button', { name: 'Create API Key' }).first().click();
      await page.getByRole('dialog').getByLabel(/^Name/i).fill(uniqueName);
      await page.getByRole('dialog').getByRole('button', { name: /^Create API Key$/i }).click();

      await expect(page.getByRole('dialog', { name: /API Key Created/i })).toBeVisible({
        timeout: 10000,
      });
      await expect(page.getByText(/Closes automatically in \d+s/)).toBeVisible();
    });
  });
});

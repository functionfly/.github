/**
 * E2E: MFA challenge + forgot-password flow.
 *
 * The spec exercises the login page UI without depending on a real
 * backend. We mock /v1/auth/login, /v1/auth/mfa/challenge, and
 * /v1/auth/forgot-password so the page renders the expected states.
 */
import { expect, test } from '@playwright/test';
import { gotoAdminDashboard } from './utils';

test.describe('MFA + forgot-password', () => {
  test.beforeEach(async ({ page, context }) => {
    await context.clearCookies();
  });

  test('shows the forgot-password form when "Forgot password?" is clicked', async ({ page }) => {
    await gotoAdminDashboard(page, '/auth/login');

    // The forgot link is rendered as a button under the password field.
    await page.getByRole('button', { name: /forgot password/i }).click();

    await expect(page.getByRole('heading', { name: /reset your password/i })).toBeVisible();
    await expect(page.locator('input#admin-forgot-email')).toBeVisible();
    await expect(page.getByRole('button', { name: /send reset link/i })).toBeVisible();
  });

  test('toggles back to the sign-in form', async ({ page }) => {
    await gotoAdminDashboard(page, '/auth/login');
    await page.getByRole('button', { name: /forgot password/i }).click();
    await expect(page.getByRole('heading', { name: /reset your password/i })).toBeVisible();
    await page.getByRole('button', { name: /back to sign in/i }).click();
    await expect(page.getByRole('heading', { name: /reset your password/i })).toBeHidden();
  });

  test('forgot-password shows a success message when the API returns 200', async ({ page }) => {
    await page.route('**/v1/auth/forgot-password', (route) => {
      route.fulfill({ status: 200, body: '{}' });
    });

    await gotoAdminDashboard(page, '/auth/login');
    await page.getByRole('button', { name: /forgot password/i }).click();
    await page.locator('input#admin-forgot-email').fill('admin@example.com');
    await page.getByRole('button', { name: /send reset link/i }).click();

    await expect(
      page.getByText(/if an account exists for that email, a password reset link has been sent/i)
    ).toBeVisible();
  });

  test('forgot-password shows a friendly error on a 500', async ({ page }) => {
    await page.route('**/v1/auth/forgot-password', (route) => {
      route.fulfill({ status: 500, body: 'oops' });
    });

    await gotoAdminDashboard(page, '/auth/login');
    await page.getByRole('button', { name: /forgot password/i }).click();
    await page.locator('input#admin-forgot-email').fill('admin@example.com');
    await page.getByRole('button', { name: /send reset link/i }).click();

    await expect(page.getByText(/unable to send reset email/i)).toBeVisible();
  });

  test('shows the MFA challenge modal when /auth/login returns mfa_required', async ({ page }) => {
    await page.route('**/v1/auth/login', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          mfa_required: true,
          challenge_token: 'challenge-abc',
        }),
      });
    });

    await gotoAdminDashboard(page, '/auth/login');
    await page.locator('input#admin-email').fill('admin@example.com');
    await page.locator('input#admin-password').fill('correct-horse-battery-staple');
    await page.getByRole('button', { name: /^sign in$/i }).click();

    const dialog = page.getByRole('dialog', { name: /two-factor required/i });
    await expect(dialog).toBeVisible();
    await expect(dialog.getByText(/6-digit code from your authenticator app/i)).toBeVisible();
    await expect(dialog.locator('input[aria-label="MFA code"]')).toBeVisible();
  });

  test('MFA challenge submit button is disabled until 6 digits are entered', async ({ page }) => {
    await page.route('**/v1/auth/login', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ mfa_required: true, challenge_token: 'c1' }),
      });
    });

    await gotoAdminDashboard(page, '/auth/login');
    await page.locator('input#admin-email').fill('a@b.c');
    await page.locator('input#admin-password').fill('hunter2');
    await page.getByRole('button', { name: /^sign in$/i }).click();

    const codeInput = page.getByRole('dialog').locator('input[aria-label="MFA code"]');
    const verifyBtn = page.getByRole('dialog').getByRole('button', { name: /verify/i });

    await expect(verifyBtn).toBeDisabled();
    await codeInput.fill('12345');
    await expect(verifyBtn).toBeDisabled();
    await codeInput.fill('123456');
    await expect(verifyBtn).toBeEnabled();
  });

  test('MFA challenge shows an error when the challenge endpoint rejects the code', async ({ page }) => {
    await page.route('**/v1/auth/login', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ mfa_required: true, challenge_token: 'c1' }),
      });
    });
    await page.route('**/v1/auth/mfa/challenge', (route) => {
      route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'invalid_code' }),
      });
    });

    await gotoAdminDashboard(page, '/auth/login');
    await page.locator('input#admin-email').fill('a@b.c');
    await page.locator('input#admin-password').fill('hunter2');
    await page.getByRole('button', { name: /^sign in$/i }).click();

    await page.getByRole('dialog').locator('input[aria-label="MFA code"]').fill('000000');
    await page.getByRole('dialog').getByRole('button', { name: /verify/i }).click();

    await expect(page.getByRole('dialog').getByText(/invalid code|invalid_code/i)).toBeVisible();
  });

  test('MFA modal can be closed via the close button', async ({ page }) => {
    await page.route('**/v1/auth/login', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ mfa_required: true, challenge_token: 'c1' }),
      });
    });

    await gotoAdminDashboard(page, '/auth/login');
    await page.locator('input#admin-email').fill('a@b.c');
    await page.locator('input#admin-password').fill('hunter2');
    await page.getByRole('button', { name: /^sign in$/i }).click();

    const dialog = page.getByRole('dialog', { name: /two-factor required/i });
    await expect(dialog).toBeVisible();
    await dialog.getByRole('button', { name: /close mfa prompt/i }).click();
    await expect(dialog).toBeHidden();
  });
});

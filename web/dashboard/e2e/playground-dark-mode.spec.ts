import { test, expect } from '@playwright/test';

test.describe('Playground Page Dark Mode', () => {
  test('playground page should have dark mode styles when dark theme is active', async ({ page }) => {
    await page.goto('/playground');

    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Check that data-theme attribute is set to 'dark'
    const htmlTheme = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
    console.log('HTML data-theme:', htmlTheme);

    // Get computed styles for key CSS variables
    const bgPrimary = await page.evaluate(() => {
      return getComputedStyle(document.documentElement).getPropertyValue('--bg-primary').trim();
    });
    console.log('--bg-primary value:', bgPrimary);

    const textPrimary = await page.evaluate(() => {
      return getComputedStyle(document.documentElement).getPropertyValue('--text-primary').trim();
    });
    console.log('--text-primary value:', textPrimary);

    // Verify theme is dark
    expect(htmlTheme).toBe('dark');

    // Check that the background color is dark (#0a0a0f = rgb(10, 10, 15))
    const bodyBg = await page.evaluate(() => {
      return getComputedStyle(document.body).backgroundColor;
    });
    console.log('body background-color:', bodyBg);

    // Verify it's a dark background (not white/light)
    // Dark backgrounds like #0a0a0f would be rgb(10, 10, 15)
    // Light backgrounds like #fafafa would be rgb(250, 250, 250)
    expect(bodyBg).not.toBe('rgb(250, 250, 250)');
    expect(bodyBg).not.toBe('rgb(255, 255, 255)');

    // Check that bg-primary CSS variable resolves to dark value
    expect(bgPrimary).toBe('#0a0a0f');
  });

  test('dark mode utility classes should work correctly', async ({ page }) => {
    await page.goto('/playground');
    await page.waitForLoadState('networkidle');

    // Check if theme variable references work in @theme block
    const colorBgPrimary = await page.evaluate(() => {
      return getComputedStyle(document.documentElement).getPropertyValue('--color-bg-primary').trim();
    });
    console.log('--color-bg-primary (from @theme):', colorBgPrimary);

    // This should resolve to the same value as --bg-primary since @theme maps it
    expect(colorBgPrimary).toBe('#0a0a0f');
  });

  test('theme toggle should switch between dark and light modes', async ({ page }) => {
    await page.goto('/playground');
    await page.waitForLoadState('networkidle');

    // Verify dark mode is active
    const initialTheme = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
    expect(initialTheme).toBe('dark');

    // Find and click theme toggle if available
    const themeToggle = page.locator('[data-theme-toggle], button[aria-label*="theme"], button[aria-label*="Theme"], button:has-text("theme"), button:has-text("Theme")').first();
    const toggleExists = await themeToggle.count() > 0;

    if (toggleExists) {
      await themeToggle.click();
      await page.waitForTimeout(500);

      const newTheme = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
      console.log('Theme after toggle:', newTheme);

      // Should switch to light mode
      expect(newTheme).toBe('light');

      // Background should now be light
      const bodyBg = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
      console.log('Light mode body background:', bodyBg);
      expect(bodyBg).toBe('rgb(250, 250, 250)');
    } else {
      console.log('No theme toggle found, skipping toggle test');
    }
  });
});
import { test, expect } from '@playwright/test';

test.describe('System Config Menu Test', () => {
  const PORTAL_URL = 'http://localhost:4200';

  test('should show system config menu items after login', async ({ page }) => {
    console.log('=== System Config Menu Test ===');

    // Navigate to portal and login
    await page.goto(PORTAL_URL);
    await page.waitForTimeout(3000);

    // Handle Dex login if redirected
    if (page.url().includes('/dex/auth')) {
      console.log('Logging in via Dex...');
      await page.waitForSelector('form', { timeout: 10000 });
      await page.locator('input[type="text"]').first().fill('admin@tanzu.local');
      await page.locator('input[type="password"]').first().fill('admin123');
      await page.locator('button[type="submit"]').first().click();
      await page.waitForTimeout(5000);
    }

    await page.waitForSelector('app-root', { timeout: 15000 });

    // Get all menu items from navigation
    const menuItems = await page.$$eval('a[routerlink]', els =>
      els.map(el => ({
        text: el.textContent?.trim(),
        routerlink: el.getAttribute('routerlink')
      }))
    );

    console.log('\n=== Available Menu Items ===');
    menuItems.forEach(item => {
      if (item.routerlink) {
        console.log(`- ${item.text}: ${item.routerlink}`);
      }
    });

    // Filter system config items
    const systemConfigs = menuItems.filter(item =>
      item.routerlink?.startsWith('system/')
    );

    console.log('\n=== System Config Menu Items ===');
    systemConfigs.forEach(item => {
      console.log(`- ${item.text}: ${item.routerlink}`);
    });

    // Verify all expected system config items exist
    const expectedItems = ['system/providers', 'system/lbs', 'system/repos'];
    const foundItems = systemConfigs.map(item => item.routerlink);

    console.log('\n=== Verification ===');
    expectedItems.forEach(expected => {
      const found = foundItems.includes(expected);
      console.log(`${expected}: ${found ? '✅ Found' : '❌ Not Found'}`);
      expect(found, `${expected} should be accessible`).toBe(true);
    });

    // Take screenshot
    await page.screenshot({ path: '/tmp/system-menu.png', fullPage: true });

    console.log('\n✅ All system config menu items are accessible after login');
  });
});

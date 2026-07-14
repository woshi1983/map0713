import { test, expect } from '@playwright/test';

test.describe('Health & Integration Tests', () => {
  const MAPD_URL = process.env.MAPD_URL || 'http://localhost:9800';
  const PORTAL_URL = process.env.PORTAL_URL || 'http://localhost:4200';

  test('should return healthy status from mapd', async ({ request }) => {
    const response = await request.get(`${MAPD_URL}/api/healthz`);
    expect(response.ok()).toBeTruthy();
    const result = await response.json();
    expect(result).toHaveProperty('ready');
    expect(result.ready).toBe(true);
  });

  test('should serve portal application', async ({ request }) => {
    const response = await request.get(PORTAL_URL);
    expect(response.ok()).toBeTruthy();
    const content = await response.text();
    expect(content).toContain('<html');
    expect(content).toContain('app-root');
  });
});

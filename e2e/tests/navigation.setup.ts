/**
 * One-time auth setup for navigation tests.
 * Logs in and saves the session to avoid hitting the rate limiter (burst=1, 5 req/min)
 * on each test's beforeEach. Reuses an existing session if it is still valid.
 */
import { test as setup, expect, request } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const AUTH_STATE_FILE = path.resolve(__dirname, '..', 'navigation-auth-state.json');

setup('authenticate for navigation tests', async ({ page, baseURL }) => {
    // Reuse the existing session if still valid — avoids the per-IP rate limiter
    if (fs.existsSync(AUTH_STATE_FILE)) {
        const ctx = await request.newContext({
            baseURL: baseURL || process.env.BASE_URL || 'http://localhost',
            storageState: AUTH_STATE_FILE,
        });
        const resp = await ctx.get('/api/v1/user');
        await ctx.dispose();
        if (resp.ok()) {
            return; // Session still valid — skip login
        }
    }

    await page.goto('/login');
    await page.fill('#email', 'test@test.com');
    await page.fill('#password', 'test');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/diary/, { timeout: 15000 });
    await page.context().storageState({ path: AUTH_STATE_FILE });
});

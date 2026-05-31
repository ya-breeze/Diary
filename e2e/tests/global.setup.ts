import { test as setup, expect, request } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const AUTH_STATE_FILE = path.resolve(__dirname, '..', 'auth-state.json');

setup('authenticate', async ({ page, baseURL }) => {
    if (fs.existsSync(AUTH_STATE_FILE)) {
        const ctx = await request.newContext({
            baseURL: baseURL || process.env.BASE_URL || 'http://localhost',
            storageState: AUTH_STATE_FILE,
        });
        const resp = await ctx.get('/api/v1/user');
        await ctx.dispose();
        if (resp.ok()) {
            return;
        }
    }

    await page.goto('/login');
    await page.fill('#email', 'test@test.com');
    await page.fill('#password', 'test');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/diary/, { timeout: 15000 });
    await page.context().storageState({ path: AUTH_STATE_FILE });
});

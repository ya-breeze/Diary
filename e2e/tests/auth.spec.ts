import { test, expect } from '@playwright/test';

const EMAIL = 'test@test.com';
const PASSWORD = 'test';

test.describe('Authentication', () => {
    test('valid login redirects to /diary', async ({ page }) => {
        await page.goto('/login');
        await page.fill('#email', EMAIL);
        await page.fill('#password', PASSWORD);
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL(/\/diary/, { timeout: 10000 });
    });

    test('wrong password shows error and stays on login', async ({ page }) => {
        await page.goto('/login');
        await page.fill('#email', EMAIL);
        await page.fill('#password', 'wrong-password');
        await page.click('button[type="submit"]');
        await expect(page.locator('.bg-red-50')).toBeVisible({ timeout: 5000 });
        await expect(page).toHaveURL(/\/login/);
    });

    test('accessing /diary while unauthenticated redirects to /login', async ({ page }) => {
        // Clear auth state by using a fresh context (no cookies/localStorage)
        await page.goto('/diary');
        // Dashboard layout calls validateSession() → fails → router.push('/login')
        await expect(page).toHaveURL(/\/login/, { timeout: 10000 });
    });
});

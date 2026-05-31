import { test, expect } from '@playwright/test';
import { ensureEntry } from './helpers';

const DATE_A = '2009-11-10';
const DATE_B = '2009-11-11';

test.describe('Date navigation', () => {
    test.beforeAll(async ({ browser }) => {
        const page = await browser.newPage();
        await ensureEntry(page, DATE_A, 'Nav Test Entry A');
        await ensureEntry(page, DATE_B, 'Nav Test Entry B');
        await page.close();
    });

    test('direct URL navigation loads the correct entry', async ({ page }) => {
        await page.goto(`/diary/${DATE_A}`);
        await expect(page).toHaveURL(new RegExp(`/diary/${DATE_A}`));
        await expect(page.getByRole('article').getByRole('heading', { name: 'Nav Test Entry A' })).toBeVisible({ timeout: 5000 });
    });

    test('"Next entry" link advances to the later date', async ({ page }) => {
        await page.goto(`/diary/${DATE_A}`);
        await expect(page.getByRole('article').getByRole('heading', { name: 'Nav Test Entry A' })).toBeVisible({ timeout: 5000 });
        await page.click('a[title="Next entry"]');
        await expect(page).toHaveURL(new RegExp(`/diary/${DATE_B}`), { timeout: 5000 });
        await expect(page.getByRole('article').getByRole('heading', { name: 'Nav Test Entry B' })).toBeVisible({ timeout: 5000 });
    });

    test('"Previous entry" link goes back to the earlier date', async ({ page }) => {
        await page.goto(`/diary/${DATE_B}`);
        await expect(page.getByRole('article').getByRole('heading', { name: 'Nav Test Entry B' })).toBeVisible({ timeout: 5000 });
        await page.click('a[title="Previous entry"]');
        await expect(page).toHaveURL(new RegExp(`/diary/${DATE_A}`), { timeout: 5000 });
        await expect(page.getByRole('article').getByRole('heading', { name: 'Nav Test Entry A' })).toBeVisible({ timeout: 5000 });
    });
});

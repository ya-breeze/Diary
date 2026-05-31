import { test, expect } from '@playwright/test';

const TEST_DATE = '2010-06-15';

test.describe('Diary entries', () => {
    test('create entry: write, save, reload, content persists', async ({ page }) => {
        // Navigate to a specific date in edit mode (creates a new entry)
        await page.goto(`/diary/${TEST_DATE}?edit=true`);

        // Fill in title and body
        await page.fill('input[placeholder="Enter a title..."]', 'E2E Test Entry');
        await page.fill('textarea[placeholder="Write your thoughts..."]', 'Body text from e2e test');

        // Save
        await page.click('button[type="submit"]:has-text("Save Changes")');

        // Should return to view mode at same date
        await expect(page).toHaveURL(new RegExp(`/diary/${TEST_DATE}`), { timeout: 10000 });

        // Reload and verify content is still there
        await page.reload();
        await expect(page.locator('h1, h2, h3').filter({ hasText: 'E2E Test Entry' })).toBeVisible({ timeout: 5000 });
    });

    test('edit existing entry: update title, save, verify update', async ({ page }) => {
        // Ensure entry exists first
        await page.goto(`/diary/${TEST_DATE}?edit=true`);
        await page.fill('input[placeholder="Enter a title..."]', 'Original Title');
        await page.fill('textarea[placeholder="Write your thoughts..."]', 'Original body');
        await page.click('button[type="submit"]:has-text("Save Changes")');
        await expect(page).toHaveURL(new RegExp(`/diary/${TEST_DATE}`), { timeout: 10000 });

        // Now edit it
        await page.click('button:has-text("Edit")');
        await expect(page).toHaveURL(new RegExp(`/diary/${TEST_DATE}\\?edit=true`), { timeout: 5000 });

        const titleInput = page.locator('input[placeholder="Enter a title..."]');
        await titleInput.clear();
        await titleInput.fill('Updated Title');
        await page.click('button[type="submit"]:has-text("Save Changes")');

        await expect(page).toHaveURL(new RegExp(`/diary/${TEST_DATE}`), { timeout: 10000 });
        await expect(page.locator('text=Updated Title')).toBeVisible({ timeout: 5000 });
    });

    test('navigating to date with no entry shows empty editor', async ({ page }) => {
        await page.goto('/diary/1990-01-01?edit=true');
        // Empty entry shows EntryEditor when navigating with ?edit=true
        await expect(page.locator('input[placeholder="Enter a title..."]')).toBeVisible({ timeout: 5000 });
    });
});

import { test, expect } from '@playwright/test';
import { ensureEntry } from './helpers';

const TEST_DATE = '2010-06-15';      // used by test 1 (create+persist)
const EDIT_TEST_DATE = '2010-06-16'; // used by test 2 (edit flow)
const BACK_NAV_DATE = '2010-06-17';  // used by test 4 (back-button history)

test.describe('Diary entries', () => {
    test('create entry: write, save, reload, content persists', async ({ page }) => {
        // Navigate to a specific date in edit mode (creates a new entry)
        await page.goto(`/diary/${TEST_DATE}?edit=true`);

        // Fill in title and body
        await page.fill('input[placeholder="Enter a title..."]', 'E2E Test Entry');
        await page.fill('textarea[placeholder="Write your thoughts..."]', 'Body text from e2e test');

        // Save
        await page.click('button[type="submit"]:has-text("Save Changes")');

        // Should return to view mode at same date (anchored to confirm edit mode exited)
        await expect(page).toHaveURL(new RegExp(`/diary/${TEST_DATE}$`), { timeout: 10000 });

        // Reload and verify content is still there
        await page.reload();
        await expect(page.getByRole('article').getByRole('heading', { name: 'E2E Test Entry' })).toBeVisible({ timeout: 5000 });
    });

    test('edit existing entry: update title, save, verify update', async ({ page }) => {
        await ensureEntry(page, EDIT_TEST_DATE, 'Original Title', 'Original body');

        // Now edit it
        await page.click('button:has-text("Edit")');
        await expect(page).toHaveURL(new RegExp(`/diary/${EDIT_TEST_DATE}\\?edit=true`), { timeout: 5000 });

        const titleInput = page.locator('input[placeholder="Enter a title..."]');
        await titleInput.clear();
        await titleInput.fill('Updated Title');
        await page.click('button[type="submit"]:has-text("Save Changes")');

        await expect(page).toHaveURL(new RegExp(`/diary/${EDIT_TEST_DATE}$`), { timeout: 10000 });
        await expect(page.getByRole('article').getByRole('heading', { name: 'Updated Title' })).toBeVisible({ timeout: 5000 });
    });

    test('navigating to date with no entry shows empty editor', async ({ page }) => {
        await page.goto('/diary/1990-01-01?edit=true');
        // Empty entry shows EntryEditor when navigating with ?edit=true
        await expect(page.locator('input[placeholder="Enter a title..."]')).toBeVisible({ timeout: 5000 });
    });

    test('back button after edit+save does not re-enter the editor', async ({ page }) => {
        await ensureEntry(page, BACK_NAV_DATE, 'Back Nav Entry', 'Back nav body');

        // Build a real history stack via in-app clicks: viewer -> edit -> save -> viewer.
        // Entering/leaving edit mode must replace (not push) history, so a single
        // Back press from the viewer must NOT land back in the editor.
        await page.goto(`/diary/${BACK_NAV_DATE}`);
        await expect(page.getByRole('article').getByRole('heading', { name: 'Back Nav Entry' })).toBeVisible({ timeout: 5000 });

        await page.click('button:has-text("Edit")');
        await expect(page).toHaveURL(new RegExp(`/diary/${BACK_NAV_DATE}\\?edit=true`), { timeout: 5000 });

        await page.click('button[type="submit"]:has-text("Save Changes")');
        await expect(page).toHaveURL(new RegExp(`/diary/${BACK_NAV_DATE}$`), { timeout: 10000 });

        // One Back press: editor must not reappear. Assert we land back on the
        // viewer (not ?edit=true, not some unexpected page) with no title input.
        await page.goBack();
        await expect(page).toHaveURL(new RegExp(`/diary/${BACK_NAV_DATE}$`), { timeout: 5000 });
        await expect(page.locator('input[placeholder="Enter a title..."]')).toHaveCount(0);
    });
});

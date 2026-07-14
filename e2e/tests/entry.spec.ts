import { test, expect } from '@playwright/test';
import { ensureEntry } from './helpers';

const TEST_DATE = '2010-06-15';      // used by test 1 (create+persist)
const EDIT_TEST_DATE = '2010-06-16'; // used by test 2 (edit flow)
const BACK_NAV_PREV_DATE = '2010-06-17'; // test 4: the back-target entry
const BACK_NAV_DATE = '2010-06-18';      // test 4: the entry we edit
const EMPTY_TITLE_DATE = '2010-06-19';  // test 5: empty title saves as Untitled

const MOBILE_TAGS_DATE = '2010-06-20';  // test 6: multiple tags remain visible on mobile
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

    test('empty title saves as Untitled', async ({ page }) => {
        await page.goto(`/diary/${EMPTY_TITLE_DATE}?edit=true`);
        await expect(page.locator('input[placeholder="Enter a title..."]')).toBeVisible({ timeout: 5000 });

        // Leave title blank, add body so the entry is not completely empty
        await page.fill('textarea[placeholder="Write your thoughts..."]', 'No title body');
        await page.click('button[type="submit"]:has-text("Save Changes")');

        await expect(page).toHaveURL(new RegExp(`/diary/${EMPTY_TITLE_DATE}$`), { timeout: 10000 });
        await expect(page.getByRole('article').getByRole('heading', { name: 'Untitled' })).toBeVisible({ timeout: 5000 });
    });

    test('all entry tags remain visible on a mobile viewport', async ({ page }) => {
        await page.goto(`/diary/${MOBILE_TAGS_DATE}?edit=true`);
        await page.fill('input[placeholder="Enter a title..."]', 'Mobile tag visibility');
        await page.fill('textarea[placeholder="Write your thoughts..."]', 'Tags should not be clipped.');

        const tags = page.locator('[data-testid="tags-input"]');
        for (const tag of ['cheerful', 'family', 'outdoors']) {
            await tags.fill(tag);
            await tags.press('Enter');
        }

        await page.click('button[type="submit"]:has-text("Save Changes")');
        await expect(page).toHaveURL(new RegExp(`/diary/${MOBILE_TAGS_DATE}$`), { timeout: 10000 });

        await page.setViewportSize({ width: 375, height: 667 });

        const entry = page.getByRole('article');
        for (const tag of ['cheerful', 'family', 'outdoors']) {
            const badge = entry.getByText(tag, { exact: true });
            await expect(badge).toBeVisible();
            await expect(badge).toBeInViewport();
        }
    });

    test('back button after edit+save does not re-enter the editor', async ({ page }) => {
        await ensureEntry(page, BACK_NAV_PREV_DATE, 'Prev Nav Entry', 'Prev body');
        await ensureEntry(page, BACK_NAV_DATE, 'Back Nav Entry', 'Back nav body');

        // Build a real back-target: land on the previous entry, then navigate to
        // the entry we edit (a distinct URL, so a genuine history entry is pushed).
        await page.goto(`/diary/${BACK_NAV_PREV_DATE}`);
        await expect(page.getByRole('article').getByRole('heading', { name: 'Prev Nav Entry' })).toBeVisible({ timeout: 5000 });
        await page.goto(`/diary/${BACK_NAV_DATE}`);
        await expect(page.getByRole('article').getByRole('heading', { name: 'Back Nav Entry' })).toBeVisible({ timeout: 5000 });

        // Edit then save. Entering/leaving edit mode must replace (not push)
        // history, so the editor does not accumulate a history entry.
        await page.click('button:has-text("Edit")');
        await expect(page).toHaveURL(new RegExp(`/diary/${BACK_NAV_DATE}\\?edit=true`), { timeout: 5000 });

        await page.click('button[type="submit"]:has-text("Save Changes")');
        await expect(page).toHaveURL(new RegExp(`/diary/${BACK_NAV_DATE}$`), { timeout: 10000 });

        // One Back press must return to the PREVIOUS entry, not re-enter the editor.
        await page.goBack();
        await expect(page).toHaveURL(new RegExp(`/diary/${BACK_NAV_PREV_DATE}$`), { timeout: 5000 });
        await expect(page.locator('input[placeholder="Enter a title..."]')).toHaveCount(0);
    });
});

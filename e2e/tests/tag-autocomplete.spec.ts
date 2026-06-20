import { test, expect } from '@playwright/test';

const TITLE_INPUT = 'input[placeholder="Enter a title..."]';
const TAGS_INPUT = 'input[placeholder="Enter tags separated by commas..."]';
const DROPDOWN = '[data-testid="tag-autocomplete"]';
const OPTION = '[data-testid="tag-autocomplete-option"]';

test.describe('Tag autocomplete', () => {
    // Seed an entry whose tag becomes part of the family vocabulary.
    test.beforeAll(async ({ browser }) => {
        const page = await browser.newPage();
        await page.goto('/diary/2012-08-09?edit=true');
        await page.fill(TITLE_INPUT, 'Seed for tags');
        await page.fill('textarea[placeholder="Write your thoughts..."]', 'body');
        await page.fill(TAGS_INPUT, 'travelmarker');
        await page.click('button[type="submit"]:has-text("Save Changes")');
        await expect(page).toHaveURL(/\/diary\/2012-08-09$/, { timeout: 10000 });
        await page.close();
    });

    test('typing a prefix shows a matching existing tag', async ({ page }) => {
        await page.goto('/diary/2012-08-10?edit=true');
        await expect(page.locator(TITLE_INPUT)).toBeVisible({ timeout: 5000 });

        const tags = page.locator(TAGS_INPUT);
        await tags.click();
        await tags.fill('travelm');
        await expect(page.locator(DROPDOWN)).toBeVisible({ timeout: 5000 });
        await expect(page.locator(OPTION, { hasText: 'travelmarker' })).toBeVisible();
    });

    test('selecting a suggestion completes the active token', async ({ page }) => {
        await page.goto('/diary/2012-08-11?edit=true');
        await expect(page.locator(TITLE_INPUT)).toBeVisible({ timeout: 5000 });

        const tags = page.locator(TAGS_INPUT);
        await tags.click();
        await tags.fill('work, trav');
        await page.locator(OPTION, { hasText: 'travelmarker' }).click();
        await expect(tags).toHaveValue('work, travelmarker, ');
    });

    test('already-entered tags are not offered again', async ({ page }) => {
        await page.goto('/diary/2012-08-12?edit=true');
        await expect(page.locator(TITLE_INPUT)).toBeVisible({ timeout: 5000 });

        const tags = page.locator(TAGS_INPUT);
        await tags.click();
        await tags.fill('travelmarker, trav');
        // The only existing match is already entered → no option for it.
        await expect(page.locator(OPTION, { hasText: 'travelmarker' })).toHaveCount(0);
    });
});

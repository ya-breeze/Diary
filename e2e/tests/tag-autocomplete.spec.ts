import { test, expect } from '@playwright/test';

const TITLE_INPUT = 'input[placeholder="Enter a title..."]';
const TAGS_INPUT = '[data-testid="tags-input"]';
const DROPDOWN = '[data-testid="tag-autocomplete"]';
const OPTION = '[data-testid="tag-autocomplete-option"]';
const CHIP = '[data-testid="tag-chip"]';
const CHIP_REMOVE = '[data-testid="tag-chip-remove"]';

test.describe('Tag autocomplete (chip input)', () => {
    // Seed an entry whose tag becomes part of the family vocabulary.
    test.beforeAll(async ({ browser }) => {
        const page = await browser.newPage();
        await page.goto('/diary/2012-08-09?edit=true');
        await page.fill(TITLE_INPUT, 'Seed for tags');
        await page.fill('textarea[placeholder="Write your thoughts..."]', 'body');
        const tags = page.locator(TAGS_INPUT);
        await tags.click();
        await tags.fill('travelmarker');
        await tags.press('Enter');
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

    test('selecting a suggestion adds it as a chip and keeps existing chips', async ({ page }) => {
        await page.goto('/diary/2012-08-11?edit=true');
        await expect(page.locator(TITLE_INPUT)).toBeVisible({ timeout: 5000 });

        const tags = page.locator(TAGS_INPUT);
        await tags.click();
        // Commit a brand-new tag first, then complete the existing one from the dropdown.
        await tags.fill('work');
        await tags.press('Enter');
        await tags.fill('trav');
        await page.locator(OPTION, { hasText: 'travelmarker' }).click();

        await expect(page.locator(CHIP)).toHaveCount(2);
        await expect(page.locator(CHIP, { hasText: 'work' })).toBeVisible();
        await expect(page.locator(CHIP, { hasText: 'travelmarker' })).toBeVisible();
        // The inline input is cleared after a selection.
        await expect(tags).toHaveValue('');
    });

    test('already-entered tags are not offered again', async ({ page }) => {
        await page.goto('/diary/2012-08-12?edit=true');
        await expect(page.locator(TITLE_INPUT)).toBeVisible({ timeout: 5000 });

        const tags = page.locator(TAGS_INPUT);
        await tags.click();
        await tags.fill('travelmarker');
        await tags.press('Enter');
        await tags.fill('trav');
        // The only existing match is already a chip → no option for it.
        await expect(page.locator(OPTION, { hasText: 'travelmarker' })).toHaveCount(0);
    });

    test('removing a chip drops only that tag', async ({ page }) => {
        await page.goto('/diary/2012-08-13?edit=true');
        await expect(page.locator(TITLE_INPUT)).toBeVisible({ timeout: 5000 });

        const tags = page.locator(TAGS_INPUT);
        await tags.click();
        await tags.fill('alpha');
        await tags.press('Enter');
        await tags.fill('beta');
        await tags.press('Enter');
        await expect(page.locator(CHIP)).toHaveCount(2);

        // Remove the first chip (alpha) via its X button.
        await page.locator(CHIP, { hasText: 'alpha' }).locator(CHIP_REMOVE).click();
        await expect(page.locator(CHIP)).toHaveCount(1);
        await expect(page.locator(CHIP, { hasText: 'beta' })).toBeVisible();
        await expect(page.locator(CHIP, { hasText: 'alpha' })).toHaveCount(0);
    });
});

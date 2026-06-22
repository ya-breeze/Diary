import { test, expect, type Page } from '@playwright/test';

// Exercises the search page's tag-chip filter: tag-only search below the 2-char
// text minimum, OR among tags, AND with text, and removing a chip. Tag names are
// suite-unique so assertions are deterministic against shared WIP data.

const TITLE_INPUT = 'input[placeholder="Enter a title..."]';
const BODY_INPUT = 'textarea[placeholder="Write your thoughts..."]';
const TAGS_INPUT = '[data-testid="tags-input"]';
const SEARCH_INPUT = 'input[placeholder="Search your journal..."]';
const SEARCH_TAG_INPUT = '[data-testid="search-tag-input"]';

async function seedEntry(page: Page, date: string, title: string, tags: string[]) {
    await page.goto(`/diary/${date}?edit=true`);
    await expect(page.locator(TITLE_INPUT)).toBeVisible({ timeout: 5000 });
    await page.fill(TITLE_INPUT, title);
    await page.fill(BODY_INPUT, 'search seed body');
    const input = page.locator(TAGS_INPUT);
    await input.click();
    for (const tag of tags) {
        await input.fill(tag);
        await input.press('Enter');
    }
    await page.click('button[type="submit"]:has-text("Save Changes")');
    await expect(page).toHaveURL(new RegExp(`/diary/${date}$`), { timeout: 10000 });
}

async function addSearchTag(page: Page, tag: string) {
    const input = page.locator(SEARCH_TAG_INPUT);
    await input.click();
    await input.fill(tag);
    await input.press('Enter');
    await expect(page.locator('[data-testid="search-tag-chip"]', { hasText: tag })).toBeVisible();
}

// Scope assertions to the main content (search results); the desktop sidebar
// also lists entry titles, which would otherwise match getByText.
function results(page: Page) {
    return page.getByRole('main');
}

test.describe('Search tag filter', () => {
    test.beforeAll(async ({ browser }) => {
        const page = await browser.newPage();
        // srchA → one entry; srchB → two entries (one of which also matches "extra").
        await seedEntry(page, '2014-01-01', 'SearchAlpha note', ['srchtaga']);
        await seedEntry(page, '2014-01-02', 'SearchBeta note', ['srchtagb']);
        await seedEntry(page, '2014-01-03', 'SearchAlpha extra', ['srchtagb']);
        await page.close();
    });

    test('tag-only filter runs below the 2-char text minimum', async ({ page }) => {
        await page.goto('/search');
        // No text typed at all — the tag chip alone drives the search.
        await addSearchTag(page, 'srchtaga');
        await expect(results(page).getByText('SearchAlpha note')).toBeVisible({ timeout: 5000 });
        await expect(results(page).getByText('SearchBeta note')).toHaveCount(0);
    });

    test('multiple tag chips use OR among tags', async ({ page }) => {
        await page.goto('/search');
        await addSearchTag(page, 'srchtaga');
        await addSearchTag(page, 'srchtagb');
        // srchtaga (one entry) OR srchtagb (two entries) → all three visible.
        await expect(results(page).getByText('SearchAlpha note')).toBeVisible({ timeout: 5000 });
        await expect(results(page).getByText('SearchBeta note')).toBeVisible();
        await expect(results(page).getByText('SearchAlpha extra')).toBeVisible();
    });

    test('text combines with tags using AND', async ({ page }) => {
        await page.goto('/search');
        await page.fill(SEARCH_INPUT, 'extra');
        await addSearchTag(page, 'srchtagb');
        // "extra" AND srchtagb → only the one entry that has both.
        await expect(results(page).getByText('SearchAlpha extra')).toBeVisible({ timeout: 5000 });
        await expect(results(page).getByText('SearchBeta note')).toHaveCount(0);
    });

    test('removing a tag chip widens the results', async ({ page }) => {
        await page.goto('/search');
        await page.fill(SEARCH_INPUT, 'extra');
        await addSearchTag(page, 'srchtaga');
        // "extra" AND srchtaga → no entry has both → no results.
        await expect(results(page).getByText('SearchAlpha extra')).toHaveCount(0);
        // Remove the chip → "extra" alone now matches the extra entry.
        await page.locator('[data-testid="search-tag-remove"]').click();
        await expect(results(page).getByText('SearchAlpha extra')).toBeVisible({ timeout: 5000 });
    });
});

import { test, expect } from '@playwright/test';

// Validates the ai-tagging "empty response informs the user" behavior: an
// interactive suggestion request that returns no tags shows an informational
// (non-error) toast rather than appearing to do nothing.

const AI_TOGGLE = '[data-testid="ai-tagging-toggle"]';
const SUGGEST_BUTTON = '[data-testid="suggest-tags-button"]';
const TITLE = 'input[placeholder="Enter a title..."]';
const INFO_TOAST = '[data-testid="toast-info"]';

async function setAiTagging(page: import('@playwright/test').Page, enabled: boolean) {
    await page.goto('/profile');
    const toggle = page.locator(AI_TOGGLE);
    await expect(toggle).toBeVisible({ timeout: 10000 });
    if ((await toggle.isChecked()) !== enabled) {
        await toggle.click();
        await expect(toggle).toBeChecked({ checked: enabled, timeout: 5000 });
    }
}

test.describe('AI tagging — empty suggestions', () => {
    test.afterAll(async ({ browser }) => {
        const page = await browser.newPage();
        await setAiTagging(page, false);
        await page.close();
    });

    test('empty suggestion result shows an informational message', async ({ page }) => {
        await setAiTagging(page, true);

        // Force the model to return no suggestions (a successful, empty result).
        await page.route('**/api/v1/items/suggest-tags', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: '{"tags":[]}',
            });
        });

        await page.goto('/diary/2011-03-04?edit=true');
        await expect(page.locator(TITLE)).toBeVisible({ timeout: 5000 });
        await page.fill(TITLE, 'A day with nothing taggable');
        await page.click(SUGGEST_BUTTON);

        const toast = page.locator(INFO_TOAST);
        await expect(toast).toBeVisible({ timeout: 5000 });
        await expect(toast).toContainText(/no tag suggestions/i);
    });
});

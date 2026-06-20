import { test, expect } from '@playwright/test';

// These tests validate the AI-tagging wiring deterministically: the per-family
// toggle persists, and the editor's "Suggest tags" control follows that setting.
// The actual suggestion content requires a live GEMINI_API_KEY and is
// non-deterministic, so it is intentionally not asserted here.

const SUGGEST_BUTTON = '[data-testid="suggest-tags-button"]';
const AI_TOGGLE = '[data-testid="ai-tagging-toggle"]';

async function setAiTagging(page: import('@playwright/test').Page, enabled: boolean) {
    await page.goto('/profile');
    const toggle = page.locator(AI_TOGGLE);
    await expect(toggle).toBeVisible({ timeout: 10000 });
    if ((await toggle.isChecked()) !== enabled) {
        await toggle.click();
        // Wait for the PATCH round-trip to settle on the new state.
        await expect(toggle).toBeChecked({ checked: enabled, timeout: 5000 });
    }
}

test.describe('AI tagging', () => {
    test.afterAll(async ({ browser }) => {
        // Leave the family setting disabled so other suites see default behavior.
        const page = await browser.newPage();
        await setAiTagging(page, false);
        await page.close();
    });

    test('toggle persists across reload', async ({ page }) => {
        await setAiTagging(page, true);
        await page.reload();
        await expect(page.locator(AI_TOGGLE)).toBeChecked();
    });

    test('suggest button shows in editor only when AI tagging is enabled', async ({ page }) => {
        await setAiTagging(page, true);
        await page.goto('/diary/2011-03-04?edit=true');
        await expect(page.locator('input[placeholder="Enter a title..."]')).toBeVisible({ timeout: 5000 });
        await expect(page.locator(SUGGEST_BUTTON)).toBeVisible({ timeout: 5000 });

        await setAiTagging(page, false);
        await page.goto('/diary/2011-03-04?edit=true');
        await expect(page.locator('input[placeholder="Enter a title..."]')).toBeVisible({ timeout: 5000 });
        await expect(page.locator(SUGGEST_BUTTON)).toHaveCount(0);
    });
});

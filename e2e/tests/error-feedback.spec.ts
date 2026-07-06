import { test, expect } from '@playwright/test';

// Validates the error-feedback capability: a failed user-initiated action
// surfaces a visible error toast instead of failing silently (regression guard
// for the "can't upload / no feedback" report).

const TITLE = 'input[placeholder="Enter a title..."]';
const BODY = 'textarea[placeholder="Write your thoughts..."]';
const SAVE = 'button[type="submit"]:has-text("Save Changes")';
const ERROR_TOAST = '[data-testid="toast-error"]';

test.describe('error feedback', () => {
    test('save failure shows an error toast', async ({ page }) => {
        await page.goto('/diary/2011-03-04?edit=true');
        await expect(page.locator(TITLE)).toBeVisible({ timeout: 5000 });
        await page.fill(TITLE, 'Error path');
        await page.fill(BODY, 'Trigger a save failure');

        // Force the save (PUT /api/v1/items, no query string) to fail. GET
        // requests carry a query string and so fall through this glob untouched.
        await page.route('**/api/v1/items', async (route) => {
            if (route.request().method() === 'PUT') {
                await route.fulfill({
                    status: 500,
                    contentType: 'application/json',
                    body: '{"error":"forced failure"}',
                });
            } else {
                await route.fallback();
            }
        });

        await page.click(SAVE);

        const toast = page.locator(ERROR_TOAST);
        await expect(toast).toBeVisible({ timeout: 5000 });
        await expect(toast).toContainText(/went wrong|server|try again/i);
    });
});

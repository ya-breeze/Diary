import { test, expect, type Page } from '@playwright/test';

// Exercises the Tags page: profile link, usage counts, browse-by-tag, rename
// (with merge-on-collision) and delete (with confirmation). Tag names are
// suite-unique so assertions are deterministic against shared WIP data.

const TITLE_INPUT = 'input[placeholder="Enter a title..."]';
const BODY_INPUT = 'textarea[placeholder="Write your thoughts..."]';
const TAGS_INPUT = '[data-testid="tags-input"]';

// Create (or overwrite) an entry on a date with the given title and tags.
async function seedEntry(page: Page, date: string, title: string, tags: string[]) {
    await page.goto(`/diary/${date}?edit=true`);
    await expect(page.locator(TITLE_INPUT)).toBeVisible({ timeout: 5000 });
    await page.fill(TITLE_INPUT, title);
    await page.fill(BODY_INPUT, 'seed body');
    const input = page.locator(TAGS_INPUT);
    await input.click();
    for (const tag of tags) {
        await input.fill(tag);
        await input.press('Enter');
    }
    await page.click('button[type="submit"]:has-text("Save Changes")');
    await expect(page).toHaveURL(new RegExp(`/diary/${date}$`), { timeout: 10000 });
}

// Locate the Tags-page row for a given tag name.
function tagRow(page: Page, name: string) {
    return page.locator('[data-testid="tag-row"]').filter({ hasText: name });
}

test.describe('Tag management', () => {
    test('profile Tags stat links to the Tags page', async ({ page }) => {
        await seedEntry(page, '2013-01-01', 'TM seed', ['tmbrowse']);
        await page.goto('/profile');
        await page.locator('[data-testid="tags-stat-card"]').click();
        await expect(page).toHaveURL(/\/tags$/, { timeout: 5000 });
        await expect(page.locator('h1', { hasText: 'Tags' })).toBeVisible();
    });

    test('browse-by-tag lists the tagged entry', async ({ page }) => {
        await seedEntry(page, '2013-02-02', 'TM browse target', ['tmbrowse2']);
        await page.goto('/tags');
        await tagRow(page, 'tmbrowse2').locator('[data-testid="tag-browse"]').click();
        // The browse view shows the seeded entry.
        await expect(page.getByText('TM browse target')).toBeVisible({ timeout: 5000 });
    });

    test('rename merges into an existing tag (collision)', async ({ page }) => {
        // typo on one entry, target on two others → after rename, target count = 3.
        await seedEntry(page, '2013-03-01', 'TM a', ['tmtypo']);
        await seedEntry(page, '2013-03-02', 'TM b', ['tmtarget']);
        await seedEntry(page, '2013-03-03', 'TM c', ['tmtarget']);

        await page.goto('/tags');
        const typoRow = tagRow(page, 'tmtypo');
        await expect(typoRow).toBeVisible({ timeout: 5000 });
        await typoRow.locator('[data-testid="tag-rename"]').click();
        // Once editing, the row shows an input (its value, not text), so target
        // the single active edit input at the page level.
        await page.locator('[data-testid="tag-rename-input"]').fill('tmtarget');
        await page.locator('[data-testid="tag-rename-save"]').click();

        // The typo is gone and the target now counts 3 entries (merged).
        await expect(tagRow(page, 'tmtypo')).toHaveCount(0, { timeout: 5000 });
        const targetRow = tagRow(page, 'tmtarget');
        await expect(targetRow.locator('[data-testid="tag-count"]')).toHaveText('3');
    });

    test('delete confirms with the affected entry count and removes the tag', async ({ page }) => {
        await seedEntry(page, '2013-04-01', 'TM d1', ['tmdelete']);
        await seedEntry(page, '2013-04-02', 'TM d2', ['tmdelete']);

        await page.goto('/tags');
        const row = tagRow(page, 'tmdelete');
        await expect(row.locator('[data-testid="tag-count"]')).toHaveText('2');
        await row.locator('[data-testid="tag-delete"]').click();

        // Confirmation references how many entries are affected.
        await expect(page.getByText(/from 2 entries/i)).toBeVisible({ timeout: 5000 });
        await page.locator('[data-testid="tag-delete-confirm"]').click();

        await expect(tagRow(page, 'tmdelete')).toHaveCount(0, { timeout: 5000 });
    });
});

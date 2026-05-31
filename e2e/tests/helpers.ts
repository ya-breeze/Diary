import { expect } from '@playwright/test';
import type { Page } from '@playwright/test';

export async function ensureEntry(page: Page, date: string, title: string, body?: string): Promise<void> {
    await page.goto(`/diary/${date}?edit=true`);
    await page.fill('input[placeholder="Enter a title..."]', title);
    await page.fill('textarea[placeholder="Write your thoughts..."]', body ?? `Entry for ${date}`);
    await page.click('button[type="submit"]:has-text("Save Changes")');
    await expect(page).toHaveURL(new RegExp(`/diary/${date}$`), { timeout: 10000 });
}

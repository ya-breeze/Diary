import { defineConfig, devices } from '@playwright/test';
import * as path from 'path';

export default defineConfig({
    testDir: './tests',
    timeout: 60000,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: 1,
    reporter: 'line',
    use: {
        baseURL: process.env.BASE_URL || 'http://localhost',
        trace: 'on-first-retry',
    },
    projects: [
        // Auth tests run without any stored session (tests login flow itself)
        {
            name: 'auth',
            testMatch: '**/auth.spec.ts',
            use: { ...devices['Desktop Chrome'] },
        },
        // One-time login shared by all non-auth tests
        {
            name: 'global-setup',
            testMatch: '**/global.setup.ts',
            use: { ...devices['Desktop Chrome'] },
        },
        // Entry tests reuse the saved session
        {
            name: 'entry',
            testMatch: '**/entry.spec.ts',
            dependencies: ['global-setup'],
            use: {
                ...devices['Desktop Chrome'],
                storageState: path.resolve(__dirname, 'auth-state.json'),
            },
        },
        // Navigation tests reuse the saved session
        {
            name: 'navigation',
            testMatch: '**/navigation.spec.ts',
            dependencies: ['global-setup'],
            use: {
                ...devices['Desktop Chrome'],
                storageState: path.resolve(__dirname, 'auth-state.json'),
            },
        },
        // AI tagging tests reuse the saved session
        {
            name: 'ai-tagging',
            testMatch: '**/ai-tagging.spec.ts',
            dependencies: ['global-setup'],
            use: {
                ...devices['Desktop Chrome'],
                storageState: path.resolve(__dirname, 'auth-state.json'),
            },
        },
        // Tag autocomplete tests reuse the saved session
        {
            name: 'tag-autocomplete',
            testMatch: '**/tag-autocomplete.spec.ts',
            dependencies: ['global-setup'],
            use: {
                ...devices['Desktop Chrome'],
                storageState: path.resolve(__dirname, 'auth-state.json'),
            },
        },
        // Tag management (Tags page: counts, browse, rename, delete)
        {
            name: 'tag-management',
            testMatch: '**/tag-management.spec.ts',
            dependencies: ['global-setup'],
            use: {
                ...devices['Desktop Chrome'],
                storageState: path.resolve(__dirname, 'auth-state.json'),
            },
        },
    ],
});

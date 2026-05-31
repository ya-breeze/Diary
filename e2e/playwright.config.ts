import { defineConfig, devices } from '@playwright/test';
import * as path from 'path';

export default defineConfig({
    testDir: './tests',
    fullyParallel: false,
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
        // Auth tests run without any stored session
        {
            name: 'auth',
            testMatch: '**/auth.spec.ts',
            use: { ...devices['Desktop Chrome'] },
        },
        // One-time login to save session for entry tests
        {
            name: 'entry-setup',
            testMatch: '**/entry.setup.ts',
            use: { ...devices['Desktop Chrome'] },
        },
        // Entry tests reuse the saved session
        {
            name: 'entry',
            testMatch: '**/entry.spec.ts',
            dependencies: ['entry-setup'],
            use: {
                ...devices['Desktop Chrome'],
                storageState: path.resolve(__dirname, 'entry-auth-state.json'),
            },
        },
        // One-time login to save session for navigation tests
        {
            name: 'navigation-setup',
            testMatch: '**/navigation.setup.ts',
            use: { ...devices['Desktop Chrome'] },
        },
        // Navigation tests reuse the saved session
        {
            name: 'navigation',
            testMatch: '**/navigation.spec.ts',
            dependencies: ['navigation-setup'],
            use: {
                ...devices['Desktop Chrome'],
                storageState: path.resolve(__dirname, 'navigation-auth-state.json'),
            },
        },
    ],
});

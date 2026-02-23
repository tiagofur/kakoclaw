import { defineConfig, devices } from '@playwright/test'

/**
 * Playwright E2E configuration.
 *
 * Tests run against the live app at http://127.0.0.1:18880.
 * Start the app before running: `makoclaw.exe web`
 *
 * Credentials used by tests come from environment variables with sensible
 * defaults matching the dev config.json.
 *
 *   MAKOCLAW_URL       defaults to http://127.0.0.1:18880
 *   MAKOCLAW_ADMIN     defaults to admin
 *   MAKOCLAW_PASSWORD  defaults to kako9812claw
 */
export default defineConfig({
  testDir: './e2e/tests',
  fullyParallel: false,   // keep sequential — tests share app state
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: [
    ['list'],
    ['html', { outputFolder: 'playwright-report', open: 'never' }],
  ],
  use: {
    baseURL: process.env.MAKOCLAW_URL ?? 'http://127.0.0.1:18880',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'off',
    // Persist auth state across tests via storageState
    storageState: undefined,
  },
  projects: [
    // Setup project: log in once and save the auth token to disk
    {
      name: 'setup',
      testMatch: /.*\.setup\.ts/,
    },
    // Main test suite (depends on setup)
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        storageState: 'e2e/.auth/admin.json',
      },
      dependencies: ['setup'],
      testMatch: /.*\.spec\.ts/,
    },
  ],
})

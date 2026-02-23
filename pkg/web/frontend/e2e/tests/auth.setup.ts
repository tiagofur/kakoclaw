/**
 * auth.setup.ts
 *
 * Logs in as admin once and saves the browser storage state (JWT token) to
 * e2e/.auth/admin.json so that all other test files can reuse the session
 * without re-logging in each time.
 */
import { test as setup, expect } from '@playwright/test'

const ADMIN = process.env.MAKOCLAW_ADMIN ?? 'admin'
const PASSWORD = process.env.MAKOCLAW_PASSWORD ?? 'kako9812claw'
const AUTH_FILE = 'e2e/.auth/admin.json'

setup('authenticate as admin', async ({ page }) => {
  await page.goto('/login')

  // Fill in credentials
  await page.locator('#username').fill(ADMIN)
  await page.locator('#password').fill(PASSWORD)
  await page.locator('button[type="submit"]').click()

  // Wait until redirected to dashboard
  await page.waitForURL('**/dashboard', { timeout: 10_000 })
  await expect(page).toHaveURL(/dashboard/)

  // Save storage state (cookies + localStorage with JWT)
  await page.context().storageState({ path: AUTH_FILE })
})

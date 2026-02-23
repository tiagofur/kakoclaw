/**
 * helpers.ts — shared utilities for Playwright E2E tests.
 */
import { type Page, expect } from '@playwright/test'

export const BASE_URL = process.env.MAKOCLAW_URL ?? 'http://127.0.0.1:18880'
export const ADMIN = process.env.MAKOCLAW_ADMIN ?? 'admin'
export const PASSWORD = process.env.MAKOCLAW_PASSWORD ?? 'admin123'
export const API = `${BASE_URL}/api/v1`

/** Navigate to a named route and wait for the page to settle. */
export async function goTo(page: Page, path: string) {
  await page.goto(path)
  await page.waitForLoadState('networkidle')
}

/** Assert a toast / status message containing the given text is visible. */
export async function expectToast(page: Page, text: string | RegExp) {
  // The app uses various notification patterns; try common selectors
  const toast = page.locator('[role="alert"], .toast, .notification, .alert').filter({ hasText: text })
  await expect(toast.first()).toBeVisible({ timeout: 8_000 })
}

/** Wait for an API response matching a URL pattern. */
export async function waitForAPI(page: Page, urlPattern: string | RegExp, method = 'GET') {
  return page.waitForResponse(
    resp => {
      const matches = typeof urlPattern === 'string'
        ? resp.url().includes(urlPattern)
        : urlPattern.test(resp.url())
      return matches && resp.request().method() === method
    },
    { timeout: 15_000 }
  )
}

/** Get a JWT token directly from the API (for use in API-level helpers). */
export async function getToken(page: Page): Promise<string> {
  const token = await page.evaluate(() => localStorage.getItem('token') ?? localStorage.getItem('jwt') ?? '')
  return token
}

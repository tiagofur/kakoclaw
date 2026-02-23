/**
 * auth.spec.ts — Authentication tests
 *
 * Covers: login, invalid credentials, logout, redirect behaviour, signup.
 * Runs with a fresh page context (no stored auth) for the no-auth cases.
 */
import { test, expect, type Page } from '@playwright/test'

const ADMIN = process.env.MAKOCLAW_ADMIN ?? 'admin'
const PASSWORD = process.env.MAKOCLAW_PASSWORD ?? 'kako9812claw'

// ─── Unauthenticated tests (override storageState to empty) ─────────────────
test.describe('Login page', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('redirects to /landing when not authenticated', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/(landing|login)/)
  })

  test('renders the login form', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('#username')).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('shows error on invalid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.locator('#username').fill('nobody')
    await page.locator('#password').fill('wrongpassword')
    await page.locator('button[type="submit"]').click()

    // Error message should appear — stay on login page
    // Server returns plain-text "invalid credentials"; Vue falls back to "Login failed. Please try again."
    await expect(
      page.locator('text=invalid credentials').or(page.locator('text=Login failed'))
    ).toBeVisible({ timeout: 8_000 })
    await expect(page).not.toHaveURL(/dashboard/)
  })

  test('successful login redirects to dashboard', async ({ page }) => {
    await page.goto('/login')
    await page.locator('#username').fill(ADMIN)
    await page.locator('#password').fill(PASSWORD)
    await page.locator('button[type="submit"]').click()

    await page.waitForURL('**/dashboard', { timeout: 10_000 })
    await expect(page).toHaveURL(/dashboard/)
  })

  test('login stores token in localStorage', async ({ page }) => {
    await page.goto('/login')
    await page.locator('#username').fill(ADMIN)
    await page.locator('#password').fill(PASSWORD)
    await page.locator('button[type="submit"]').click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })

    const token = await page.evaluate(() =>
      localStorage.getItem('auth.token') ?? ''
    )
    expect(token.length).toBeGreaterThan(20)
  })

  test('already-authenticated users are redirected away from /login', async ({ page }) => {
    // Log in first
    await page.goto('/login')
    await page.locator('#username').fill(ADMIN)
    await page.locator('#password').fill(PASSWORD)
    await page.locator('button[type="submit"]').click()
    await page.waitForURL('**/dashboard', { timeout: 10_000 })

    // Now try to visit /login again
    await page.goto('/login')
    await expect(page).not.toHaveURL(/\/login/)
  })
})

// ─── Logout (uses stored admin auth from setup) ──────────────────────────────
test.describe('Logout', () => {
  test('logout clears session and redirects to landing/login', async ({ page }) => {
    await page.goto('/dashboard')
    await page.waitForLoadState('networkidle')

    // Find and click the logout button — could be in a sidebar, nav menu or dropdown
    const logoutBtn = page.locator(
      'button:has-text("Logout"), button:has-text("Sign out"), a:has-text("Logout"), [data-testid="logout"]'
    ).first()

    // Open user/profile menu if logout is inside a dropdown
    const profileMenu = page.locator(
      '[data-testid="user-menu"], button:has-text("admin"), [aria-label="User menu"]'
    ).first()

    if (await profileMenu.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await profileMenu.click()
    }

    await expect(logoutBtn).toBeVisible({ timeout: 5_000 })
    await logoutBtn.click()

    // Should land on landing or login page
    await expect(page).toHaveURL(/\/(landing|login|$)/, { timeout: 8_000 })

    // Token should be cleared
    const token = await page.evaluate(() =>
      localStorage.getItem('auth.token') ?? ''
    )
    expect(token).toBe('')
  })
})

// ─── Signup ──────────────────────────────────────────────────────────────────
test.describe('Signup', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('renders signup form with required fields', async ({ page }) => {
    await page.goto('/signup')
    await expect(page.locator('#username')).toBeVisible()
    await expect(page.locator('#email')).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('rejects short password', async ({ page }) => {
    await page.goto('/signup')
    await page.locator('#username').fill('testshortpw')
    await page.locator('#email').fill('testshortpw@example.com')
    await page.locator('#password').fill('short')
    await page.locator('#confirmPassword').fill('short')
    await page.locator('button[type="submit"]').click()

    // Should stay on signup — browser validation or server error
    await expect(page).not.toHaveURL(/dashboard/, { timeout: 5_000 })
  })

  test('successful signup auto-logs in and redirects to dashboard or onboarding', async ({ page }) => {
    const unique = Date.now()
    await page.goto('/signup')
    await page.locator('#username').fill(`e2euser${unique}`)
    await page.locator('#email').fill(`e2euser${unique}@example.com`)
    await page.locator('#password').fill('TestPass123!')
    await page.locator('#confirmPassword').fill('TestPass123!')
    await page.locator('button[type="submit"]').click()

    await page.waitForURL(/\/(dashboard|onboarding)/, { timeout: 12_000 })
    await expect(page).toHaveURL(/\/(dashboard|onboarding)/)
  })
})

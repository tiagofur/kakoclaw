/**
 * settings.spec.ts — SettingsView E2E tests
 *
 * Covers: page load, tab navigation (Profile / Providers / Channels /
 * Agents / System), profile field save, change-password modal open/close,
 * and provider config tab rendering.
 *
 * Pre-condition: admin auth state restored from e2e/.auth/admin.json.
 */
import { test, expect } from '@playwright/test'
import { goTo } from '../fixtures/helpers'

test.describe('SettingsView — /settings', () => {
  test('loads the Settings page with tabs', async ({ page }) => {
    await goTo(page, '/settings')
    await expect(page.locator('h2:has-text("Settings")')).toBeVisible()
    // Tab bar has Profile tab active by default
    await expect(page.locator('button:has-text("Profile")').first()).toBeVisible()
  })

  test('Profile tab is active by default', async ({ page }) => {
    await goTo(page, '/settings')
    // ProfileSettingsTab renders "Profile Information" heading
    await expect(page.locator('text=Profile Information')).toBeVisible({ timeout: 8_000 })
  })

  test('profile form shows username and email fields', async ({ page }) => {
    await goTo(page, '/settings')
    // Wait for profile data to load
    await expect(page.locator('text=Profile Information')).toBeVisible({ timeout: 8_000 })
    // Username and email inputs (no id, identified by label text)
    const usernameInput = page.locator('input[type="text"]').first()
    const emailInput = page.locator('input[type="email"]').first()
    await expect(usernameInput).toBeVisible()
    await expect(emailInput).toBeVisible()
  })

  test('Save Profile button is present', async ({ page }) => {
    await goTo(page, '/settings')
    await expect(page.locator('text=Profile Information')).toBeVisible({ timeout: 8_000 })
    await expect(page.locator('button:has-text("Save Profile")')).toBeVisible()
  })

  test('Change Password button opens the modal', async ({ page }) => {
    await goTo(page, '/settings')
    await expect(page.locator('text=Profile Information')).toBeVisible({ timeout: 8_000 })
    await page.locator('button:has-text("Change Password")').click()
    // Modal title
    await expect(page.locator('text=Change Password').nth(1)).toBeVisible({ timeout: 5_000 })
    // Three password inputs
    const pwInputs = page.locator('input[type="password"]')
    await expect(pwInputs).toHaveCount(3)
  })

  test('Change Password modal closes on Cancel', async ({ page }) => {
    await goTo(page, '/settings')
    await expect(page.locator('text=Profile Information')).toBeVisible({ timeout: 8_000 })
    await page.locator('button:has-text("Change Password")').click()
    await expect(page.locator('text=Change Password').nth(1)).toBeVisible({ timeout: 5_000 })
    await page.locator('button:has-text("Cancel")').click()
    await expect(page.locator('input[type="password"]')).toHaveCount(0, { timeout: 5_000 })
  })

  test('Agents tab loads', async ({ page }) => {
    await goTo(page, '/settings?tab=agents')
    await expect(page.locator('button:has-text("Agents")')).toBeVisible()
    // AgentSettingsTab has an Orchestrator section
    await expect(
      page.locator('text=Orchestrator').or(page.locator('text=Specialist')).first()
    ).toBeVisible({ timeout: 8_000 })
  })

  test('Providers tab loads', async ({ page }) => {
    await goTo(page, '/settings?tab=providers')
    await expect(page.locator('button:has-text("Providers")')).toBeVisible()
    // ProvidersSettingsTab lists provider names
    await expect(
      page.locator('text=OpenAI').or(
        page.locator('text=Anthropic').or(page.locator('text=Provider'))
      ).first()
    ).toBeVisible({ timeout: 8_000 })
  })

  test('Channels tab loads', async ({ page }) => {
    await goTo(page, '/settings?tab=channels')
    await expect(page.locator('button:has-text("Channels")')).toBeVisible()
    // ChannelsSettingsTab lists channel names (Telegram, Discord, etc.)
    await expect(
      page.locator('text=Telegram').or(page.locator('text=Discord').or(page.locator('text=Channel'))).first()
    ).toBeVisible({ timeout: 8_000 })
  })

  test('System tab loads', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('button:has-text("System")')).toBeVisible()
    // System tab shows "Web Server" and "Backup & Restore" headings
    await expect(
      page.locator('text=Web Server').or(page.locator('text=Backup')).first()
    ).toBeVisible({ timeout: 8_000 })
  })

  test('System tab shows Backup & Restore section', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Backup & Restore')).toBeVisible({ timeout: 8_000 })
    await expect(page.locator('text=Export Backup')).toBeVisible()
    await expect(page.locator('text=Import Backup')).toBeVisible()
  })

  test('Download Backup button is present in System tab', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(
      page.locator('button:has-text("Download Backup")')
    ).toBeVisible({ timeout: 8_000 })
  })

  test('Users tab is visible for admin and shows user list', async ({ page }) => {
    await goTo(page, '/settings?tab=users')
    // Admin-only tab
    await expect(page.locator('button:has-text("Users")')).toBeVisible({ timeout: 5_000 })
    // User list table headings
    await expect(page.locator('th:has-text("Username")')).toBeVisible({ timeout: 8_000 })
    await expect(page.locator('th:has-text("Role")')).toBeVisible()
    // "admin" user row is present
    await expect(page.locator('td:has-text("admin")').first()).toBeVisible()
  })

  test('Add User button is present on Users tab (admin)', async ({ page }) => {
    await goTo(page, '/settings?tab=users')
    await expect(page.locator('button:has-text("Add User")')).toBeVisible({ timeout: 8_000 })
  })
})

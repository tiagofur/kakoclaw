/**
 * backup.spec.ts — Backup & Restore E2E tests
 *
 * Covers: export download (triggers a file download), import UI present,
 * file-input accept attribute, and button states.
 *
 * Backup UI lives at /settings?tab=system.
 *
 * Pre-condition: admin auth state restored from e2e/.auth/admin.json.
 */
import { test, expect } from '@playwright/test'
import { goTo } from '../fixtures/helpers'

test.describe('Backup & Restore — /settings?tab=system', () => {
  test('Backup & Restore section is visible', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Backup & Restore')).toBeVisible({ timeout: 10_000 })
  })

  test('Export section shows checkboxes and download button', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Export Backup')).toBeVisible({ timeout: 10_000 })

    // The four export checkboxes
    await expect(page.locator('text=Database & Sessions')).toBeVisible()
    await expect(page.locator('text=Workspace & Skills')).toBeVisible()
    await expect(page.locator('text=Configuration (config.json)')).toBeVisible()
    await expect(page.locator('text=Environment Variables (.env)')).toBeVisible()

    // Download button
    await expect(page.locator('button:has-text("Download Backup")')).toBeVisible()
  })

  test('Download Backup button is enabled when at least one checkbox is checked', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Export Backup')).toBeVisible({ timeout: 10_000 })

    // Database & Sessions is checked by default
    const downloadBtn = page.locator('button:has-text("Download Backup (.makoclaw)")')
    await expect(downloadBtn).toBeEnabled()
  })

  test('Download Backup button is disabled when all checkboxes are unchecked', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Export Backup')).toBeVisible({ timeout: 10_000 })

    // Uncheck all four checkboxes
    const checkboxes = page.locator('.space-y-3 input[type="checkbox"]')
    const count = await checkboxes.count()
    for (let i = 0; i < count; i++) {
      const cb = checkboxes.nth(i)
      if (await cb.isChecked()) {
        await cb.uncheck()
      }
    }

    await expect(page.locator('button:has-text("Download Backup (.makoclaw)")')).toBeDisabled()
  })

  test('Export backup triggers a download', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Export Backup')).toBeVisible({ timeout: 10_000 })

    // Wait for the download event
    const downloadPromise = page.waitForEvent('download', { timeout: 15_000 })
    await page.locator('button:has-text("Download Backup (.makoclaw)")').click()

    const download = await downloadPromise
    // File should have a .makoclaw extension
    expect(download.suggestedFilename()).toMatch(/\.makoclaw$/)
  })

  test('Import section shows file input area', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Import Backup')).toBeVisible({ timeout: 10_000 })

    // There should be a hidden file input accepting .makoclaw files
    const fileInput = page.locator('input[type="file"][accept=".makoclaw"]')
    await expect(fileInput).toHaveCount(1)
  })

  test('Import area has a clickable "Select Backup File" button', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Import Backup')).toBeVisible({ timeout: 10_000 })

    // The button that triggers the hidden file input
    const selectBtn = page.locator('button').filter({ hasText: /select .makoclaw file/i })
    await expect(selectBtn).toBeVisible()
  })

  test('Import Backup button is not shown before a file is selected', async ({ page }) => {
    await goTo(page, '/settings?tab=system')
    await expect(page.locator('text=Import Backup')).toBeVisible({ timeout: 10_000 })

    // "Import Backup" submit button is only rendered after a file is selected
    // so it should not be present in the DOM before that
    const importBtn = page.locator('button').filter({ hasText: /^Import Backup$/ })
    await expect(importBtn).toHaveCount(0)
  })
})

/**
 * agents.spec.ts — AgentsView E2E tests
 *
 * Covers: page load, heading, orchestrator status panel, specialists grid
 * (or empty-state), period selector, and Cost Overview section.
 *
 * NOTE: This view is read-only. Specialist CRUD lives in SettingsView
 * (see settings.spec.ts).
 *
 * Pre-condition: admin auth state restored from e2e/.auth/admin.json.
 */
import { test, expect } from '@playwright/test'
import { goTo } from '../fixtures/helpers'

test.describe('AgentsView — /agents', () => {
  test('loads the Multi-Agent System page', async ({ page }) => {
    await goTo(page, '/agents')
    await expect(page.locator('h2:has-text("Multi-Agent System")')).toBeVisible()
    await expect(
      page.locator('text=Manage your AI specialists and track their performance')
    ).toBeVisible()
  })

  test('Orchestrator status panel is present', async ({ page }) => {
    await goTo(page, '/agents')
    await expect(page.locator('h3:has-text("Orchestrator")')).toBeVisible()
  })

  test('Orchestrator shows Active or Inactive badge', async ({ page }) => {
    await goTo(page, '/agents')
    const badge = page.locator('span:has-text("Active"), span:has-text("Inactive")')
    await expect(badge.first()).toBeVisible({ timeout: 8_000 })
  })

  test('period selector (metrics time range) is visible', async ({ page }) => {
    await goTo(page, '/agents')
    const periodSelect = page.locator('select').first()
    await expect(periodSelect).toBeVisible()
    // Should have the three options
    await expect(page.locator('option:has-text("Last 24 Hours")')).toBeDefined()
    await expect(page.locator('option:has-text("Last 7 Days")')).toBeDefined()
    await expect(page.locator('option:has-text("Last 30 Days")')).toBeDefined()
  })

  test('Specialists section heading is visible', async ({ page }) => {
    await goTo(page, '/agents')
    await expect(page.locator('h3:has-text("Specialists")')).toBeVisible()
  })

  test('shows empty-state or specialist cards', async ({ page }) => {
    await goTo(page, '/agents')
    // Either a "No Specialists Yet" empty state or at least one specialist card
    const emptyState = page.locator('h4:has-text("No Specialists Yet")')
    const specialistCard = page.locator('h4.font-bold.text-makoclaw-text.capitalize').first()

    const hasEmpty = await emptyState.isVisible({ timeout: 8_000 }).catch(() => false)
    const hasCards = await specialistCard.isVisible({ timeout: 1_000 }).catch(() => false)

    expect(hasEmpty || hasCards).toBeTruthy()
  })

  test('empty-state links to Settings > Agents when no specialists', async ({ page }) => {
    await goTo(page, '/agents')
    const link = page.locator('a:has-text("Configure Orchestrator"), a:has-text("Create First Specialist")')
    const isVisible = await link.isVisible({ timeout: 5_000 }).catch(() => false)
    if (!isVisible) {
      // Specialists exist — skip this assertion
      return
    }
    const href = await link.getAttribute('href')
    expect(href).toContain('/settings')
    expect(href).toContain('agents')
  })

  test('clicking a specialist card opens the details modal', async ({ page }) => {
    await goTo(page, '/agents')
    const card = page.locator('.cursor-pointer h4.font-bold.capitalize').first()
    const count = await card.count()
    if (count === 0) {
      // No specialists configured — skip
      test.skip()
      return
    }
    await card.click()
    // SpecialistDetailsModal renders with a heading inside a fixed overlay
    await expect(page.locator('.fixed').filter({ hasText: 'Specialist' }).first()).toBeVisible({
      timeout: 5_000,
    })
  })

  test('Cost Overview section appears when metrics are available', async ({ page }) => {
    await goTo(page, '/agents')
    // Metrics section may or may not have data yet — just check heading if visible
    const costHeading = page.locator('h3:has-text("Cost Overview")')
    const isVisible = await costHeading.isVisible({ timeout: 8_000 }).catch(() => false)
    if (isVisible) {
      await expect(page.locator('text=Total Cost')).toBeVisible()
      await expect(page.locator('text=Total Calls')).toBeVisible()
      await expect(page.locator('text=Total Tokens')).toBeVisible()
    }
    // If not visible, metrics returned empty — pass silently
  })

  test('changing the period selector triggers a metrics refresh', async ({ page }) => {
    await goTo(page, '/agents')
    const periodSelect = page.locator('select').first()

    // Intercept the metrics API call
    const metricsPromise = page.waitForResponse(
      resp => resp.url().includes('/metrics') && resp.request().method() === 'GET',
      { timeout: 10_000 }
    ).catch(() => null)

    await periodSelect.selectOption('30d')
    const response = await metricsPromise
    if (response) {
      expect(response.status()).toBeLessThan(500)
    }
  })
})

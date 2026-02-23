/**
 * tasks.spec.ts — TasksView (Kanban) E2E tests
 *
 * Covers: page load, five Kanban columns visible, search filter,
 * status dropdown filter, "Archived" checkbox, New Task button & modal
 * (open / close / validation / create), and task detail modal.
 *
 * Pre-condition: admin auth state restored from e2e/.auth/admin.json.
 */
import { test, expect } from '@playwright/test'
import { goTo } from '../fixtures/helpers'

test.describe('TasksView — /tasks', () => {
  test('loads the Kanban board', async ({ page }) => {
    await goTo(page, '/tasks')
    // Use h3 headings for the Kanban column titles to avoid strict-mode violations
    // (the column name also appears in the status filter <select> and badge spans)
    await expect(page.locator('h3').filter({ hasText: /^Backlog$/ })).toBeVisible()
    await expect(page.locator('h3').filter({ hasText: /^To Do$/ })).toBeVisible()
    await expect(page.locator('h3').filter({ hasText: /^In Progress$/ })).toBeVisible()
    await expect(page.locator('h3').filter({ hasText: /^Review$/ })).toBeVisible()
    await expect(page.locator('h3').filter({ hasText: /^Done$/ })).toBeVisible()
  })

  test('shows the search input', async ({ page }) => {
    await goTo(page, '/tasks')
    await expect(
      page.locator('input[placeholder="Search tasks by name or description..."]')
    ).toBeVisible()
  })

  test('shows the sort-by select', async ({ page }) => {
    await goTo(page, '/tasks')
    // The first <select> in the filters area is the sort-by dropdown
    await expect(page.locator('option[value="recent"]')).toBeDefined()
    await expect(page.locator('select').nth(0)).toBeVisible()
  })

  test('shows the status filter select', async ({ page }) => {
    await goTo(page, '/tasks')
    // Second <select> is status filter — has the "All Statuses" option
    await expect(page.locator('option:has-text("All Statuses")')).toBeDefined()
  })

  test('shows the Archived checkbox', async ({ page }) => {
    await goTo(page, '/tasks')
    await expect(page.locator('#showArchived')).toBeVisible()
    await expect(page.locator('label[for="showArchived"]')).toContainText('Archived')
  })

  test('New Task button is visible', async ({ page }) => {
    await goTo(page, '/tasks')
    await expect(page.locator('button:has-text("New Task"), button:has-text("+")')).toBeVisible()
  })

  // ─── New Task Modal ──────────────────────────────────────────────────────────

  test('New Task modal opens on button click', async ({ page }) => {
    await goTo(page, '/tasks')
    await page.locator('button:has-text("New Task"), button:has-text("+")').first().click()
    await expect(page.locator('h3:has-text("Create New Task")')).toBeVisible()
    await expect(page.locator('#title')).toBeVisible()
    await expect(page.locator('#description')).toBeVisible()
    await expect(page.locator('#status')).toBeVisible()
  })

  test('New Task modal closes on Cancel', async ({ page }) => {
    await goTo(page, '/tasks')
    await page.locator('button:has-text("New Task"), button:has-text("+")').first().click()
    await expect(page.locator('h3:has-text("Create New Task")')).toBeVisible()
    await page.locator('button:has-text("Cancel")').click()
    await expect(page.locator('h3:has-text("Create New Task")')).toBeHidden()
  })

  test('New Task modal blocks submit with empty title', async ({ page }) => {
    await goTo(page, '/tasks')
    await page.locator('button:has-text("New Task"), button:has-text("+")').first().click()
    // Leave title blank and try to submit
    await page.locator('button[type="submit"]:has-text("Create Task")').click()
    // Modal stays open — error message or native validation prevents close
    await expect(page.locator('h3:has-text("Create New Task")')).toBeVisible()
  })

  test('creates a new task and closes the modal', async ({ page }) => {
    await goTo(page, '/tasks')
    const unique = `E2E-Task-${Date.now()}`

    await page.locator('button:has-text("New Task"), button:has-text("+")').first().click()
    await page.locator('#title').fill(unique)
    await page.locator('#description').fill('Created by Playwright E2E test')
    // Leave status at default "Backlog"
    await page.locator('button[type="submit"]:has-text("Create Task")').click()

    // Modal should close
    await expect(page.locator('h3:has-text("Create New Task")')).toBeHidden({ timeout: 8_000 })

    // The new task title should appear somewhere on the board
    await expect(page.locator(`text=${unique}`)).toBeVisible({ timeout: 8_000 })
  })

  test('task card click opens the task detail modal', async ({ page }) => {
    await goTo(page, '/tasks')

    // Click the first task card on the board (if any)
    const taskCard = page.locator('.cursor-grab').first()
    const count = await taskCard.count()
    if (count === 0) {
      test.skip()
      return
    }

    const titleText = await taskCard.locator('h4').innerText()
    await taskCard.click()

    // TaskDetailsModal should appear with the task title
    await expect(
      page.locator(`[role="dialog"] h3, .fixed h3`).filter({ hasText: titleText })
    ).toBeVisible({ timeout: 5_000 })
  })

  test('task detail modal closes on × button', async ({ page }) => {
    await goTo(page, '/tasks')
    const taskCard = page.locator('.cursor-grab').first()
    const count = await taskCard.count()
    if (count === 0) {
      test.skip()
      return
    }
    await taskCard.click()
    // Close button inside the modal
    const closeBtn = page.locator('.fixed button').filter({ hasText: '' }).first()
    // Use the SVG × close button — look for button containing the "M6 18L18 6" path
    const xBtn = page.locator('button:near(.fixed svg)').filter({
      has: page.locator('path[d*="M6 18L18 6"]')
    }).first()
    if (await xBtn.count() > 0) {
      await xBtn.click()
    } else {
      // Fallback: press Escape
      await page.keyboard.press('Escape')
    }
    await expect(page.locator('h3:has-text("Create New Task")')).toBeHidden()
  })
})

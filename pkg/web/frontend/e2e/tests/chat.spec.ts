/**
 * chat.spec.ts — ChatView E2E tests
 *
 * Covers: page load, sidebar toggle, model selector, message send form,
 * new session, session context menu (rename / archive / delete), slash
 * command autocomplete, file-attach button, Prompt Library button, and
 * Tools-manager popover.
 *
 * Pre-condition: admin auth state restored from e2e/.auth/admin.json
 * (provided by the "setup" project in playwright.config.ts).
 */
import { test, expect } from '@playwright/test'
import { goTo } from '../fixtures/helpers'

test.describe('ChatView — /chat', () => {
  test('loads the chat interface', async ({ page }) => {
    await goTo(page, '/chat')
    // The message textarea is always present
    await expect(
      page.locator('textarea[placeholder="Type a message or / for commands..."]')
    ).toBeVisible()
  })

  test('shows the history sidebar by default', async ({ page }) => {
    await goTo(page, '/chat')
    // Sidebar heading "History" is visible (sidebar open by default)
    await expect(page.locator('h2:has-text("History")')).toBeVisible()
  })

  test('sidebar toggle hides and shows the history panel', async ({ page }) => {
    await goTo(page, '/chat')
    const toggleBtn = page.locator('button[title="Toggle Sidebar"]')
    await expect(toggleBtn).toBeVisible()

    // The sidebar container uses CSS classes w-56/w-64 when open, w-0 when closed.
    // We check the container's class rather than h2 visibility because the h2
    // stays in the DOM (CSS-only transition) and Playwright still sees it as visible.
    const sidebar = page.locator('div.flex-shrink-0.border-r').first()

    // Close the sidebar
    await toggleBtn.click()
    await expect(sidebar).toHaveClass(/w-0/, { timeout: 3_000 })

    // Re-open
    await toggleBtn.click()
    await expect(sidebar).not.toHaveClass(/w-0/, { timeout: 3_000 })
  })

  test('model selector is present', async ({ page }) => {
    await goTo(page, '/chat')
    // <select> in the top bar for model selection
    const modelSelect = page.locator('select').first()
    await expect(modelSelect).toBeVisible()
  })

  test('send button is disabled when textarea is empty', async ({ page }) => {
    await goTo(page, '/chat')
    const sendBtn = page.locator('button[title="Enviar mensaje"]')
    await expect(sendBtn).toBeDisabled()
  })

  test('send button becomes enabled when textarea has text', async ({ page }) => {
    await goTo(page, '/chat')
    const textarea = page.locator('textarea[placeholder="Type a message or / for commands..."]')
    await textarea.fill('Hello')
    const sendBtn = page.locator('button[title="Enviar mensaje"]')
    await expect(sendBtn).toBeEnabled()
  })

  test('empty state message is visible when no messages', async ({ page }) => {
    await goTo(page, '/chat')
    // Start a fresh session so messages panel is empty
    const newChatBtn = page.locator('button[title="Nuevo Chat"]').first()
    if (await newChatBtn.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await newChatBtn.click()
    }
    await expect(
      page.locator('text=Start a conversation').or(page.locator('text=Ask anything or run a task')).first()
    ).toBeVisible({ timeout: 5_000 })
  })

  test('slash-command autocomplete appears on "/" input', async ({ page }) => {
    await goTo(page, '/chat')
    const textarea = page.locator('textarea[placeholder="Type a message or / for commands..."]')
    await textarea.fill('/')
    // Autocomplete list should show at least one item with a /command monospace span
    await expect(page.locator('span.font-mono').first()).toBeVisible({ timeout: 3_000 })
  })

  test('slash-command autocomplete disappears when input is cleared', async ({ page }) => {
    await goTo(page, '/chat')
    const textarea = page.locator('textarea[placeholder="Type a message or / for commands..."]')
    await textarea.fill('/')
    await textarea.fill('')
    await expect(page.locator('span.font-mono').first()).toBeHidden({ timeout: 2_000 })
  })

  test('attach-file button is visible', async ({ page }) => {
    await goTo(page, '/chat')
    await expect(page.locator('button[title="Attach file"]')).toBeVisible()
  })

  test('Prompt Library button is visible', async ({ page }) => {
    await goTo(page, '/chat')
    await expect(page.locator('button[title="Prompt Library"]')).toBeVisible()
  })

  test('Tools manager button opens popover', async ({ page }) => {
    await goTo(page, '/chat')
    const toolsBtn = page.locator('button[title="Manage AI Tools"]')
    await expect(toolsBtn).toBeVisible()
    await toolsBtn.click()
    // Popover heading "AI Tools"
    await expect(page.locator('text=AI Tools')).toBeVisible({ timeout: 3_000 })
  })

  test('Tools manager popover closes on backdrop click', async ({ page }) => {
    await goTo(page, '/chat')
    await page.locator('button[title="Manage AI Tools"]').click()
    await expect(page.locator('text=AI Tools')).toBeVisible({ timeout: 3_000 })
    // Click the backdrop overlay (fixed inset-0 div before the popover)
    await page.mouse.click(10, 10)
    await expect(page.locator('text=AI Tools')).toBeHidden({ timeout: 3_000 })
  })

  test('new chat button (in sidebar) resets the conversation', async ({ page }) => {
    await goTo(page, '/chat')
    const newChatBtn = page.locator('button[title="Nuevo Chat"]').first()
    await expect(newChatBtn).toBeVisible()
    await newChatBtn.click()
    // URL should no longer have a session query param
    await expect(page).not.toHaveURL(/[?&]id=/)
  })

  test('session context menu appears on three-dot button click', async ({ page }) => {
    await goTo(page, '/chat')
    // The context-menu trigger is inside each session item — hover to reveal it
    const sessionBtn = page.locator('button[title="Acciones de sesión"]').first()
    // Some sessions may exist; if not, skip gracefully
    const count = await sessionBtn.count()
    if (count === 0) {
      test.skip()
      return
    }
    await sessionBtn.first().click()
    // The context-menu contains Rename / Archive / Delete
    await expect(page.locator('button:has-text("Rename")')).toBeVisible({ timeout: 3_000 })
    await expect(page.locator('button:has-text("Archive")')).toBeVisible()
    await expect(page.locator('button:has-text("Delete")')).toBeVisible()
  })
})

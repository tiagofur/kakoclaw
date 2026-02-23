/**
 * admin.spec.ts — Admin User Management E2E tests
 *
 * Covers: list users, create user, edit user role/password,
 * block user (with reason), unblock user.
 *
 * User management UI lives at /settings?tab=users (admin only).
 *
 * Key selectors from SettingsView.vue:
 *   - "Add User" button: button with text "Add User"
 *   - Edit button: icon-only, title attribute not set — first button in Actions column
 *   - Block button: title="Bloquear usuario"
 *   - Unblock button: title="Desbloquear usuario"
 *   - User modal username input: input[type="text"] inside the modal (no id/placeholder)
 *   - Save button in modal: text "Save"
 *   - Block confirm button: text "Bloquear"
 *   - Unblock confirm button: text "Desbloquear"
 *
 * Pre-condition: admin auth state restored from e2e/.auth/admin.json.
 */
import { test, expect } from '@playwright/test'
import { goTo } from '../fixtures/helpers'

const USERS_URL = '/settings?tab=users'

// Unique suffix so parallel runs or reruns don't collide
const suffix = () => Date.now().toString(36)

// Helper: open modal and wait for "Create User" title
async function openCreateUserModal(page: any) {
  await page.locator('button').filter({ hasText: 'Add User' }).click()
  await expect(page.locator('text=Create User')).toBeVisible({ timeout: 5_000 })
}

// Helper: fill create user form and save
async function createUser(page: any, username: string, password = 'TestPass123!') {
  await openCreateUserModal(page)
  // Username input: plain input[type="text"] in the modal
  const modal = page.locator('.fixed.inset-0').filter({ hasText: 'Create User' })
  await modal.locator('input[type="text"]').fill(username)
  // Password input
  await modal.locator('input[type="password"]').fill(password)
  // Save
  await modal.locator('button').filter({ hasText: 'Save' }).click()
  await expect(page.locator('text=Create User')).not.toBeVisible({ timeout: 5_000 })
  // Row appears in table
  await expect(page.locator('td').filter({ hasText: username }).first()).toBeVisible({
    timeout: 8_000,
  })
}

test.describe('Admin — User Management (/settings?tab=users)', () => {
  test('Users tab is visible for admin', async ({ page }) => {
    await goTo(page, USERS_URL)
    // The tab button should exist
    await expect(page.locator('button, a').filter({ hasText: /^Users$/ })).toBeVisible({
      timeout: 10_000,
    })
    // The table header "Username" column should be present
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })
  })

  test('Users list shows the admin user', async ({ page }) => {
    await goTo(page, USERS_URL)
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })
    // At least one td cell with the admin username
    await expect(page.locator('td').filter({ hasText: /^admin$/ }).first()).toBeVisible()
  })

  test('Add User button opens the create modal', async ({ page }) => {
    await goTo(page, USERS_URL)
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })
    await openCreateUserModal(page)

    // Modal is open: verify the two inputs exist
    const modal = page.locator('.fixed.inset-0').filter({ hasText: 'Create User' })
    await expect(modal.locator('input[type="text"]')).toBeVisible()
    await expect(modal.locator('input[type="password"]')).toBeVisible()
  })

  test('Create a new user successfully', async ({ page }) => {
    const username = `e2euser_${suffix()}`
    await goTo(page, USERS_URL)
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })
    await createUser(page, username)
  })

  test('Edit button opens modal pre-filled with user data', async ({ page }) => {
    await goTo(page, USERS_URL)
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })

    // The edit button is an icon-only button; it is the first button in each row's Actions cell
    // Click the first one (admin row)
    const firstEditBtn = page.locator('tbody tr').first().locator('button').first()
    await firstEditBtn.click()

    // Modal should appear with Edit User title
    await expect(page.locator('text=Edit User')).toBeVisible({ timeout: 5_000 })

    // Username field should be disabled on edit
    const modal = page.locator('.fixed.inset-0').filter({ hasText: 'Edit User' })
    const usernameInput = modal.locator('input[type="text"]')
    await expect(usernameInput).toBeDisabled()
  })

  test('Block button opens block modal with reason textarea', async ({ page }) => {
    const username = `e2eblock_${suffix()}`
    await goTo(page, USERS_URL)
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })

    await createUser(page, username)

    // Find the row for the new user and click its block button (title="Bloquear usuario")
    const userRow = page.locator('tr').filter({ hasText: username })
    await userRow.locator('button[title="Bloquear usuario"]').click()

    // Block modal should open — "Bloquear Usuario" in the header
    await expect(page.locator('text=Bloquear Usuario')).toBeVisible({ timeout: 5_000 })

    // Reason textarea should be present
    const modal = page.locator('.fixed.inset-0').filter({ hasText: 'Bloquear Usuario' })
    await expect(modal.locator('textarea')).toBeVisible()
  })

  test('Block user requires a reason of at least 10 characters', async ({ page }) => {
    const username = `e2eblk2_${suffix()}`
    await goTo(page, USERS_URL)
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })

    await createUser(page, username)

    // Open block modal
    const userRow = page.locator('tr').filter({ hasText: username })
    await userRow.locator('button[title="Bloquear usuario"]').click()
    const modal = page.locator('.fixed.inset-0').filter({ hasText: 'Bloquear Usuario' })
    await expect(modal.locator('textarea')).toBeVisible({ timeout: 5_000 })

    // Short reason — Bloquear button stays disabled
    await modal.locator('textarea').fill('short')
    const confirmBtn = modal.locator('button').filter({ hasText: 'Bloquear' })
    await expect(confirmBtn).toBeDisabled()

    // Sufficient reason — Bloquear button enables
    await modal.locator('textarea').fill('This is a valid block reason.')
    await expect(confirmBtn).toBeEnabled()
  })

  test('Block then unblock a user', async ({ page }) => {
    const username = `e2eunblk_${suffix()}`
    await goTo(page, USERS_URL)
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })

    await createUser(page, username)

    // Block the user
    let userRow = page.locator('tr').filter({ hasText: username })
    await userRow.locator('button[title="Bloquear usuario"]').click()

    const blockModal = page.locator('.fixed.inset-0').filter({ hasText: 'Bloquear Usuario' })
    await expect(blockModal.locator('textarea')).toBeVisible({ timeout: 5_000 })
    await blockModal.locator('textarea').fill('Automated E2E block test reason.')
    await blockModal.locator('button').filter({ hasText: 'Bloquear' }).click()

    // Wait for modal to close
    await expect(page.locator('text=Bloquear Usuario')).not.toBeVisible({ timeout: 5_000 })

    // Row should now show blocked status (🔴 Bloqueado)
    userRow = page.locator('tr').filter({ hasText: username })
    await expect(userRow.locator('text=Bloqueado')).toBeVisible({ timeout: 8_000 })

    // Unblock: button has title="Desbloquear usuario"
    await userRow.locator('button[title="Desbloquear usuario"]').click()

    // Unblock modal
    const unblockModal = page.locator('.fixed.inset-0').filter({ hasText: 'Desbloquear Usuario' })
    await expect(unblockModal).toBeVisible({ timeout: 5_000 })
    await unblockModal.locator('button').filter({ hasText: 'Desbloquear' }).click()

    // Wait for modal to close
    await expect(page.locator('text=Desbloquear Usuario')).not.toBeVisible({ timeout: 5_000 })

    // Row should show active status (🟢 Activo)
    userRow = page.locator('tr').filter({ hasText: username })
    await expect(userRow.locator('text=Activo')).toBeVisible({ timeout: 8_000 })
  })

  test('Cannot block the last active admin (block button disabled)', async ({ page }) => {
    await goTo(page, USERS_URL)
    await expect(page.locator('text=Username').first()).toBeVisible({ timeout: 10_000 })

    // Find the admin row by locating the td cell whose exact text is "admin",
    // then walking up to the parent tr via Playwright's locator API.
    // We filter tbody rows that contain a td with exact text "admin".
    const adminRow = page.locator('tbody tr').filter({
      has: page.locator('td').filter({ hasText: /^admin$/ })
    }).first()

    // The block button (title="Bloquear usuario") is disabled for the admin user
    // because authStore.user.username === u.username (logged in as admin).
    const blockBtn = adminRow.locator('button[title="Bloquear usuario"]')
    await expect(blockBtn).toBeDisabled()
  })
})

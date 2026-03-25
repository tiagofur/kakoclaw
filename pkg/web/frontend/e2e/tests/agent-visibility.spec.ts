import { test, expect, type Page } from '@playwright/test'
import { goTo } from '../fixtures/helpers'

type MockSocketEvent = { data: string }
type MockSocketHandler = ((event: MockSocketEvent) => void) | null
type MockOpenCloseHandler = ((event: { target: unknown }) => void) | null

type MockApiState = {
  extendedThinking: boolean
  sessions: Array<Record<string, unknown>>
  sessionMessages: Record<string, { messages: Array<Record<string, unknown>> }>
}

async function installWebSocketMock(page: Page) {
  await page.addInitScript(() => {
    class MockWebSocket {
      static CONNECTING = 0
      static OPEN = 1
      static CLOSING = 2
      static CLOSED = 3

      readyState = MockWebSocket.CONNECTING
      url
      onopen: MockOpenCloseHandler = null
      onmessage: MockSocketHandler = null
      onerror: ((event: unknown) => void) | null = null
      onclose: MockOpenCloseHandler = null

      constructor(url: string) {
        this.url = url
        ;(window as any).__mockChatWs.sockets.push(this)
        setTimeout(() => {
          this.readyState = MockWebSocket.OPEN
          this.onopen?.({ target: this })
        }, 0)
      }

      send(data: string) {
        let parsed = data
        try {
          parsed = JSON.parse(data)
        } catch {
          // Keep the raw payload when parsing fails.
        }

        const controller = (window as any).__mockChatWs
        controller.sent.push(parsed)

        const nextBatch = controller.queue.shift() || []
        nextBatch.forEach((event: unknown, index: number) => {
          setTimeout(() => controller.emit(event, this), index * 10)
        })
      }

      close() {
        this.readyState = MockWebSocket.CLOSED
        this.onclose?.({ target: this })
      }

      addEventListener(type: string, handler: (event: unknown) => void) {
        ;(this as any)[`on${type}`] = handler
      }

      removeEventListener(type: string) {
        ;(this as any)[`on${type}`] = null
      }
    }

    ;(window as any).__mockChatWs = {
      sockets: [],
      queue: [],
      sent: [],
      enqueue(events: unknown[]) {
        this.queue.push(events)
      },
      emit(event: unknown, socket?: { onmessage?: (payload: { data: string }) => void; url?: string }) {
        const target = socket || this.sockets.find((candidate: { url?: string }) => String(candidate.url || '').includes('/ws/chat'))
        target?.onmessage?.({ data: JSON.stringify(event) })
      }
    }

    ;(window as any).WebSocket = MockWebSocket
  })
}

async function installApiMocks(page: Page, state: MockApiState) {
  await page.route('**/api/v1/**', async (route) => {
    const request = route.request()
    const url = new URL(request.url())
    const path = url.pathname

    if (path === '/api/v1/chat/sessions' && request.method() === 'GET') {
      await route.fulfill({ json: { sessions: state.sessions } })
      return
    }

    if (path.startsWith('/api/v1/chat/sessions/') && request.method() === 'GET') {
      const sessionId = decodeURIComponent(path.replace('/api/v1/chat/sessions/', ''))
      await route.fulfill({ json: state.sessionMessages[sessionId] || { messages: [] } })
      return
    }

    if (path === '/api/v1/models' && request.method() === 'GET') {
      await route.fulfill({
        json: {
          current_model: 'claude-3-7-sonnet',
          providers: [
            {
              name: 'anthropic',
              enabled: true,
              is_active: true,
              models: [{ id: 'claude-3-7-sonnet', provider: 'anthropic' }]
            }
          ]
        }
      })
      return
    }

    if (path === '/api/v1/me/config' && request.method() === 'GET') {
      await route.fulfill({ json: { config: {} } })
      return
    }

    if (path === '/api/v1/auth/profile' && request.method() === 'GET') {
      await route.fulfill({
        json: {
          username: 'admin',
          email: 'admin@example.com',
          role: 'admin',
          created_at: '2026-03-25T12:00:00.000Z'
        }
      })
      return
    }

    if (path === '/api/v1/user/config' && request.method() === 'GET') {
      await route.fulfill({ json: { extended_thinking: state.extendedThinking } })
      return
    }

    if (path === '/api/v1/user/config' && request.method() === 'PUT') {
      const payload = request.postDataJSON() as { extended_thinking?: boolean }
      state.extendedThinking = !!payload?.extended_thinking
      await route.fulfill({ json: { extended_thinking: state.extendedThinking } })
      return
    }

    await route.fulfill({ json: {} })
  })
}

async function queueChatEvents(page: Page, events: Array<Record<string, unknown>>) {
  await page.evaluate((queuedEvents) => {
    ;(window as any).__mockChatWs.enqueue(queuedEvents)
  }, events)
}

async function sendChatMessage(page: Page, content: string) {
  await page.locator('textarea[placeholder="Type a message or / for commands..."]').fill(content)
  await page.locator('button[title="Enviar mensaje"]').click()
}

function createHistoricalSessionMessages() {
  const toolCalls = Array.from({ length: 5 }, (_, index) => ({
    id: `tool-${index + 1}`,
    function: {
      name: `history_tool_${index + 1}`,
      arguments: JSON.stringify({ query: `input-${index + 1}` })
    }
  }))

  const toolResults = toolCalls.map((toolCall, index) => ({
    role: 'tool',
    tool_call_id: toolCall.id,
    content: `result-${index + 1}`
  }))

  return {
    messages: [
      {
        id: 'msg-history-1',
        role: 'assistant',
        content: 'Historical answer',
        created_at: '2026-03-25T12:00:00.000Z',
        metadata: JSON.stringify({
          messages: [
            {
              role: 'assistant',
              tool_calls: toolCalls
            },
            ...toolResults
          ]
        })
      }
    ]
  }
}

test.describe('Agent visibility realtime', () => {
  test.beforeEach(async ({ page }) => {
    await installWebSocketMock(page)
  })

  test('historical messages render tool calls collapsed by default', async ({ page }) => {
    const state: MockApiState = {
      extendedThinking: false,
      sessions: [{ session_id: 'history-session', title: 'History session' }],
      sessionMessages: {
        'history-session': createHistoricalSessionMessages()
      }
    }

    await installApiMocks(page, state)
    await goTo(page, '/chat?id=history-session')

    await expect(page.getByText('history_tool_1')).toBeVisible()
    await expect(page.getByText('history_tool_5')).toBeVisible()
    await expect(page.getByText('Arguments')).toHaveCount(0)
    await expect(page.getByText('Result')).toHaveCount(0)

    await page.locator('button').filter({ hasText: 'history_tool_1' }).click()

    await expect(page.getByText('Arguments')).toBeVisible()
    await expect(page.getByText('Result')).toBeVisible()
    await expect(page.getByText('result-1')).toBeVisible()
  })

  test('extended thinking toggle persists after reload and enables thinking blocks', async ({ page }) => {
    const state: MockApiState = {
      extendedThinking: false,
      sessions: [],
      sessionMessages: {}
    }

    await installApiMocks(page, state)
    await goTo(page, '/settings?tab=profile')

    await expect(page.getByText('Extended Thinking (Claude only)')).toBeVisible()

    const toggle = page.locator('input[type="checkbox"]').first()
    const toggleLabel = page.locator('label').filter({ has: toggle }).first()

    await expect(toggle).not.toBeChecked()
    await toggleLabel.click()
    await expect(toggle).toBeChecked()

    await page.reload()
    await expect(page.getByText('Extended Thinking (Claude only)')).toBeVisible()
    await expect(toggle).toBeChecked()

    await goTo(page, '/chat')
    await queueChatEvents(page, [
      { type: 'stream_start' },
      { type: 'thinking_delta', content: 'Visible thinking content' },
      { type: 'stream', content: 'Assistant reply' }
    ])

    await sendChatMessage(page, 'Show your reasoning')

    await expect(page.getByText('Visible thinking content')).toBeVisible()
  })

  test('team activity panel groups in-flight tool calls by agent', async ({ page }) => {
    const state: MockApiState = {
      extendedThinking: false,
      sessions: [],
      sessionMessages: {}
    }

    await installApiMocks(page, state)
    await goTo(page, '/chat')

    await queueChatEvents(page, [
      { type: 'stream_start' },
      { type: 'agent_status', agent: 'orchestrator', status: 'delegating', specialist_name: 'developer', reason: 'Need code review' },
      { type: 'tool_call', name: 'code_analysis', args: { file: 'chatStore.js' }, status: 'started' },
      { type: 'agent_status', agent: 'developer', status: 'requesting_help', specialist_name: 'researcher', reason: 'Need docs lookup' },
      { type: 'tool_call', name: 'doc_search', args: { query: 'thinking blocks' }, status: 'started' }
    ])

    await sendChatMessage(page, 'Coordinate a multi-agent answer')

    await expect(page.getByRole('button', { name: /Team Activity/i })).toBeVisible()
    await page.getByRole('button', { name: /Team Activity/i }).click()

    await expect(page.getByText('developer')).toBeVisible()
    await expect(page.getByText('researcher')).toBeVisible()
    await expect(page.getByText('code_analysis')).toBeVisible()
    await expect(page.getByText('doc_search')).toBeVisible()
    await expect(page.getByText('executing…')).toHaveCount(2)
  })

  test('default users do not see thinking blocks until extended thinking is enabled', async ({ page }) => {
    const state: MockApiState = {
      extendedThinking: false,
      sessions: [],
      sessionMessages: {}
    }

    await installApiMocks(page, state)
    await goTo(page, '/chat')

    await queueChatEvents(page, [
      { type: 'stream_start' },
      { type: 'stream', content: 'Plain answer' },
      { type: 'stream_end', content: 'Plain answer', agents: [] }
    ])

    await sendChatMessage(page, 'Answer without extended thinking')

    await expect(page.getByText('Hidden reasoning content')).toHaveCount(0)

    await goTo(page, '/settings?tab=profile')
    const toggle = page.locator('input[type="checkbox"]').first()
    const toggleLabel = page.locator('label').filter({ has: toggle }).first()
    await toggleLabel.click()
    await expect(toggle).toBeChecked()

    await goTo(page, '/chat')
    await queueChatEvents(page, [
      { type: 'stream_start' },
      { type: 'thinking_delta', content: 'Hidden reasoning content' },
      { type: 'stream', content: 'Answer with reasoning' }
    ])

    await sendChatMessage(page, 'Answer with extended thinking')

    await expect(page.getByText('Hidden reasoning content')).toBeVisible()
  })
})

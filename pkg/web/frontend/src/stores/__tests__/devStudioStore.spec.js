import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('../../services/api', () => ({
  default: {
	get: vi.fn(),
	post: vi.fn()
  }
}))

import { useDevStudioStore } from '../devStudioStore'
import api from '../../services/api'

class MockWebSocket {
  static OPEN = 1

  constructor(url) {
    this.url = url
    this.readyState = MockWebSocket.OPEN
    this.onopen = null
    this.onmessage = null
    this.onerror = null
    this.onclose = null
  }

  send() {}
  close() {}
}

describe('devStudioStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    api.get.mockReset()
    api.post.mockReset()
    localStorage.clear()
    localStorage.setItem('auth.token', 'test-token')
    global.WebSocket = MockWebSocket
    global.fetch = vi.fn()
  })

  it('increments token state when a result event arrives', () => {
    const store = useDevStudioStore()
    store.sessionTokens = 500

    store.processTerminalEvent({
      type: 'result',
      tokens_used: 300,
      cost_usd: 0.02,
      num_turns: 300
    })

    expect(store.sessionTokens).toBe(800)
    expect(store.sessionCostUsd).toBe(0.02)
    expect(store.numTurns).toBe(300)
  })

  it('resets counters and updates sessionId on session_reset event', () => {
    const store = useDevStudioStore()
    store.sessionTokens = 1050
    store.sessionCostUsd = 0.11
    store.numTurns = 42
    store.sessionId = 'old-session'

    store.processTerminalEvent({
      type: 'session_reset',
      new_session_id: 'abc-123',
      message: 'Session auto-reset: token limit reached'
    })

    expect(store.sessionTokens).toBe(0)
    expect(store.sessionCostUsd).toBe(0)
    expect(store.numTurns).toBe(0)
    expect(store.sessionId).toBe('abc-123')
    expect(store.terminalHistory.at(-1)).toMatchObject({
      type: 'session_reset',
      message: 'Session auto-reset: token limit reached'
    })
  })

  it('loads token state from stats endpoint when panel opens', async () => {
    api.get
      .mockResolvedValueOnce({ data: { status: 'idle', project_dir: '' } })
      .mockResolvedValueOnce({
        data: {
          session_id: 'session-a',
          tokens_used: 750,
          cost_usd: 0.015,
          token_limit: 1000,
          num_turns: 9
        }
      })

    const store = useDevStudioStore()
    await store.checkStatus()

    expect(api.get).toHaveBeenCalledWith('/dev/bridge/status')
    expect(api.get).toHaveBeenCalledWith('/dev/session/stats')
    expect(store.sessionId).toBe('session-a')
    expect(store.sessionTokens).toBe(750)
    expect(store.sessionCostUsd).toBe(0.015)
    expect(store.tokenLimit).toBe(1000)
    expect(store.numTurns).toBe(9)
  })
})

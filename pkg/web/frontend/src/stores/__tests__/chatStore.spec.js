import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useChatStore } from '../chatStore'

describe('chatStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('addToolCall() assigns agentName from currentAgent', () => {
    const store = useChatStore()
    const messageId = store.startStreamingMessage()

    store.currentAgent = 'developer'
    store.addToolCall({ name: 'search', status: 'started' })

    const toolCall = store.messages.find((msg) => msg.id === messageId).toolCalls[0]

    expect(toolCall.agentName).toBe('developer')
  })

  it("addToolCall() falls back to 'main' when currentAgent is null", () => {
    const store = useChatStore()
    const messageId = store.startStreamingMessage()

    store.currentAgent = null
    store.addToolCall({ name: 'search', status: 'started' })

    const toolCall = store.messages.find((msg) => msg.id === messageId).toolCalls[0]

    expect(toolCall.agentName).toBe('main')
  })

  it("addToolCall() sets expanded=true when streaming and status is 'started'", () => {
    const store = useChatStore()
    const messageId = store.startStreamingMessage()

    store.addToolCall({ name: 'search', status: 'started' })

    const toolCall = store.messages.find((msg) => msg.id === messageId).toolCalls[0]

    expect(toolCall.expanded).toBe(true)
  })

  it("addToolCall() sets expanded=false when not streaming or status is not 'started'", () => {
    const store = useChatStore()
    const messageId = store.startStreamingMessage()
    const message = store.messages.find((msg) => msg.id === messageId)

    message.streaming = false
    store.addToolCall({ name: 'search', status: 'finished' })

    const toolCall = store.messages.find((msg) => msg.id === messageId).toolCalls[0]

    expect(toolCall.expanded).toBe(false)
  })

  it('updateCurrentAgent() updates currentAgent', () => {
    const store = useChatStore()

    store.updateCurrentAgent('developer')

    expect(store.currentAgent).toBe('developer')
  })

  it('appendThinkingDelta() accumulates content in the active thinking block', () => {
    const store = useChatStore()
    const messageId = store.startStreamingMessage()

    store.appendThinkingDelta('thinking content 1')
    store.appendThinkingDelta('thinking content 2')

    const message = store.messages.find((msg) => msg.id === messageId)

    expect(message.thinkingBlocks).toHaveLength(1)
    expect(message.thinkingBlocks[0].content).toBe('thinking content 1thinking content 2')
    expect(message.thinkingBlocks[0].expanded).toBe(true)
  })

  it('endStreamingMessage() collapses thinking blocks that were not manually toggled', () => {
    const store = useChatStore()
    const messageId = store.startStreamingMessage()
    const message = store.messages.find((msg) => msg.id === messageId)

    message.thinkingBlocks = [
      {
        id: 'block-1',
        content: 'developer reasoning',
        expanded: true,
        finalized: false,
        manuallyToggled: true
      },
      {
        id: 'block-2',
        content: 'main reasoning',
        expanded: true,
        finalized: false,
        manuallyToggled: false
      }
    ]

    store.endStreamingMessage('done')

    expect(message.thinkingBlocks[0].expanded).toBe(true)
    expect(message.thinkingBlocks[0].finalized).toBe(true)
    expect(message.thinkingBlocks[1].expanded).toBe(false)
    expect(message.thinkingBlocks[1].finalized).toBe(true)
  })
})

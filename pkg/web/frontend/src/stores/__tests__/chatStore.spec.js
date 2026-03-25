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
})

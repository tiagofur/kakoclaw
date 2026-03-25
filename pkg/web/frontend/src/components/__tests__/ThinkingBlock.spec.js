import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ThinkingBlock from '../Chat/ThinkingBlock.vue'

function mountThinkingBlock(overrides = {}) {
  const thinkingBlock = {
    id: 'thinking-1',
    content: 'internal reasoning',
    expanded: true,
    finalized: false,
    manuallyToggled: false,
    ...overrides.thinkingBlock
  }

  return mount(ThinkingBlock, {
    props: {
      thinkingBlock,
      isStreaming: false,
      ...overrides.props
    }
  })
}

describe('ThinkingBlock', () => {
  it('auto-expands while the stream is active', () => {
    const wrapper = mountThinkingBlock({
      thinkingBlock: { expanded: true, manuallyToggled: false },
      props: { isStreaming: true }
    })

    expect(wrapper.vm.isExpanded).toBe(true)
    expect(wrapper.text()).toContain('internal reasoning')
  })

  it('stays collapsed after stream end when not manually toggled', () => {
    const wrapper = mountThinkingBlock({
      thinkingBlock: { expanded: false, manuallyToggled: false },
      props: { isStreaming: false }
    })

    expect(wrapper.vm.isExpanded).toBe(false)
    expect(wrapper.text()).not.toContain('internal reasoning')
  })

  it('marks manual toggles and preserves the flag across subsequent clicks', async () => {
    const wrapper = mountThinkingBlock({
      thinkingBlock: { expanded: false, manuallyToggled: false },
      props: { isStreaming: false }
    })

    const header = wrapper.find('button')

    await header.trigger('click')
    expect(wrapper.props('thinkingBlock').manuallyToggled).toBe(true)
    expect(wrapper.vm.isExpanded).toBe(true)

    await header.trigger('click')
    expect(wrapper.props('thinkingBlock').manuallyToggled).toBe(true)
    expect(wrapper.vm.isExpanded).toBe(false)
  })

  it('renders the agent badge and brain icon in the header', () => {
    const wrapper = mountThinkingBlock({
      thinkingBlock: { agentName: 'developer' }
    })

    expect(wrapper.text()).toContain('developer')
    expect(wrapper.text()).toContain('🧠')
  })

  it('applies the expected visual styles', () => {
    const wrapper = mountThinkingBlock({
      thinkingBlock: { expanded: true }
    })

    const content = wrapper.find('pre')

    expect(wrapper.classes()).toContain('glass-panel')
    expect(wrapper.classes()).toContain('border-l-4')
    expect(wrapper.classes()).toContain('border-l-blue-500/50')
    expect(content.classes()).toContain('font-mono')
    expect(content.classes()).toContain('text-makoclaw-text-secondary')
  })
})

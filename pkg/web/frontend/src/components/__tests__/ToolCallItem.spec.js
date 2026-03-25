import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ToolCallItem from '../ToolCallItem.vue'

function mountToolCallItem(overrides = {}) {
  const toolCall = {
    name: 'search',
    status: 'started',
    args: { query: 'agent visibility' },
    result: null,
    timestamp: '2026-03-25T12:00:00.000Z',
    ...overrides.toolCall
  }

  const msg = {
    streaming: true,
    ...overrides.msg
  }

  return mount(ToolCallItem, {
    props: {
      toolCall,
      msg
    }
  })
}

function getStatusBadge(wrapper) {
  return wrapper.findAll('span').find((node) => ['executing…', 'done', 'error'].includes(node.text()))
}

describe('ToolCallItem', () => {
  it("renders a started tool call expanded while streaming", () => {
    const wrapper = mountToolCallItem()

    expect(wrapper.vm.isExpanded).toBe(true)
    expect(wrapper.text()).toContain('Arguments')
    expect(wrapper.find('pre').text()).toContain('agent visibility')
  })

  it("renders a completed tool call collapsed and stays collapsed after status change", async () => {
    const wrapper = mountToolCallItem({
      toolCall: { status: 'finished' },
      msg: { streaming: true }
    })

    expect(wrapper.vm.isExpanded).toBe(false)
    expect(wrapper.text()).not.toContain('Arguments')

    await wrapper.setProps({
      toolCall: {
        ...wrapper.props('toolCall'),
        status: 'finished'
      }
    })

    expect(wrapper.vm.isExpanded).toBe(false)
    expect(wrapper.find('pre').exists()).toBe(false)
  })

  it('toggles manual expansion on header click', async () => {
    const wrapper = mountToolCallItem({
      toolCall: { status: 'finished' },
      msg: { streaming: false }
    })

    const header = wrapper.find('button')

    await header.trigger('click')
    expect(wrapper.vm.localExpanded).toBe(true)
    expect(wrapper.text()).toContain('Arguments')

    await header.trigger('click')
    expect(wrapper.vm.localExpanded).toBe(false)
    expect(wrapper.text()).not.toContain('Arguments')
  })

  it.each([
    ['started', 'executing…', 'text-makoclaw-warning'],
    ['finished', 'done', 'text-makoclaw-success'],
    ['error', 'error', 'text-makoclaw-error']
  ])('renders semantic badge for %s status', (status, label, colorClass) => {
    const wrapper = mountToolCallItem({
      toolCall: { status }
    })

    const badge = getStatusBadge(wrapper)

    expect(badge?.exists()).toBe(true)
    expect(badge?.text()).toBe(label)
    expect(badge?.classes()).toContain(colorClass)
  })
})

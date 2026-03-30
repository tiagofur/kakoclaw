import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
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

  return mount(ToolCallItem, {
    props: { toolCall },
    global: {
      plugins: [createPinia()]
    }
  })
}

function getStatusBadge(wrapper) {
  return wrapper.findAll('span').find((node) => ['executing…', 'done', 'error'].includes(node.text()))
}

describe('ToolCallItem', () => {
  it('starts collapsed regardless of status or streaming', () => {
    const wrapper = mountToolCallItem()

    expect(wrapper.vm.isExpanded).toBe(false)
    expect(wrapper.find('pre').exists()).toBe(false)
  })

  it('stays collapsed when status is finished', () => {
    const wrapper = mountToolCallItem({ toolCall: { status: 'finished' } })

    expect(wrapper.vm.isExpanded).toBe(false)
    expect(wrapper.text()).not.toContain('Arguments')
  })

  it('toggles manual expansion on header click', async () => {
    const wrapper = mountToolCallItem({ toolCall: { status: 'finished' } })

    const header = wrapper.find('button')

    await header.trigger('click')
    expect(wrapper.vm.localExpanded).toBe(true)
    expect(wrapper.text()).toContain('Arguments')

    await header.trigger('click')
    expect(wrapper.vm.localExpanded).toBe(false)
    expect(wrapper.text()).not.toContain('Arguments')
  })

  it('can be opened while executing and stays open until user closes it', async () => {
    const wrapper = mountToolCallItem({ toolCall: { status: 'started' } })

    expect(wrapper.vm.isExpanded).toBe(false)

    await wrapper.find('button').trigger('click')
    expect(wrapper.vm.isExpanded).toBe(true)

    await wrapper.find('button').trigger('click')
    expect(wrapper.vm.isExpanded).toBe(false)
  })

  it.each([
    ['started', 'executing…', 'text-makoclaw-warning'],
    ['finished', 'done', 'text-makoclaw-success'],
    ['error', 'error', 'text-makoclaw-error']
  ])('renders semantic badge for %s status', (status, label, colorClass) => {
    const wrapper = mountToolCallItem({ toolCall: { status } })

    const badge = getStatusBadge(wrapper)

    expect(badge?.exists()).toBe(true)
    expect(badge?.text()).toBe(label)
    expect(badge?.classes()).toContain(colorClass)
  })
})

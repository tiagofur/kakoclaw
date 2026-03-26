import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import SpecialistSegment from '../Chat/SpecialistSegment.vue'

function mountSpecialistSegment(overrides = {}) {
  const segment = {
    agent: 'developer',
    content: 'Line one\nLine two\nLine three',
    status: 'complete',
    confidence: 0.93,
    timestamp: '2026-03-25T12:00:00.000Z',
    ...overrides.segment
  }

  return mount(SpecialistSegment, {
    props: {
      segment,
      isStreaming: false,
      ...overrides.props
    },
    global: {
      stubs: {
        MarkdownRenderer: {
          props: ['content'],
          template: '<div class="markdown-stub">{{ content }}</div>'
        }
      }
    }
  })
}

describe('SpecialistSegment', () => {
  it('renders collapsed by default with a two-line preview', () => {
    const wrapper = mountSpecialistSegment()

    expect(wrapper.vm.isExpanded).toBe(false)
    expect(wrapper.find('[data-testid="specialist-content"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="specialist-preview"]').text()).toContain('Line one')
    expect(wrapper.find('[data-testid="specialist-preview"]').text()).toContain('Line two')
    expect(wrapper.find('[data-testid="specialist-preview"]').text()).not.toContain('Line three')
  })

  it('toggles expansion on header click', async () => {
    const wrapper = mountSpecialistSegment()
    const header = wrapper.find('[data-testid="specialist-header"]')

    await header.trigger('click')
    expect(wrapper.vm.localExpanded).toBe(true)
    expect(wrapper.find('[data-testid="specialist-content"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Line three')

    await header.trigger('click')
    expect(wrapper.vm.localExpanded).toBe(false)
    expect(wrapper.find('[data-testid="specialist-content"]').exists()).toBe(false)
  })

  it.each([
    ['working', 'text-makoclaw-warning'],
    ['complete', 'text-makoclaw-success'],
    ['error', 'text-makoclaw-error']
  ])('renders semantic badge styling for %s status', (status, className) => {
    const wrapper = mountSpecialistSegment({
      segment: { status }
    })

    expect(wrapper.find('[data-testid="specialist-status-badge"]').classes()).toContain(className)
  })
})

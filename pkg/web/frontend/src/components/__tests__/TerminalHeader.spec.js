import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import TerminalHeader from '../DevStudio/TerminalHeader.vue'
import { useDevStudioStore } from '../../stores/devStudioStore'

function mountHeader() {
  return mount(TerminalHeader, {
    global: {
      plugins: [useDevStudioStore().$pinia]
    }
  })
}

describe('TerminalHeader', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders live token usage badge text', () => {
    const store = useDevStudioStore()
    store.sessionTokens = 750
    store.tokenLimit = 1000

    const wrapper = mountHeader()
    const badge = wrapper.get('[data-testid="token-badge"]')

    expect(badge.text()).toContain('750 / 1000 tokens')
    expect(badge.classes()).toContain('token-badge-default')
  })

  it('applies warning class at 90 percent usage', () => {
    const store = useDevStudioStore()
    store.sessionTokens = 900
    store.tokenLimit = 1000

    const wrapper = mountHeader()
    expect(wrapper.get('[data-testid="token-badge"]').classes()).toContain('token-badge-warning')
  })

  it('applies danger class at 100 percent usage', () => {
    const store = useDevStudioStore()
    store.sessionTokens = 1000
    store.tokenLimit = 1000

    const wrapper = mountHeader()
    expect(wrapper.get('[data-testid="token-badge"]').classes()).toContain('token-badge-danger')
  })
})

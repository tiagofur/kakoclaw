import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import SynthesisIndicator from '../Chat/SynthesisIndicator.vue'

describe('SynthesisIndicator', () => {
  it('shows the indicator when synthesizing', () => {
    const wrapper = mount(SynthesisIndicator, {
      props: {
        isSynthesizing: true
      }
    })

    expect(wrapper.find('[data-testid="synthesis-indicator"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Orchestrator synthesizing')
  })

  it('hides the indicator when not synthesizing', () => {
    const wrapper = mount(SynthesisIndicator, {
      props: {
        isSynthesizing: false
      }
    })

    expect(wrapper.find('[data-testid="synthesis-indicator"]').exists()).toBe(false)
  })
})

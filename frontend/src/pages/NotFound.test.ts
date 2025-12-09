import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import NotFound from './NotFound.vue'

describe('NotFound page', () => {
  function mountNotFound() {
    return mount(NotFound, {
      global: {
        stubs: {
          RouterLink: { template: '<a :href="String(to)"><slot /></a>', props: ['to'] },
          Button: { template: '<button><slot /></button>' },
        },
      },
    })
  }

  it('renders 404 heading', () => {
    const wrapper = mountNotFound()
    expect(wrapper.html()).toContain('404')
  })

  it('renders "Page not found" text', () => {
    const wrapper = mountNotFound()
    expect(wrapper.html()).toContain('Page not found')
  })

  it('renders a link to home', () => {
    const wrapper = mountNotFound()
    expect(wrapper.html()).toContain('/')
    expect(wrapper.html()).toContain('Go home')
  })
})

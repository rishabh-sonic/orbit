import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { makeUser } from '@/test/factories'
import Home from './Home.vue'

function mountHome() {
  return mount(Home, {
    global: {
      stubs: {
        PostFeed: { template: '<div data-testid="post-feed" />' },
        Button: { template: '<button><slot /></button>' },
        RouterLink: { template: '<a :href="String(to)"><slot /></a>', props: ['to'] },
      },
    },
  })
}

describe('Home page', () => {
  it('renders page heading', () => {
    setActivePinia(createPinia())
    const wrapper = mountHome()
    expect(wrapper.html()).toContain('Feed')
  })

  it('renders PostFeed component', () => {
    setActivePinia(createPinia())
    const wrapper = mountHome()
    expect(wrapper.find('[data-testid="post-feed"]').exists()).toBe(true)
  })

  it('shows New Post button for logged-in users', () => {
    setActivePinia(createPinia())
    const auth = useAuthStore()
    auth.setToken('test-token')
    auth.setUser(makeUser({ id: '1', username: 'alice', role: 'USER' }))
    const wrapper = mountHome()
    expect(wrapper.html()).toContain('/posts/new')
  })

  it('hides New Post button for guests', () => {
    setActivePinia(createPinia())
    const wrapper = mountHome()
    // Not logged in — no token
    expect(wrapper.find('a[href="/posts/new"]').exists()).toBe(false)
  })
})

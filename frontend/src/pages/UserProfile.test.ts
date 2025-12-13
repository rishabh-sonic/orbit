import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { makeUser } from '@/test/factories'
import UserProfile from './UserProfile.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [{ path: '/users/:id', component: UserProfile }],
})

const mockUser = {
  id: 'user-1',
  username: 'alice',
  avatar: null,
  introduction: 'Hello there!',
  follower_count: 10,
  following_count: 5,
  post_count: 20,
  is_following: false,
  created_at: new Date().toISOString(),
}

function mountProfile() {
  return mount(UserProfile, {
    global: {
      plugins: [router],
      stubs: {
        PostFeed: { template: '<div data-testid="post-feed" />' },
        UserAvatar: { template: '<div class="avatar" />' },
        Skeleton: { template: '<div class="skeleton" />' },
        Tabs: { template: '<div><slot /></div>' },
        TabsList: { template: '<div><slot /></div>' },
        TabsTrigger: { template: '<button><slot /></button>' },
        TabsContent: { template: '<div><slot /></div>' },
        Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
      },
    },
  })
}

describe('UserProfile page', () => {
  beforeEach(async () => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
    await router.push('/users/alice')
  })

  it('shows loading skeleton while fetching', () => {
    vi.mocked(fetch).mockImplementation(() => new Promise(() => {}))
    const wrapper = mountProfile()
    expect(wrapper.find('.skeleton').exists()).toBe(true)
  })

  function mockProfileFetch() {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockUser }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
      )
  }

  it('renders username after load', async () => {
    mockProfileFetch()
    const wrapper = mountProfile()
    await flushPromises()
    expect(wrapper.html()).toContain('alice')
  })

  it('renders follower and following counts', async () => {
    mockProfileFetch()
    const wrapper = mountProfile()
    await flushPromises()
    expect(wrapper.html()).toContain('10')
    expect(wrapper.html()).toContain('5')
  })

  it('renders intro text', async () => {
    mockProfileFetch()
    const wrapper = mountProfile()
    await flushPromises()
    expect(wrapper.html()).toContain('Hello there!')
  })

  it('shows Follow button for other users when authenticated', async () => {
    const auth = useAuthStore()
    auth.setToken('test-token')
    auth.setUser(makeUser({ id: 'other-id', username: 'bob', role: 'USER' }))
    mockProfileFetch()
    const wrapper = mountProfile()
    await flushPromises()
    const buttons = wrapper.findAll('button')
    const followBtn = buttons.find(b => b.text().toLowerCase().includes('follow'))
    expect(followBtn).toBeDefined()
  })

  it('hides Follow button on own profile', async () => {
    const auth = useAuthStore()
    auth.setToken('test-token')
    auth.setUser(makeUser({ id: 'user-1', username: 'alice', role: 'USER' }))
    mockProfileFetch()
    const wrapper = mountProfile()
    await flushPromises()
    const buttons = wrapper.findAll('button')
    const followBtn = buttons.find(b => b.text().toLowerCase() === 'follow')
    expect(followBtn).toBeUndefined()
  })

  it('renders PostFeed component', async () => {
    mockProfileFetch()
    const wrapper = mountProfile()
    await flushPromises()
    expect(wrapper.find('[data-testid="post-feed"]').exists()).toBe(true)
  })
})

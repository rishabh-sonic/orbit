import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import Search from '@/pages/Search.vue'

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/search', component: Search },
      { path: '/posts/:id', component: { template: '<div />' } },
      { path: '/users/:id', component: { template: '<div />' } },
    ],
  })
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(fetch).mockReset()
})

describe('Search page', () => {
  it('renders a search input', () => {
    const wrapper = mount(Search, {
      global: { plugins: [makeRouter()] },
    })
    expect(wrapper.find('input').exists()).toBe(true)
  })

  it('shows tabs for Posts and Users after searching', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: { posts: [], users: [] } }), { status: 200 })),
    )
    const router = makeRouter()
    await router.push({ path: '/search', query: { q: 'vue' } })
    await router.isReady()

    const wrapper = mount(Search, { global: { plugins: [router] } })
    await flushPromises()

    expect(wrapper.text()).toContain('Posts')
    expect(wrapper.text()).toContain('Users')
  })

  it('calls search API when form is submitted', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: { posts: [], users: [] } }), { status: 200 })),
    )
    const wrapper = mount(Search, { global: { plugins: [makeRouter()] } })
    await wrapper.find('input').setValue('hello')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('hello'),
      expect.anything(),
    )
  })

  it('renders post results', async () => {
    const post = {
      id: 'p1',
      title: 'Search Result Post',
      content: 'Content here',
      author: { id: 'u1', username: 'alice', avatar: null },
      views: 10,
      comment_count: 2,
      pinned_at: null,
      status: 'open',
      created_at: new Date().toISOString(),
    }
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify({ data: { posts: [post], users: [] } }), { status: 200 }),
      ),
    )
    const router = makeRouter()
    await router.push({ path: '/search', query: { q: 'result' } })
    const wrapper = mount(Search, { global: { plugins: [router] } })
    await flushPromises()
    expect(wrapper.text()).toContain('Search Result Post')
  })

  it('shows empty state when no results', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: { posts: [], users: [] } }), { status: 200 })),
    )
    const router = makeRouter()
    await router.push({ path: '/search', query: { q: 'noresults' } })
    const wrapper = mount(Search, { global: { plugins: [router] } })
    await flushPromises()
    // No posts found message appears in posts tab
    expect(wrapper.text()).toMatch(/no posts found|0\)/i)
  })
})

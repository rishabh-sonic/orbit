import { describe, it, expect, vi, beforeEach } from 'vitest'
import { nextTick } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import PostFeed from './PostFeed.vue'
import { makePost } from '@/test/factories'

const router = createRouter({
  history: createWebHistory(),
  routes: [{ path: '/', component: { template: '<div />' } }],
})

const mockPosts = Array.from({ length: 5 }, (_, i) =>
  makePost({ id: `post-${i}`, title: `Post ${i}`, content: `Content ${i}`, comment_count: i, views: i * 10 })
)

function mountFeed(props: { endpoint?: string; userId?: string } = {}) {
  return mount(PostFeed, {
    props,
    global: {
      plugins: [router],
      stubs: {
        PostCard: {
          template: '<div class="post-card" :data-id="post.id">{{ post.title }}</div>',
          props: ['post'],
        },
        Skeleton: { template: '<div class="skeleton" />' },
        Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
      },
    },
  })
}

describe('PostFeed component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
  })

  it('fetches posts on mount', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPosts }), { status: 200 }))
    )
    mountFeed()
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/posts'),
      expect.anything()
    )
  })

  it('renders post cards', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPosts }), { status: 200 }))
    )
    const wrapper = mountFeed()
    await flushPromises()
    expect(wrapper.findAll('.post-card').length).toBe(5)
  })

  it('shows empty state when no posts', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    const wrapper = mountFeed()
    await flushPromises()
    expect(wrapper.html().toLowerCase()).toContain('no posts')
  })

  it('shows loading skeleton on initial load', async () => {
    vi.mocked(fetch).mockImplementation(() => new Promise(() => {}))
    const wrapper = mountFeed()
    await nextTick()
    expect(wrapper.find('.skeleton').exists()).toBe(true)
  })

  it('uses custom endpoint when provided', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    mountFeed({ endpoint: '/posts/featured' })
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/posts/featured'),
      expect.anything()
    )
  })

  it('appends userId param when provided', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    mountFeed({ userId: 'user-1' })
    await flushPromises()
    const url = String(vi.mocked(fetch).mock.calls[0][0])
    expect(url).toContain('user_id=user-1')
  })

  it('shows Load More button when there are more posts', async () => {
    // Return exactly `limit` (20) posts to trigger hasMore
    const manyPosts = Array.from({ length: 20 }, (_, i) => ({ ...mockPosts[0], id: `p${i}` }))
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: manyPosts }), { status: 200 }))
    )
    const wrapper = mountFeed()
    await flushPromises()
    const buttons = wrapper.findAll('button')
    const loadMoreBtn = buttons.find(b => b.text().toLowerCase().includes('more'))
    expect(loadMoreBtn).toBeDefined()
  })

  it('hides Load More when fewer than limit posts returned', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPosts }), { status: 200 }))
    )
    const wrapper = mountFeed()
    await flushPromises()
    // 5 posts < limit of 20 — no more button
    const buttons = wrapper.findAll('button')
    const loadMoreBtn = buttons.find(b => b.text().toLowerCase().includes('more'))
    expect(loadMoreBtn).toBeUndefined()
  })
})

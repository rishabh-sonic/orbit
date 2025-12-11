import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { makeUser } from '@/test/factories'
import PostDetail from './PostDetail.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/posts/:id', component: PostDetail },
    { path: '/', component: { template: '<div>home</div>' } },
  ],
})

const mockPost = {
  id: 'post-1',
  title: 'Test Post Title',
  content: 'Post body content',
  author: { id: 'user-1', username: 'alice', avatar: null },
  comment_count: 5,
  view_count: 100,
  pinned_at: null,
  closed_at: null,
  deleted_at: null,
  created_at: new Date().toISOString(),
  subscribed: false,
}

function mountPostDetail(_postId = 'post-1') {
  return mount(PostDetail, {
    global: {
      plugins: [router],
      stubs: {
        CommentList: { template: '<div data-testid="comment-list" />' },
        UserAvatar: { template: '<div class="avatar" />' },
        Skeleton: { template: '<div class="skeleton" />' },
        Badge: { template: '<span class="badge"><slot /></span>' },
      },
    },
  })
}

describe('PostDetail page', () => {
  beforeEach(async () => {
    vi.mocked(fetch).mockClear()
    setActivePinia(createPinia())
    await router.push('/posts/post-1')
  })

  it('shows loading skeleton while fetching', () => {
    vi.mocked(fetch).mockImplementation(() => new Promise(() => {}))
    const wrapper = mountPostDetail()
    expect(wrapper.find('.skeleton').exists()).toBe(true)
  })

  it('renders post title and content after load', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPost }), { status: 200 }))
    )
    const wrapper = mountPostDetail()
    await flushPromises()
    expect(wrapper.html()).toContain('Test Post Title')
    expect(wrapper.html()).toContain('Post body content')
  })

  it('renders author name', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPost }), { status: 200 }))
    )
    const wrapper = mountPostDetail()
    await flushPromises()
    expect(wrapper.html()).toContain('alice')
  })

  it('shows "Pinned" badge when post is pinned', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({ data: { ...mockPost, pinned_at: new Date().toISOString() } }),
          { status: 200 }
        )
      )
    )
    const wrapper = mountPostDetail()
    await flushPromises()
    expect(wrapper.html().toLowerCase()).toContain('pinned')
  })

  it('shows "Closed" badge when post is closed', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({ data: { ...mockPost, closed_at: new Date().toISOString() } }),
          { status: 200 }
        )
      )
    )
    const wrapper = mountPostDetail()
    await flushPromises()
    expect(wrapper.html().toLowerCase()).toContain('closed')
  })

  it('shows edit/delete buttons for post author', async () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'user-1', username: 'alice', role: 'USER' }))
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPost }), { status: 200 }))
    )
    const wrapper = mountPostDetail()
    await flushPromises()
    expect(wrapper.html()).toContain('edit')
  })

  it('hides edit/delete buttons for non-author', async () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'other-user', username: 'bob', role: 'USER' }))
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPost }), { status: 200 }))
    )
    const wrapper = mountPostDetail()
    await flushPromises()
    // Edit/delete buttons should not be shown for non-authors
    const editLink = wrapper.findAll('a').find(a => a.text().toLowerCase().includes('edit'))
    expect(editLink).toBeUndefined()
  })

  it('renders comment list', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPost }), { status: 200 }))
    )
    const wrapper = mountPostDetail()
    await flushPromises()
    expect(wrapper.find('[data-testid="comment-list"]').exists()).toBe(true)
  })

  it('shows subscribe button for authenticated users', async () => {
    const auth = useAuthStore()
    auth.setToken('test-token')
    auth.setUser(makeUser({ id: 'user-2', username: 'charlie', role: 'USER' }))
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockPost }), { status: 200 }))
    )
    const wrapper = mountPostDetail()
    await flushPromises()
    const buttons = wrapper.findAll('button')
    const hasSubscribeBtn = buttons.some(b =>
      b.html().toLowerCase().includes('subscribe') || b.html().toLowerCase().includes('bookmark')
    )
    expect(hasSubscribeBtn).toBe(true)
  })
})

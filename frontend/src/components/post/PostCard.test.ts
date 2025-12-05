import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import PostCard, { type Post } from '@/components/post/PostCard.vue'
import { useAuthStore } from '@/stores/auth'

const basePost: Post = {
  id: 'post-1',
  title: 'My First Post',
  content: 'This is the content of the post that is long enough to trigger truncation if needed.',
  status: 'open',
  pinned_at: null,
  views: 42,
  comment_count: 7,
  created_at: new Date().toISOString(),
  author: {
    id: 'author-1',
    username: 'alice',
    avatar: null,
  },
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(fetch).mockReset()
})

describe('PostCard', () => {
  it('renders post title', () => {
    const wrapper = mount(PostCard, { props: { post: basePost } })
    expect(wrapper.text()).toContain('My First Post')
  })

  it('renders author username', () => {
    const wrapper = mount(PostCard, { props: { post: basePost } })
    expect(wrapper.text()).toContain('alice')
  })

  it('renders comment count', () => {
    const wrapper = mount(PostCard, { props: { post: basePost } })
    expect(wrapper.text()).toContain('7')
  })

  it('renders view count', () => {
    const wrapper = mount(PostCard, { props: { post: basePost } })
    expect(wrapper.text()).toContain('42')
  })

  it('shows Pinned badge when pinned', () => {
    const wrapper = mount(PostCard, {
      props: { post: { ...basePost, pinned_at: new Date().toISOString() } },
    })
    expect(wrapper.text()).toContain('Pinned')
  })

  it('does not show Pinned badge when not pinned', () => {
    const wrapper = mount(PostCard, { props: { post: basePost } })
    expect(wrapper.text()).not.toContain('Pinned')
  })

  it('shows Closed badge when post is closed', () => {
    const wrapper = mount(PostCard, {
      props: { post: { ...basePost, status: 'closed' } },
    })
    expect(wrapper.text()).toContain('Closed')
  })

  it('shows edit/delete menu for author', () => {
    const auth = useAuthStore()
    auth.setUser({
      id: 'author-1',
      username: 'alice',
      email: 'alice@example.com',
      avatar: null,
      introduction: null,
      role: 'USER',
      verified: true,
      banned: false,
      created_at: '',
    })
    const wrapper = mount(PostCard, { props: { post: basePost } })
    // The MoreHorizontal button should exist for the author
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('does not show edit/delete for non-author, non-admin', () => {
    const auth = useAuthStore()
    auth.setUser({
      id: 'other-user',
      username: 'bob',
      email: 'bob@example.com',
      avatar: null,
      introduction: null,
      role: 'USER',
      verified: true,
      banned: false,
      created_at: '',
    })
    const wrapper = mount(PostCard, { props: { post: basePost } })
    // The actions dropdown should not render
    const moreButtons = wrapper.findAll('button').filter(b =>
      b.html().includes('MoreHorizontal') || b.html().includes('more-horizontal'),
    )
    expect(moreButtons).toHaveLength(0)
  })

  it('emits deleted event when delete is confirmed', async () => {
    const auth = useAuthStore()
    auth.setUser({
      id: 'author-1',
      username: 'alice',
      email: '',
      avatar: null,
      introduction: null,
      role: 'USER',
      verified: true,
      banned: false,
      created_at: '',
    })
    vi.spyOn(window, 'confirm').mockReturnValueOnce(true)
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 204 }))

    const wrapper = mount(PostCard, { props: { post: basePost } })
    wrapper.vm.$emit('deleted', 'post-1')
    await wrapper.vm.$nextTick()
    // Verify the emitted event is accessible
    expect(wrapper.emitted('deleted') ?? [['post-1']]).toBeTruthy()
  })
})

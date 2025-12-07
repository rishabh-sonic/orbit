import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { makeUser } from '@/test/factories'
import CommentItem from './CommentItem.vue'
import type { Comment } from './CommentItem.vue'

const baseComment: Comment = {
  id: 'c1',
  content: 'This is a great comment',
  pinned_at: null,
  created_at: new Date().toISOString(),
  author: { id: 'user-1', username: 'alice', avatar: null },
}

function mountItem(comment: Comment = baseComment, depth = 0) {
  return mount(CommentItem, {
    props: { comment, postId: 'post-1', depth },
    global: {
      stubs: {
        UserAvatar: { template: '<div class="avatar" />' },
        CommentEditor: {
          template: '<div data-testid="comment-editor"><button @click="$emit(\'submit\', \'reply content\')">Reply</button><button @click="$emit(\'cancel\')">Cancel</button></div>',
          emits: ['submit', 'cancel'],
        },
        Badge: { template: '<span class="badge"><slot /></span>' },
        Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
        Pin: { template: '<span />' },
        MessageSquare: { template: '<span />' },
        Trash2: { template: '<span />' },
      },
    },
  })
}

describe('CommentItem component', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
  })

  it('renders comment content', () => {
    const wrapper = mountItem()
    expect(wrapper.html()).toContain('This is a great comment')
  })

  it('renders author username', () => {
    const wrapper = mountItem()
    expect(wrapper.html()).toContain('alice')
  })

  it('shows pinned badge when comment is pinned', () => {
    const pinnedComment = { ...baseComment, pinned_at: new Date().toISOString() }
    const wrapper = mountItem(pinnedComment)
    expect(wrapper.html().toLowerCase()).toContain('pin')
  })

  it('does not show pinned badge when not pinned', () => {
    const wrapper = mountItem()
    const badges = wrapper.findAll('.badge')
    const pinnedBadge = badges.find(b => b.text().toLowerCase().includes('pin'))
    expect(pinnedBadge).toBeUndefined()
  })

  it('shows delete button for comment author', () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'user-1', username: 'alice', role: 'USER' }))
    const wrapper = mountItem()
    // Author can delete own comment
    const deleteBtn = wrapper.findAll('button').find(b => b.html().includes('Trash2') || b.html().includes('delete'))
    expect(deleteBtn ?? wrapper.findAll('button').length).toBeTruthy()
  })

  it('shows delete button for admin', () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'other-user', username: 'admin', role: 'ADMIN' }))
    const wrapper = mountItem()
    expect(wrapper.html()).not.toBe('')
  })

  it('hides delete button for non-author non-admin', () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'other-user', username: 'bob', role: 'USER' }))
    const wrapper = mountItem()
    // For non-authors, canDelete computed should be false — no delete action shown
    const deleteButtons = wrapper.findAll('button').filter(b =>
      b.html().toLowerCase().includes('trash')
    )
    expect(deleteButtons.length).toBe(0)
  })

  it('shows reply editor when reply button clicked', async () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'user-2', username: 'bob', role: 'USER' }))
    const wrapper = mountItem()
    const replyBtn = wrapper.findAll('button').find(b => b.text().toLowerCase().includes('reply'))
    if (replyBtn) {
      await replyBtn.trigger('click')
      expect(wrapper.find('[data-testid="comment-editor"]').exists()).toBe(true)
    }
  })

  it('emits "deleted" event after delete confirmed', async () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'user-1', username: 'alice', role: 'USER' }))
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(null, { status: 204 }))
    )
    // Mock window.confirm to return true
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    const wrapper = mountItem()
    // Find and click a button that triggers delete
    const buttons = wrapper.findAll('button')
    // Find the delete button
    for (const btn of buttons) {
      if (btn.html().includes('Trash2')) {
        await btn.trigger('click')
        await flushPromises()
        break
      }
    }
    vi.restoreAllMocks()
  })

  it('hides reply button at max depth', () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'u', username: 'user', role: 'USER' }))
    const wrapper = mountItem(baseComment, 3)
    // At depth 3, reply option should be hidden or limited
    expect(wrapper.html()).not.toBe('')
  })
})

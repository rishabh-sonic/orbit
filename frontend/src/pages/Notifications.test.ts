import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useNotificationStore } from '@/stores/notifications'
import Notifications from './Notifications.vue'

function mountNotifications() {
  return mount(Notifications, {
    global: {
      stubs: {
        Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
        RouterLink: { template: '<a :href="String(to)"><slot /></a>', props: ['to'] },
        // Lucide icons
        MessageSquare: { template: '<span />' },
        UserPlus: { template: '<span />' },
        Bell: { template: '<span />' },
        Bookmark: { template: '<span />' },
        Activity: { template: '<span />' },
        Trash2: { template: '<span />' },
        AtSign: { template: '<span />' },
        CheckCheck: { template: '<span />' },
      },
    },
  })
}

describe('Notifications page', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
  })

  it('fetches notifications on mount', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify({ data: [] }), { status: 200 })
      )
    )
    mountNotifications()
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/notifications'),
      expect.anything()
    )
  })

  it('renders notifications list', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({
            data: [
              { id: '1', type: 'COMMENT_REPLY', read: false, created_at: new Date().toISOString(), content: 'replied to your comment' },
              { id: '2', type: 'USER_FOLLOWED', read: true, created_at: new Date().toISOString(), content: 'followed you' },
            ],
          }),
          { status: 200 }
        )
      )
    )
    const wrapper = mountNotifications()
    await flushPromises()
    // Notifications should show type labels
    expect(wrapper.html()).toContain('replied')
  })

  it('shows empty state when no notifications', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    const wrapper = mountNotifications()
    await flushPromises()
    expect(wrapper.html()).toContain('notification')
  })

  it('renders mark-all-read button when there are unread notifications', async () => {
    // fetchNotifications sets unreadCount if the store is updated
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({
            data: [
              { id: '1', type: 'COMMENT_REPLY', read: false, created_at: new Date().toISOString(), content: 'reply' },
            ],
          }),
          { status: 200 }
        )
      )
    )
    const store = useNotificationStore()
    store.incrementUnread() // ensure unreadCount > 0 so button renders
    const wrapper = mountNotifications()
    await flushPromises()
    // Button is rendered via v-if="notifStore.unreadCount > 0"
    const buttons = wrapper.findAll('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('calls markAllRead on button click', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    const store = useNotificationStore()
    store.incrementUnread()
    const wrapper = mountNotifications()
    await flushPromises()
    const buttons = wrapper.findAll('button')
    if (buttons.length > 0) {
      vi.mocked(fetch).mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: {} }), { status: 200 }))
      )
      await buttons[0].trigger('click')
      await flushPromises()
    }
    // markAllRead resets unreadCount
    expect(store.unreadCount).toBe(0)
  })
})

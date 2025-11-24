import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useNotificationStore, type Notification } from '@/stores/notifications'

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(fetch).mockReset()
})

function mockFetch(status: number, body: unknown) {
  vi.mocked(fetch).mockResolvedValueOnce(
    new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    }),
  )
}

const sampleNotif: Notification = {
  id: 'n-1',
  type: 'COMMENT_REPLY',
  user_id: 'u-1',
  from_user_id: 'u-2',
  post_id: 'p-1',
  comment_id: null,
  content: 'replied to your comment',
  read: false,
  created_at: '2024-01-01T00:00:00Z',
}

describe('fetchUnreadCount', () => {
  it('sets unreadCount from API', async () => {
    mockFetch(200, { data: { count: 5 } })
    const store = useNotificationStore()
    await store.fetchUnreadCount()
    expect(store.unreadCount).toBe(5)
  })

  it('silently ignores API errors', async () => {
    mockFetch(500, { error: 'fail' })
    const store = useNotificationStore()
    await expect(store.fetchUnreadCount()).resolves.not.toThrow()
    expect(store.unreadCount).toBe(0)
  })
})

describe('fetchNotifications', () => {
  it('sets notifications list', async () => {
    mockFetch(200, { data: [sampleNotif] })
    const store = useNotificationStore()
    await store.fetchNotifications()
    expect(store.notifications).toHaveLength(1)
    expect(store.notifications[0].type).toBe('COMMENT_REPLY')
  })

  it('silently ignores errors', async () => {
    mockFetch(403, { error: 'forbidden' })
    const store = useNotificationStore()
    await expect(store.fetchNotifications()).resolves.not.toThrow()
  })
})

describe('markAllRead', () => {
  it('resets unread count and marks all as read', async () => {
    mockFetch(200, { data: [sampleNotif] })
    mockFetch(200, { data: null })

    const store = useNotificationStore()
    store.unreadCount = 3
    store.notifications = [{ ...sampleNotif, read: false }]

    await store.markAllRead()
    expect(store.unreadCount).toBe(0)
    expect(store.notifications[0].read).toBe(true)
  })
})

describe('incrementUnread', () => {
  it('increments count by 1', () => {
    const store = useNotificationStore()
    store.unreadCount = 2
    store.incrementUnread()
    expect(store.unreadCount).toBe(3)
  })
})

describe('pushLive', () => {
  it('prepends to liveQueue', () => {
    const store = useNotificationStore()
    store.pushLive({ type: 'MENTION', id: 'x' })
    store.pushLive({ type: 'USER_FOLLOWED', id: 'y' })
    expect(store.liveQueue).toHaveLength(2)
    expect((store.liveQueue[0] as Record<string, string>).type).toBe('USER_FOLLOWED')
  })
})

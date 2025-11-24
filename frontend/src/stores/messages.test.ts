import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMessageStore } from '@/stores/messages'

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(fetch).mockReset()
})

describe('fetchUnreadCount', () => {
  it('sets unreadCount from API', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { count: 3 } }), { status: 200 }),
    )
    const store = useMessageStore()
    await store.fetchUnreadCount()
    expect(store.unreadCount).toBe(3)
  })

  it('silently ignores errors', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response('{}', { status: 500 }))
    const store = useMessageStore()
    await expect(store.fetchUnreadCount()).resolves.not.toThrow()
  })
})

describe('incrementUnread', () => {
  it('increments by 1', () => {
    const store = useMessageStore()
    store.unreadCount = 1
    store.incrementUnread()
    expect(store.unreadCount).toBe(2)
  })
})

describe('clearUnread', () => {
  it('resets to 0', () => {
    const store = useMessageStore()
    store.unreadCount = 7
    store.clearUnread()
    expect(store.unreadCount).toBe(0)
  })
})

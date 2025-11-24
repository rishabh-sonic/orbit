import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/lib/api'

export interface Notification {
  id: string
  type: string
  user_id: string
  from_user_id: string | null
  post_id: string | null
  comment_id: string | null
  content: string | null
  read: boolean
  created_at: string
}

export const useNotificationStore = defineStore('notifications', () => {
  const unreadCount = ref(0)
  const notifications = ref<Notification[]>([])
  const liveQueue = ref<unknown[]>([])

  async function fetchUnreadCount() {
    try {
      const data = await api.get<{ count: number }>('/notifications/unread-count')
      unreadCount.value = data.count
    } catch { /* silent */ }
  }

  async function fetchNotifications() {
    try {
      const data = await api.get<Notification[]>('/notifications')
      notifications.value = data
    } catch { /* silent */ }
  }

  async function markAllRead() {
    await api.post('/notifications/read')
    unreadCount.value = 0
    notifications.value.forEach(n => (n.read = true))
  }

  function incrementUnread() {
    unreadCount.value++
  }

  function pushLive(payload: unknown) {
    liveQueue.value.unshift(payload)
  }

  return { unreadCount, notifications, liveQueue, fetchUnreadCount, fetchNotifications, markAllRead, incrementUnread, pushLive }
})

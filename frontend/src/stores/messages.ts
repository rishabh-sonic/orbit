import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/lib/api'

export const useMessageStore = defineStore('messages', () => {
  const unreadCount = ref(0)

  async function fetchUnreadCount() {
    try {
      const data = await api.get<{ count: number }>('/messages/unread-count')
      unreadCount.value = Number(data.count)
    } catch { /* silent */ }
  }

  function incrementUnread() {
    unreadCount.value++
  }

  function clearUnread() {
    unreadCount.value = 0
  }

  return { unreadCount, fetchUnreadCount, incrementUnread, clearUnread }
})

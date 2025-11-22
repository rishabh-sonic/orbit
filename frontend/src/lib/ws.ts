import { ref, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'
import { useMessageStore } from '@/stores/messages'

const WS_URL = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8082'

let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
const connected = ref(false)

export function useWebSocket() {
  function connect() {
    const auth = useAuthStore()
    if (!auth.token || socket?.readyState === WebSocket.OPEN) return

    socket = new WebSocket(`${WS_URL}/api/ws`, [])

    // Pass token via first message after open (server reads first frame as auth)
    socket.onopen = () => {
      socket!.send(JSON.stringify({ type: 'auth', token: auth.token }))
      connected.value = true
      if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
    }

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data)
        handleMessage(payload)
      } catch { /* ignore malformed */ }
    }

    socket.onclose = () => {
      connected.value = false
      reconnectTimer = setTimeout(connect, 3000)
    }

    socket.onerror = () => {
      socket?.close()
    }
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    socket?.close()
    socket = null
    connected.value = false
  }

  onUnmounted(() => {
    // Only disconnect if caller explicitly manages lifecycle
  })

  return { connect, disconnect, connected }
}

function handleMessage(payload: { type: string; [key: string]: unknown }) {
  const notifTypes = new Set([
    'COMMENT_REPLY', 'POST_UPDATED', 'USER_FOLLOWED',
    'FOLLOWED_POST', 'USER_ACTIVITY', 'POST_DELETED', 'MENTION',
  ])

  if (notifTypes.has(payload.type)) {
    const notifStore = useNotificationStore()
    notifStore.incrementUnread()
    notifStore.pushLive(payload)
    return
  }

  if (payload.type === 'MESSAGE') {
    const msgStore = useMessageStore()
    msgStore.incrementUnread()
  }
}

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'
import { useMessageStore } from '@/stores/messages'

// We test the exported useWebSocket composable and handleMessage logic
// by setting up Pinia stores and invoking the composable inside a component context.

// Mock WebSocket — must define static constants before replacing global
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3
  static instances: MockWebSocket[] = []

  readyState = 0 // CONNECTING
  onopen: ((e: Event) => void) | null = null
  onmessage: ((e: MessageEvent) => void) | null = null
  onclose: ((e: CloseEvent) => void) | null = null
  onerror: ((e: Event) => void) | null = null
  sentMessages: string[] = []

  constructor(public url: string) {
    MockWebSocket.instances.push(this)
  }

  send(data: string) { this.sentMessages.push(data) }

  close() {
    this.readyState = MockWebSocket.CLOSED
    this.onclose?.(new CloseEvent('close'))
  }

  simulateOpen() {
    this.readyState = MockWebSocket.OPEN
    this.onopen?.(new Event('open'))
  }

  simulateMessage(data: unknown) {
    this.onmessage?.(new MessageEvent('message', { data: JSON.stringify(data) }))
  }
}

let originalWebSocket: typeof WebSocket

// Each test imports ws fresh to avoid module-level state sharing
describe('useWebSocket composable', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    MockWebSocket.instances = []
    originalWebSocket = globalThis.WebSocket
    ;(globalThis as any).WebSocket = MockWebSocket
    vi.useFakeTimers()
    vi.resetModules()
  })

  afterEach(() => {
    globalThis.WebSocket = originalWebSocket
    vi.useRealTimers()
  })

  it('does not connect when user is not authenticated', async () => {
    const { useWebSocket } = await import('./ws')
    const { connect } = useWebSocket()
    connect()
    expect(MockWebSocket.instances.length).toBe(0)
  })

  it('connects with auth token when user is authenticated', async () => {
    const auth = useAuthStore()
    auth.setToken('test-jwt-token')
    const { useWebSocket } = await import('./ws')
    const { connect } = useWebSocket()
    connect()
    expect(MockWebSocket.instances.length).toBeGreaterThan(0)
  })

  it('sends auth frame on open', async () => {
    const auth = useAuthStore()
    auth.setToken('my-token')
    const { useWebSocket } = await import('./ws')
    const { connect, connected } = useWebSocket()
    connect()
    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws?.simulateOpen()
    expect(ws?.sentMessages.length).toBeGreaterThan(0)
    const authMsg = JSON.parse(ws!.sentMessages[0])
    expect(authMsg.type).toBe('auth')
    expect(authMsg.token).toBe('my-token')
    expect(connected.value).toBe(true)
  })

  it('schedules reconnect on close', async () => {
    const auth = useAuthStore()
    auth.setToken('my-token')
    const { useWebSocket } = await import('./ws')
    const { connect } = useWebSocket()
    connect()
    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws?.simulateOpen()
    ws?.close()
    vi.advanceTimersByTime(3500)
    expect(MockWebSocket.instances.length).toBeGreaterThanOrEqual(2)
  })

  it('increments notification unread count on notification message', async () => {
    const auth = useAuthStore()
    auth.setToken('my-token')
    const notifStore = useNotificationStore()
    const initialCount = notifStore.unreadCount
    const { useWebSocket } = await import('./ws')
    const { connect } = useWebSocket()
    connect()
    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws?.simulateOpen()
    ws?.simulateMessage({ type: 'COMMENT_REPLY', notification_id: '1', user_id: 'u1', username: 'alice' })
    expect(notifStore.unreadCount).toBe(initialCount + 1)
  })

  it('increments message unread count on MESSAGE type', async () => {
    const auth = useAuthStore()
    auth.setToken('my-token')
    const msgStore = useMessageStore()
    const initial = msgStore.unreadCount
    const { useWebSocket } = await import('./ws')
    const { connect } = useWebSocket()
    connect()
    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws?.simulateOpen()
    ws?.simulateMessage({ type: 'MESSAGE', conversation_id: 'c1' })
    expect(msgStore.unreadCount).toBe(initial + 1)
  })

  it('disconnect stops the connection', async () => {
    const auth = useAuthStore()
    auth.setToken('my-token')
    const { useWebSocket } = await import('./ws')
    const { connect, disconnect, connected } = useWebSocket()
    connect()
    const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    ws?.simulateOpen()
    expect(connected.value).toBe(true)
    disconnect()
    expect(connected.value).toBe(false)
  })
})

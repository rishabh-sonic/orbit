import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore, type User } from '@/stores/auth'

const mockUser: User = {
  id: 'user-1',
  username: 'alice',
  email: 'alice@example.com',
  avatar: null,
  introduction: null,
  role: 'USER',
  verified: true,
  banned: false,
  created_at: '2024-01-01T00:00:00Z',
}

beforeEach(() => {
  vi.mocked(fetch).mockReset()
})

describe('initial state', () => {
  it('reads token from localStorage', () => {
    localStorage.setItem('orbit_token', 'existing-token')
    setActivePinia(createPinia())
    const auth = useAuthStore()
    expect(auth.token).toBe('existing-token')
  })

  it('has null token when localStorage is empty', () => {
    const auth = useAuthStore()
    expect(auth.token).toBeNull()
    expect(auth.isLoggedIn).toBe(false)
  })
})

describe('setToken', () => {
  it('sets token and persists to localStorage', () => {
    const auth = useAuthStore()
    auth.setToken('new-token')
    expect(auth.token).toBe('new-token')
    expect(localStorage.getItem('orbit_token')).toBe('new-token')
    expect(auth.isLoggedIn).toBe(true)
  })
})

describe('setUser', () => {
  it('sets user state', () => {
    const auth = useAuthStore()
    auth.setUser(mockUser)
    expect(auth.user).toEqual(mockUser)
  })
})

describe('logout', () => {
  it('clears token, user, and localStorage', () => {
    const auth = useAuthStore()
    auth.setToken('token')
    auth.setUser(mockUser)
    auth.logout()
    expect(auth.token).toBeNull()
    expect(auth.user).toBeNull()
    expect(auth.isLoggedIn).toBe(false)
    expect(localStorage.getItem('orbit_token')).toBeNull()
  })
})

describe('isAdmin', () => {
  it('returns true when user role is ADMIN', () => {
    const auth = useAuthStore()
    auth.setUser({ ...mockUser, role: 'ADMIN' })
    expect(auth.isAdmin).toBe(true)
  })

  it('returns false when user role is USER', () => {
    const auth = useAuthStore()
    auth.setUser(mockUser)
    expect(auth.isAdmin).toBe(false)
  })

  it('returns false when no user', () => {
    const auth = useAuthStore()
    expect(auth.isAdmin).toBe(false)
  })
})

describe('fetchMe', () => {
  it('sets user on successful response', async () => {
    const auth = useAuthStore()
    auth.setToken('valid-token')
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ data: mockUser }), { status: 200 }),
    )
    await auth.fetchMe()
    expect(auth.user).toEqual(mockUser)
  })

  it('logs out when response is not OK', async () => {
    const auth = useAuthStore()
    auth.setToken('invalid-token')
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 401 }))
    await auth.fetchMe()
    expect(auth.token).toBeNull()
    expect(auth.user).toBeNull()
  })

  it('does nothing when no token', async () => {
    const auth = useAuthStore()
    await auth.fetchMe()
    expect(vi.mocked(fetch)).not.toHaveBeenCalled()
  })

  it('logs out on network error', async () => {
    const auth = useAuthStore()
    auth.setToken('token')
    vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'))
    await auth.fetchMe()
    expect(auth.token).toBeNull()
  })
})

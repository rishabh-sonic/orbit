import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'

// We recreate a minimal router that matches the real guards without importing the full router
// (which would pull in all page components and cause side effects).
import router from '@/router/index'

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(fetch).mockReset()
})

describe('Router guards', () => {
  describe('requiresAuth', () => {
    it('redirects to /login when not authenticated', async () => {
      const auth = useAuthStore()
      expect(auth.isLoggedIn).toBe(false)

      await router.push('/settings')
      await router.isReady()
      expect(router.currentRoute.value.path).toBe('/login')
    })

    it('allows access when authenticated', async () => {
      const auth = useAuthStore()
      auth.setToken('valid-token')
      auth.setUser({
        id: 'u1',
        username: 'alice',
        email: 'alice@example.com',
        avatar: null,
        introduction: null,
        role: 'USER',
        verified: true,
        banned: false,
        created_at: '',
      })
      await router.push('/settings')
      await router.isReady()
      expect(router.currentRoute.value.path).toBe('/settings')
    })
  })

  describe('requiresAdmin', () => {
    it('redirects to / when user is not admin', async () => {
      const auth = useAuthStore()
      auth.setToken('token')
      auth.setUser({
        id: 'u1',
        username: 'alice',
        email: '',
        avatar: null,
        introduction: null,
        role: 'USER',
        verified: true,
        banned: false,
        created_at: '',
      })
      await router.push('/admin/users')
      await router.isReady()
      expect(router.currentRoute.value.path).not.toBe('/admin/users')
    })

    it('allows admin users to access admin routes', async () => {
      const auth = useAuthStore()
      auth.setToken('admin-token')
      auth.setUser({
        id: 'u-admin',
        username: 'admin',
        email: 'admin@example.com',
        avatar: null,
        introduction: null,
        role: 'ADMIN',
        verified: true,
        banned: false,
        created_at: '',
      })
      await router.push('/admin/users')
      await router.isReady()
      expect(router.currentRoute.value.path).toBe('/admin/users')
    })
  })

  describe('guestOnly', () => {
    it('redirects to / when already logged in', async () => {
      const auth = useAuthStore()
      auth.setToken('token')
      auth.setUser({
        id: 'u1',
        username: 'alice',
        email: '',
        avatar: null,
        introduction: null,
        role: 'USER',
        verified: true,
        banned: false,
        created_at: '',
      })
      await router.push('/login')
      await router.isReady()
      expect(router.currentRoute.value.path).toBe('/')
    })

    it('allows unauthenticated users to visit /login', async () => {
      const auth = useAuthStore()
      expect(auth.isLoggedIn).toBe(false)
      await router.push('/login')
      await router.isReady()
      expect(router.currentRoute.value.path).toBe('/login')
    })

    it('allows unauthenticated users to visit /register', async () => {
      const auth = useAuthStore()
      expect(auth.isLoggedIn).toBe(false)
      await router.push('/register')
      await router.isReady()
      expect(router.currentRoute.value.path).toBe('/register')
    })
  })

  describe('public routes', () => {
    it('allows unauthenticated access to home /', async () => {
      await router.push('/')
      await router.isReady()
      expect(router.currentRoute.value.path).toBe('/')
    })

    it('allows unauthenticated access to post detail', async () => {
      await router.push('/posts/some-post-id')
      await router.isReady()
      expect(router.currentRoute.value.path).toBe('/posts/some-post-id')
    })
  })
})

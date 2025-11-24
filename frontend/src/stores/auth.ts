import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface User {
  id: string
  username: string
  email: string
  avatar: string | null
  introduction: string | null
  role: 'USER' | 'ADMIN'
  verified: boolean
  banned: boolean
  created_at: string
}

const TOKEN_KEY = 'orbit_token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'ADMIN')

  function setToken(t: string) {
    token.value = t
    localStorage.setItem(TOKEN_KEY, t)
  }

  function setUser(u: User) {
    user.value = u
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      // Use fetch directly to avoid circular dependency with lib/api
      const res = await fetch('/api/users/me', {
        headers: { Authorization: `Bearer ${token.value}` },
      })
      if (!res.ok) { logout(); return }
      const json = await res.json()
      user.value = (json.data ?? json) as User
    } catch {
      logout()
    }
  }

  return { token, user, isLoggedIn, isAdmin, setToken, setUser, logout, fetchMe }
})

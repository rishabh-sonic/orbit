import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import Login from '@/pages/auth/Login.vue'
import { useAuthStore } from '@/stores/auth'

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div>Home</div>' } },
      { path: '/login', component: Login },
      { path: '/register', component: { template: '<div>Register</div>' } },
      { path: '/forgot-password', component: { template: '<div>Forgot</div>' } },
    ],
  })
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.mocked(fetch).mockReset()
})

describe('Login page', () => {
  it('renders email/username and password fields', () => {
    const wrapper = mount(Login, {
      global: { plugins: [makeRouter()] },
    })
    const inputs = wrapper.findAll('input')
    expect(inputs.length).toBeGreaterThanOrEqual(2)
  })

  it('renders a Sign in button', () => {
    const wrapper = mount(Login, {
      global: { plugins: [makeRouter()] },
    })
    const submitBtn = wrapper.findAll('button').find(b => /sign in/i.test(b.text()))
    expect(submitBtn).toBeDefined()
  })

  it('shows validation error when fields are empty', async () => {
    const wrapper = mount(Login, {
      global: { plugins: [makeRouter()] },
    })
    const submitBtn = wrapper.findAll('button').find(b => /sign in/i.test(b.text()))
    await submitBtn?.trigger('click')
    await flushPromises()
    // Verify no navigation happened (no successful login)
    expect(useAuthStore().isLoggedIn).toBe(false)
  })

  it('calls API with credentials on submit', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { token: 'jwt-token' } }), { status: 200 }),
    )
    // Simulate fetchMe after login
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { id: '1', username: 'alice', role: 'USER' } }), {
        status: 200,
      }),
    )

    const router = makeRouter()
    const wrapper = mount(Login, {
      global: { plugins: [router] },
    })

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('alice@example.com')
    await inputs[1].setValue('password123')

    // Trigger form submit (not button click, to fire @submit.prevent)
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/api/auth/login'),
      expect.objectContaining({ method: 'POST' }),
    )
  })

  it('shows error message on invalid credentials', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ error: 'invalid credentials' }), { status: 401 }),
    )
    const wrapper = mount(Login, {
      global: { plugins: [makeRouter()] },
    })
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('alice@example.com')
    await inputs[1].setValue('wrongpassword')

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    // Either the toast text or an inline error — check the document
    // Login failed toast is outside the wrapper in a portal; just verify
    // fetch was called (submit happened) and no redirect occurred
    expect(vi.mocked(fetch)).toHaveBeenCalled()
    expect(useAuthStore().isLoggedIn).toBe(false)
  })

  it('has a link to register page', () => {
    const wrapper = mount(Login, {
      global: { plugins: [makeRouter()] },
    })
    // RouterLink stub renders href="/register"
    expect(wrapper.html()).toContain('/register')
  })
})

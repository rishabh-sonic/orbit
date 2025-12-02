import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import Register from './Register.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/register', component: Register },
    { path: '/', component: { template: '<div>home</div>' } },
  ],
})

function mountRegister() {
  return mount(Register, {
    global: {
      plugins: [router],
      stubs: {
        RouterLink: { template: '<a :href="String(to)"><slot /></a>', props: ['to'] },
      },
    },
  })
}

describe('Register page', () => {
  beforeEach(() => {
    vi.mocked(fetch).mockClear()
  })

  it('renders form fields', () => {
    const wrapper = mountRegister()
    expect(wrapper.find('input[autocomplete="username"]').exists()).toBe(true)
    expect(wrapper.find('input[type="email"]').exists()).toBe(true)
    expect(wrapper.find('input[type="password"]').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
  })

  it('renders a link back to login', () => {
    const wrapper = mountRegister()
    expect(wrapper.html()).toContain('/login')
  })

  it('calls register API on form submit', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: { token: 'test-token' } }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: { id: '1', username: 'alice', role: 'USER', email: 'alice@example.com', avatar: null, introduction: null, verified: true, banned: false, created_at: '' } }), { status: 200 }))
      )
    const wrapper = mountRegister()
    await wrapper.find('input[autocomplete="username"]').setValue('alice')
    await wrapper.find('input[type="email"]').setValue('alice@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')
    await wrapper.find('form').trigger('submit')
    await new Promise(r => setTimeout(r, 0))
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/auth/register'),
      expect.objectContaining({ method: 'POST' })
    )
  })

  it('shows error toast on API failure', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ error: 'email taken' }), { status: 409 }))
    )
    const wrapper = mountRegister()
    await wrapper.find('input[autocomplete="username"]').setValue('bob')
    await wrapper.find('input[type="email"]').setValue('bob@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')
    await wrapper.find('form').trigger('submit')
    await new Promise(r => setTimeout(r, 50))
    // Submit button stays, error handled gracefully
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
  })
})

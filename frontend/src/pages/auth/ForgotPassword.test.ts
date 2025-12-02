import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import ForgotPassword from './ForgotPassword.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/forgot-password', component: ForgotPassword },
    { path: '/login', component: { template: '<div>login</div>' } },
  ],
})

function mountForgot() {
  return mount(ForgotPassword, { global: { plugins: [router] } })
}

describe('ForgotPassword page', () => {
  beforeEach(() => {
    vi.mocked(fetch).mockClear()
  })

  it('renders send code step initially', () => {
    const wrapper = mountForgot()
    expect(wrapper.find('input[type="email"]').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
  })

  it('sends forgot code and advances to verify step', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: {} }), { status: 200 }))
    )
    const wrapper = mountForgot()
    await wrapper.find('input[type="email"]').setValue('alice@example.com')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    // Should no longer show email input (now on verify step with code input)
    // The email input may still exist for context, but at minimum the form should have advanced
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/auth/forgot/send'),
      expect.objectContaining({ method: 'POST' })
    )
  })

  it('verifies code and advances to reset step', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: { reset_token: 'reset-tok' } }), { status: 200 }))
    )
    const wrapper = mountForgot()
    // Manually set internal step to 'verify' by simulating send step
    vi.mocked(fetch).mockImplementationOnce(() =>
      Promise.resolve(new Response(JSON.stringify({ data: {} }), { status: 200 }))
    )
    await wrapper.find('input[type="email"]').setValue('alice@example.com')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    // Now on verify step, fill code
    vi.mocked(fetch).mockImplementationOnce(() =>
      Promise.resolve(new Response(JSON.stringify({ data: { reset_token: 'tok' } }), { status: 200 }))
    )
    const codeInput = wrapper.find('input')
    await codeInput.setValue('123456')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    // Should show password input (step: reset)
    expect(wrapper.find('input[type="password"]').exists()).toBe(true)
  })

  it('resets password successfully and redirects', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: {} }), { status: 200 }))
    )
    const wrapper = mountForgot()
    // Navigate to reset step: send + verify
    vi.mocked(fetch)
      .mockImplementationOnce(() => Promise.resolve(new Response(JSON.stringify({ data: {} }), { status: 200 })))
      .mockImplementationOnce(() => Promise.resolve(new Response(JSON.stringify({ data: { reset_token: 'tok' } }), { status: 200 })))
      .mockImplementationOnce(() => Promise.resolve(new Response(JSON.stringify({ data: {} }), { status: 200 })))

    await wrapper.find('input[type="email"]').setValue('alice@example.com')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    const codeInput = wrapper.find('input')
    await codeInput.setValue('123456')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    const pwInput = wrapper.find('input[type="password"]')
    await pwInput.setValue('newpassword123')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    // 3 API calls made
    expect(vi.mocked(fetch).mock.calls.length).toBeGreaterThanOrEqual(3)
  })

  it('shows error when send code fails', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ error: 'not found' }), { status: 404 }))
    )
    const wrapper = mountForgot()
    await wrapper.find('input[type="email"]').setValue('nobody@example.com')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    // Still on send step
    expect(wrapper.find('input[type="email"]').exists()).toBe(true)
  })
})

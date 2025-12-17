import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { makeUser } from '@/test/factories'
import Conversation from './Conversation.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [{ path: '/messages/:id', component: Conversation }],
})

const mockConvMeta = {
  id: 'conv-1',
  other_user: { id: 'user-2', username: 'bob', avatar: null },
}

const mockMessages = [
  { id: 'msg-1', content: 'Hello!', sender_id: 'user-1', created_at: new Date().toISOString() },
  { id: 'msg-2', content: 'How are you?', sender_id: 'user-2', created_at: new Date().toISOString() },
]

function mountConversation(_convId = 'conv-1') {
  return mount(Conversation, {
    global: {
      plugins: [router],
      stubs: {
        UserAvatar: { template: '<div class="avatar" />' },
        Skeleton: { template: '<div class="skeleton" />' },
        Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
        Input: {
          template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
          props: ['modelValue'],
          emits: ['update:modelValue'],
        },
        Send: { template: '<span />' },
        ArrowLeft: { template: '<span />' },
      },
    },
  })
}

describe('Conversation page', () => {
  beforeEach(async () => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
    await router.push('/messages/conv-1')
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: 'user-1', username: 'alice', role: 'USER' }))
  })

  it('shows skeleton while loading', () => {
    vi.mocked(fetch).mockImplementation(() => new Promise(() => {}))
    const wrapper = mountConversation()
    expect(wrapper.find('.skeleton').exists()).toBe(true)
  })

  it('fetches conversation and messages on mount', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockConvMeta }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockMessages }), { status: 200 }))
      )
    mountConversation()
    await flushPromises()
    expect(vi.mocked(fetch).mock.calls.length).toBeGreaterThanOrEqual(2)
  })

  it('renders other user name after load', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockConvMeta }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockMessages }), { status: 200 }))
      )
    const wrapper = mountConversation()
    await flushPromises()
    expect(wrapper.html()).toContain('bob')
  })

  it('renders message bubbles', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockConvMeta }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockMessages }), { status: 200 }))
      )
    const wrapper = mountConversation()
    await flushPromises()
    expect(wrapper.html()).toContain('Hello!')
    expect(wrapper.html()).toContain('How are you?')
  })

  it('has message input and send button', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockConvMeta }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockMessages }), { status: 200 }))
      )
    const wrapper = mountConversation()
    await flushPromises()
    expect(wrapper.find('input').exists()).toBe(true)
    expect(wrapper.find('button').exists()).toBe(true)
  })

  it('sends message when send button clicked', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockConvMeta }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockMessages }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(
          new Response(
            JSON.stringify({ data: { id: 'msg-3', content: 'New msg', sender_id: 'user-1', created_at: new Date().toISOString() } }),
            { status: 201 }
          )
        )
      )
    const wrapper = mountConversation()
    await flushPromises()
    await wrapper.find('input').setValue('New msg')
    // Conversation uses a button @click to send, not a form submit
    const buttons = wrapper.findAll('button')
    const sendBtn = buttons[buttons.length - 1] // last button is the Send button
    await sendBtn.trigger('click')
    await flushPromises()
    const calls = vi.mocked(fetch).mock.calls
    // Verify a POST call was made for sending (after 2 GET calls on mount)
    expect(calls.length).toBeGreaterThanOrEqual(3)
  })
})

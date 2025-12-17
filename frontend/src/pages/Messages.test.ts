import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import Messages from './Messages.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/messages', component: Messages },
    { path: '/messages/:id', component: { template: '<div>conversation</div>' } },
  ],
})

function mountMessages() {
  return mount(Messages, {
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
        MessageSquare: { template: '<span />' },
        Plus: { template: '<span />' },
      },
    },
  })
}

describe('Messages page', () => {
  beforeEach(async () => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
    await router.push('/messages')
  })

  it('fetches conversations on mount', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    mountMessages()
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/messages/conversations'),
      expect.anything()
    )
  })

  it('renders conversations list', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({
            data: [
              {
                id: 'conv-1',
                other_user: { id: 'user-2', username: 'bob', avatar: null },
                last_message: 'Hey there!',
                last_message_at: new Date().toISOString(),
                unread_count: 2,
              },
            ],
          }),
          { status: 200 }
        )
      )
    )
    const wrapper = mountMessages()
    await flushPromises()
    expect(wrapper.html()).toContain('bob')
    expect(wrapper.html()).toContain('Hey there!')
  })

  it('shows empty state when no conversations', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    const wrapper = mountMessages()
    await flushPromises()
    expect(wrapper.html()).toContain('message')
  })

  it('shows loading skeleton while fetching', () => {
    vi.mocked(fetch).mockImplementation(() => new Promise(() => {}))
    const wrapper = mountMessages()
    expect(wrapper.find('.skeleton').exists()).toBe(true)
  })

  it('has a button to start a new conversation', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    const wrapper = mountMessages()
    await flushPromises()
    const buttons = wrapper.findAll('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('shows unread count badge when unread_count > 0', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({
            data: [
              {
                id: 'conv-2',
                other_user: { id: 'u3', username: 'charlie', avatar: null },
                last_message: null,
                last_message_at: null,
                unread_count: 3,
              },
            ],
          }),
          { status: 200 }
        )
      )
    )
    const wrapper = mountMessages()
    await flushPromises()
    expect(wrapper.html()).toContain('3')
  })
})

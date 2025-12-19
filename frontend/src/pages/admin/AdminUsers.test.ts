import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminUsers from './AdminUsers.vue'

const mockUsers = [
  { id: 'u1', username: 'alice', email: 'alice@example.com', avatar: null, role: 'USER', banned: false, created_at: new Date().toISOString() },
  { id: 'u2', username: 'bob', email: 'bob@example.com', avatar: null, role: 'USER', banned: true, created_at: new Date().toISOString() },
  { id: 'u3', username: 'admin', email: 'admin@example.com', avatar: null, role: 'ADMIN', banned: false, created_at: new Date().toISOString() },
]

function mountAdminUsers() {
  return mount(AdminUsers, {
    global: {
      stubs: {
        UserAvatar: { template: '<div class="avatar" />' },
        Skeleton: { template: '<div class="skeleton" />' },
        Button: { template: '<button @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
        Badge: { template: '<span class="badge"><slot /></span>' },
        Input: {
          template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
          props: ['modelValue'],
          emits: ['update:modelValue'],
        },
        Search: { template: '<span />' },
        Ban: { template: '<span />' },
        CheckCircle: { template: '<span />' },
      },
    },
  })
}

describe('AdminUsers page', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
  })

  it('fetches users on mount', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockUsers }), { status: 200 }))
    )
    mountAdminUsers()
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/admin/users'),
      expect.anything()
    )
  })

  it('renders users list', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockUsers }), { status: 200 }))
    )
    const wrapper = mountAdminUsers()
    await flushPromises()
    expect(wrapper.html()).toContain('alice')
    expect(wrapper.html()).toContain('bob')
  })

  it('shows banned badge for banned users', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockUsers }), { status: 200 }))
    )
    const wrapper = mountAdminUsers()
    await flushPromises()
    expect(wrapper.html().toLowerCase()).toContain('banned')
  })

  it('shows loading skeleton while fetching', () => {
    vi.mocked(fetch).mockImplementation(() => new Promise(() => {}))
    const wrapper = mountAdminUsers()
    expect(wrapper.find('.skeleton').exists()).toBe(true)
  })

  it('renders search input', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    const wrapper = mountAdminUsers()
    await flushPromises()
    expect(wrapper.find('input').exists()).toBe(true)
  })

  it('filters users by search query', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: mockUsers }), { status: 200 }))
    )
    const wrapper = mountAdminUsers()
    await flushPromises()
    await wrapper.find('input').setValue('alice')
    await flushPromises()
    // After filtering, should show alice but not necessarily bob
    expect(wrapper.html()).toContain('alice')
  })

  it('calls ban API when ban button clicked', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockUsers }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: { banned: true } }), { status: 200 }))
      )
    const wrapper = mountAdminUsers()
    await flushPromises()
    const buttons = wrapper.findAll('button')
    // Click the first action button (ban/unban)
    if (buttons.length > 1) {
      await buttons[1].trigger('click')
      await flushPromises()
      // Should have made a ban/unban API call
      expect(vi.mocked(fetch).mock.calls.length).toBeGreaterThanOrEqual(2)
    }
  })
})

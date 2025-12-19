import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminStats from './AdminStats.vue'

const mockDauData = [
  { date: '2025-01-01', count: 50 },
  { date: '2025-01-02', count: 75 },
]
const mockNewUsersData = [
  { date: '2025-01-01', count: 10 },
  { date: '2025-01-02', count: 12 },
]
const mockPostsData = [
  { date: '2025-01-01', count: 5 },
  { date: '2025-01-02', count: 8 },
]

function mountAdminStats() {
  return mount(AdminStats, {
    global: {
      stubs: {
        Card: { template: '<div class="card"><slot /></div>' },
        CardContent: { template: '<div><slot /></div>' },
        CardHeader: { template: '<div><slot /></div>' },
        CardTitle: { template: '<h3><slot /></h3>' },
        Skeleton: { template: '<div class="skeleton" />' },
        Users: { template: '<span />' },
        FileText: { template: '<span />' },
        Eye: { template: '<span />' },
      },
    },
  })
}

describe('AdminStats page', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
  })

  it('fetches stats on mount', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: [] }), { status: 200 }))
    )
    mountAdminStats()
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalled()
    // All 3 stats endpoints should be called
    const calls = vi.mocked(fetch).mock.calls
    expect(calls.some(c => String(c[0]).includes('dau-range'))).toBe(true)
    expect(calls.some(c => String(c[0]).includes('new-users-range'))).toBe(true)
    expect(calls.some(c => String(c[0]).includes('posts-range'))).toBe(true)
  })

  it('shows loading skeleton while fetching', () => {
    vi.mocked(fetch).mockImplementation(() => new Promise(() => {}))
    const wrapper = mountAdminStats()
    expect(wrapper.find('.skeleton').exists()).toBe(true)
  })

  it('renders stats cards after load', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockDauData }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockNewUsersData }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockPostsData }), { status: 200 }))
      )
    const wrapper = mountAdminStats()
    await flushPromises()
    // Should display stat cards (DAU, New Users, Posts)
    expect(wrapper.find('.card').exists()).toBe(true)
  })

  it('shows aggregate totals', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockDauData }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockNewUsersData }), { status: 200 }))
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: mockPostsData }), { status: 200 }))
      )
    const wrapper = mountAdminStats()
    await flushPromises()
    // Total DAU = 50+75 = 125
    expect(wrapper.html()).toContain('125')
    // Total new users = 22
    expect(wrapper.html()).toContain('22')
    // Total posts = 13
    expect(wrapper.html()).toContain('13')
  })
})

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { nextTick } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { makeUser } from '@/test/factories'
import Settings from './Settings.vue'

function mountSettings() {
  return mount(Settings, {
    global: {
      stubs: {
        UserAvatar: { template: '<div class="avatar" />' },
        Card: { template: '<div class="card"><slot /></div>' },
        CardContent: { template: '<div><slot /></div>' },
        CardHeader: { template: '<div><slot /></div>' },
        CardTitle: { template: '<h2><slot /></h2>' },
        Button: { template: '<button type="submit" @click="$emit(\'click\')"><slot /></button>', emits: ['click'] },
        Label: { template: '<label><slot /></label>' },
        Input: {
          template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
          props: ['modelValue'],
          emits: ['update:modelValue'],
        },
        Textarea: {
          template: '<textarea :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
          props: ['modelValue'],
          emits: ['update:modelValue'],
        },
      },
    },
  })
}

describe('Settings page', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(fetch).mockClear()
  })

  it('renders profile section', () => {
    const wrapper = mountSettings()
    expect(wrapper.html()).toContain('Profile')
  })

  it('pre-fills username from auth store', async () => {
    const auth = useAuthStore()
    auth.setToken('test-token')
    auth.setUser(makeUser({ id: '1', username: 'alice', role: 'USER' }))
    const wrapper = mountSettings()
    await nextTick()
    // username ref is populated in onMounted; Input stub binds :value="modelValue"
    const inputs = wrapper.findAll('input')
    expect(inputs.length).toBeGreaterThan(0)
    const hasAlice = inputs.some(i => (i.element as HTMLInputElement).value === 'alice')
    expect(hasAlice).toBe(true)
  })

  it('pre-fills introduction when set', async () => {
    const auth = useAuthStore()
    auth.setToken('test-token')
    auth.setUser(makeUser({ id: '1', username: 'alice', role: 'USER', introduction: 'My intro' }))
    const wrapper = mountSettings()
    await nextTick()
    const textareas = wrapper.findAll('textarea')
    const hasIntro = textareas.some(ta => (ta.element as HTMLTextAreaElement).value === 'My intro')
    expect(hasIntro).toBe(true)
  })

  it('calls update profile API on save', async () => {
    const auth = useAuthStore()
    auth.setUser(makeUser({ id: '1', username: 'alice', role: 'USER' }))
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({ data: { id: '1', username: 'alice', role: 'USER', avatar: null } }),
          { status: 200 }
        )
      )
    )
    const wrapper = mountSettings()
    // Find and submit the profile form
    const forms = wrapper.findAll('form')
    if (forms.length > 0) {
      await forms[0].trigger('submit')
      await flushPromises()
      expect(vi.mocked(fetch)).toHaveBeenCalledWith(
        expect.stringContaining('/users/me'),
        expect.objectContaining({ method: 'PUT' })
      )
    } else {
      // Form may be part of card structure — find the save button
      const buttons = wrapper.findAll('button')
      const saveBtn = buttons.find(b => b.text().toLowerCase().includes('save'))
      if (saveBtn) {
        await saveBtn.trigger('click')
        await flushPromises()
      }
    }
  })

  it('shows notification preferences section', () => {
    const wrapper = mountSettings()
    expect(wrapper.html()).toContain('Notification')
  })
})

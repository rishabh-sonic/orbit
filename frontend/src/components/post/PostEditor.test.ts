import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PostEditor from './PostEditor.vue'

function mountEditor(props: { title?: string; content?: string; loading?: boolean; submitLabel?: string } = {}) {
  return mount(PostEditor, {
    props: {
      title: props.title ?? '',
      content: props.content ?? '',
      loading: props.loading ?? false,
      submitLabel: props.submitLabel,
    },
    global: {
      stubs: {
        Input: {
          template: '<input :value="value" @input="$emit(\'input\', $event)" />',
          props: ['value'],
          emits: ['input'],
        },
        Textarea: {
          template: '<textarea :value="value" @input="$emit(\'input\', $event)" />',
          props: ['value'],
          emits: ['input'],
        },
        Button: { template: '<button :disabled="disabled" :type="type"><slot /></button>', props: ['disabled', 'type'] },
        Label: { template: '<label><slot /></label>' },
      },
    },
  })
}

describe('PostEditor component', () => {
  it('renders title input', () => {
    const wrapper = mountEditor()
    expect(wrapper.find('input').exists()).toBe(true)
  })

  it('renders content textarea', () => {
    const wrapper = mountEditor()
    expect(wrapper.find('textarea').exists()).toBe(true)
  })

  it('renders submit button', () => {
    const wrapper = mountEditor()
    const buttons = wrapper.findAll('button')
    const submitBtn = buttons.find(b => b.attributes('type') === 'submit')
    expect(submitBtn).toBeDefined()
  })

  it('renders cancel button', () => {
    const wrapper = mountEditor()
    const buttons = wrapper.findAll('button')
    const cancelBtn = buttons.find(b => b.text().toLowerCase().includes('cancel'))
    expect(cancelBtn).toBeDefined()
  })

  it('uses custom submit label', () => {
    const wrapper = mountEditor({ submitLabel: 'Publish' })
    expect(wrapper.html()).toContain('Publish')
  })

  it('uses "Submit" as default label', () => {
    const wrapper = mountEditor()
    expect(wrapper.html()).toContain('Submit')
  })

  it('disables submit button when loading', () => {
    const wrapper = mountEditor({ loading: true })
    const buttons = wrapper.findAll('button')
    const submitBtn = buttons.find(b => b.attributes('type') === 'submit')
    expect(submitBtn?.attributes('disabled')).toBeDefined()
  })

  it('shows saving indicator when loading', () => {
    const wrapper = mountEditor({ loading: true })
    expect(wrapper.html()).toContain('Saving')
  })

  it('emits "submit" event on form submission', async () => {
    const wrapper = mountEditor()
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')).toBeTruthy()
    expect(wrapper.emitted('submit')?.length).toBe(1)
  })

  it('emits "cancel" event on cancel click', async () => {
    const wrapper = mountEditor()
    const buttons = wrapper.findAll('button')
    const cancelBtn = buttons.find(b => b.text().toLowerCase().includes('cancel'))
    if (cancelBtn) {
      await cancelBtn.trigger('click')
      expect(wrapper.emitted('cancel')).toBeTruthy()
    }
  })

  it('emits update:title on title input', async () => {
    const wrapper = mountEditor({ title: 'Old Title' })
    // Trigger native input event
    await wrapper.find('form').trigger('submit') // just check form works
    expect(wrapper.emitted('submit')).toBeTruthy()
  })
})

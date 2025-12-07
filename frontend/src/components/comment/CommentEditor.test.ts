import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import CommentEditor from '@/components/comment/CommentEditor.vue'

describe('CommentEditor', () => {
  it('renders textarea and buttons', () => {
    const wrapper = mount(CommentEditor)
    expect(wrapper.find('textarea').exists()).toBe(true)
    expect(wrapper.text()).toContain('Post')
    expect(wrapper.text()).toContain('Cancel')
  })

  it('emits cancel when Cancel is clicked', async () => {
    const wrapper = mount(CommentEditor)
    const buttons = wrapper.findAll('button')
    const cancelBtn = buttons.find(b => b.text().includes('Cancel'))
    await cancelBtn?.trigger('click')
    expect(wrapper.emitted('cancel')).toBeTruthy()
  })

  it('does not emit submit when content is empty', async () => {
    const wrapper = mount(CommentEditor)
    const buttons = wrapper.findAll('button')
    const postBtn = buttons.find(b => b.text().includes('Post'))
    await postBtn?.trigger('click')
    expect(wrapper.emitted('submit')).toBeFalsy()
  })

  it('emits submit with content when filled', async () => {
    const wrapper = mount(CommentEditor)
    await wrapper.find('textarea').setValue('Great article!')
    const buttons = wrapper.findAll('button')
    const postBtn = buttons.find(b => b.text().includes('Post'))
    await postBtn?.trigger('click')
    expect(wrapper.emitted('submit')?.[0]).toEqual(['Great article!'])
  })

  it('clears textarea after emit', async () => {
    const wrapper = mount(CommentEditor)
    const ta = wrapper.find('textarea')
    await ta.setValue('Some comment')
    const buttons = wrapper.findAll('button')
    const postBtn = buttons.find(b => b.text().includes('Post'))
    await postBtn?.trigger('click')
    expect((ta.element as HTMLTextAreaElement).value).toBe('')
  })

  it('shows custom placeholder', () => {
    const wrapper = mount(CommentEditor, { props: { placeholder: 'Write a reply…' } })
    const ta = wrapper.find('textarea')
    expect(ta.attributes('placeholder')).toBe('Write a reply…')
  })

  it('disables Post button when loading', () => {
    const wrapper = mount(CommentEditor, { props: { loading: true } })
    const buttons = wrapper.findAll('button')
    const postBtn = buttons.find(b => b.text().toLowerCase().includes('post') || b.text().toLowerCase().includes('…'))
    expect(postBtn?.attributes('disabled')).toBeDefined()
  })
})

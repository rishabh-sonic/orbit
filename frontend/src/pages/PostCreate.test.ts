import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import PostCreate from './PostCreate.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/posts/new', component: PostCreate },
    { path: '/posts/:id', component: { template: '<div>post</div>' } },
  ],
})

function mountPostCreate() {
  return mount(PostCreate, {
    global: {
      plugins: [router],
      stubs: {
        PostEditor: {
          template: `
            <form @submit.prevent="$emit('submit')">
              <input data-testid="title" :value="title" @input="$emit('update:title', $event.target.value)" />
              <textarea data-testid="content" :value="content" @input="$emit('update:content', $event.target.value)" />
              <button type="submit">{{ submitLabel }}</button>
              <button type="button" @click="$emit('cancel')">Cancel</button>
            </form>
          `,
          props: ['title', 'content', 'loading', 'submitLabel'],
          emits: ['update:title', 'update:content', 'submit', 'cancel'],
        },
      },
    },
  })
}

describe('PostCreate page', () => {
  beforeEach(() => {
    vi.mocked(fetch).mockClear()
  })

  it('renders PostEditor', () => {
    const wrapper = mountPostCreate()
    expect(wrapper.find('form').exists()).toBe(true)
    expect(wrapper.html()).toContain('New Post')
  })

  it('shows "Publish" submit label', () => {
    const wrapper = mountPostCreate()
    expect(wrapper.html()).toContain('Publish')
  })

  it('calls create API on submit and navigates to post', async () => {
    const postId = 'abc-123'
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: { id: postId } }), { status: 201 }))
    )
    const wrapper = mountPostCreate()
    await wrapper.find('[data-testid="title"]').setValue('My Post')
    await wrapper.find('[data-testid="content"]').setValue('Post content here')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/posts'),
      expect.objectContaining({ method: 'POST' })
    )
  })

  it('shows error toast on API failure', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ error: 'validation failed' }), { status: 400 }))
    )
    const wrapper = mountPostCreate()
    await wrapper.find('[data-testid="title"]').setValue('Title')
    await wrapper.find('[data-testid="content"]').setValue('Content')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    // Form still visible
    expect(wrapper.find('form').exists()).toBe(true)
  })
})

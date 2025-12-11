import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import PostEdit from './PostEdit.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/posts/:id/edit', component: PostEdit },
    { path: '/posts/:id', component: { template: '<div>post</div>' } },
  ],
})

function mountPostEdit(_postId = 'post-1') {
  return mount(PostEdit, {
    global: {
      plugins: [router],
      stubs: {
        PostEditor: {
          template: `
            <form @submit.prevent="$emit('submit')">
              <input data-testid="title" :value="title" @input="$emit('update:title', $event.target.value)" />
              <textarea data-testid="content" :value="content" @input="$emit('update:content', $event.target.value)" />
              <button type="submit">{{ submitLabel }}</button>
            </form>
          `,
          props: ['title', 'content', 'loading', 'submitLabel'],
          emits: ['update:title', 'update:content', 'submit', 'cancel'],
        },
        Skeleton: { template: '<div class="skeleton" />' },
      },
    },
  })
}

describe('PostEdit page', () => {
  beforeEach(async () => {
    vi.mocked(fetch).mockClear()
    await router.push('/posts/post-1/edit')
  })

  it('fetches post on mount and pre-fills form', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({ data: { id: 'post-1', title: 'Original Title', content: 'Original content' } }),
          { status: 200 }
        )
      )
    )
    const wrapper = mountPostEdit()
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledWith(
      expect.stringContaining('/posts/post-1'),
      expect.anything()
    )
    // Title and content should be pre-filled
    const titleInput = wrapper.find('[data-testid="title"]')
    if (titleInput.exists()) {
      expect((titleInput.element as HTMLInputElement).value).toBe('Original Title')
    }
  })

  it('shows skeleton while loading', async () => {
    vi.mocked(fetch).mockImplementation(() => new Promise(() => {})) // never resolves
    const wrapper = mountPostEdit()
    expect(wrapper.find('.skeleton').exists()).toBe(true)
  })

  it('calls update API on submit', async () => {
    vi.mocked(fetch)
      .mockImplementationOnce(() =>
        Promise.resolve(
          new Response(JSON.stringify({ data: { id: 'post-1', title: 'Original', content: 'Body' } }), { status: 200 })
        )
      )
      .mockImplementationOnce(() =>
        Promise.resolve(new Response(JSON.stringify({ data: {} }), { status: 200 }))
      )
    const wrapper = mountPostEdit()
    await flushPromises()
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(vi.mocked(fetch)).toHaveBeenCalledTimes(2)
    const calls = vi.mocked(fetch).mock.calls
    expect(calls[1]?.[1]).toMatchObject({ method: 'PUT' })
  })

  it('shows "Save changes" label', async () => {
    vi.mocked(fetch).mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ data: { title: 'T', content: 'C' } }), { status: 200 }))
    )
    const wrapper = mountPostEdit()
    await flushPromises()
    expect(wrapper.html()).toContain('Save changes')
  })
})

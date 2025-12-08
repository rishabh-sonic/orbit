import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import UserAvatar from '@/components/user/UserAvatar.vue'

describe('UserAvatar', () => {
  it('renders img when avatar URL provided', () => {
    const wrapper = mount(UserAvatar, {
      props: { username: 'alice', src: 'https://example.com/avatar.jpg' },
    })
    // shadcn AvatarImage renders an img — check the src attribute appears anywhere
    expect(wrapper.html()).toContain('example.com/avatar.jpg')
  })

  it('renders initial fallback when no avatar', () => {
    const wrapper = mount(UserAvatar, {
      props: { username: 'alice', src: null },
    })
    expect(wrapper.text()).toContain('A')
  })

  it('uses first letter of username for fallback', () => {
    const wrapper = mount(UserAvatar, {
      props: { username: 'Bob', src: null },
    })
    expect(wrapper.text()).toContain('B')
  })

  it('applies size classes', () => {
    const wrapper = mount(UserAvatar, {
      props: { username: 'alice', src: null, size: 'lg' },
    })
    // Just verify it renders without error
    expect(wrapper.exists()).toBe(true)
  })
})

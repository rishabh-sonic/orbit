import type { User } from '@/stores/auth'
import type { Post } from '@/components/post/PostCard.vue'

/** Build a full User object, merging any overrides. */
export function makeUser(overrides: Partial<User> & Pick<User, 'id' | 'username' | 'role'>): User {
  return {
    email: `${overrides.username}@example.com`,
    avatar: null,
    introduction: null,
    verified: true,
    banned: false,
    created_at: new Date().toISOString(),
    ...overrides,
  }
}

/** Build a full Post object, merging any overrides. */
export function makePost(overrides: Partial<Post> & Pick<Post, 'id' | 'title'>): Post {
  return {
    content: 'Post content',
    author: { id: 'author-1', username: 'alice', avatar: null },
    comment_count: 0,
    views: 0,
    status: 'open',
    pinned_at: null,
    created_at: new Date().toISOString(),
    ...overrides,
  }
}

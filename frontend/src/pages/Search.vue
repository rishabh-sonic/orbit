<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import UserAvatar from '@/components/user/UserAvatar.vue'
import PostCard, { type Post } from '@/components/post/PostCard.vue'
import { Search as SearchIcon } from 'lucide-vue-next'

interface UserResult {
  id: string
  username: string
  avatar: string | null
  introduction: string | null
}

interface SearchResults {
  posts: Post[]
  users: UserResult[]
}

const route = useRoute()
const router = useRouter()

const query = ref((route.query.q as string) ?? '')
const results = ref<SearchResults>({ posts: [], users: [] })
const loading = ref(false)

onMounted(() => { if (query.value) doSearch() })
watch(() => route.query.q, (q) => { query.value = q as string ?? ''; if (query.value) doSearch() })

async function onSearch() {
  if (!query.value.trim()) return
  router.replace({ path: '/search', query: { q: query.value } })
  doSearch()
}

async function doSearch() {
  if (!query.value.trim()) return
  loading.value = true
  try {
    const data = await api.get<SearchResults>(`/search/global?q=${encodeURIComponent(query.value)}`)
    results.value = data
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-4">Search</h1>

    <form @submit.prevent="onSearch" class="flex gap-2 mb-6">
      <div class="relative flex-1">
        <SearchIcon class="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input v-model="query" placeholder="Search posts and users…" class="pl-9" />
      </div>
      <Button type="submit" :disabled="loading">Search</Button>
    </form>

    <template v-if="loading">
      <div class="space-y-3">
        <Skeleton v-for="i in 4" :key="i" class="h-20 w-full rounded-lg" />
      </div>
    </template>

    <template v-else-if="query">
      <Tabs default-value="posts">
        <TabsList class="mb-4">
          <TabsTrigger value="posts">Posts ({{ results.posts.length }})</TabsTrigger>
          <TabsTrigger value="users">Users ({{ results.users.length }})</TabsTrigger>
        </TabsList>

        <TabsContent value="posts">
          <div v-if="results.posts.length === 0" class="text-muted-foreground text-sm py-8 text-center">
            No posts found for "{{ query }}".
          </div>
          <div v-else class="space-y-3">
            <PostCard v-for="post in results.posts" :key="post.id" :post="post" />
          </div>
        </TabsContent>

        <TabsContent value="users">
          <div v-if="results.users.length === 0" class="text-muted-foreground text-sm py-8 text-center">
            No users found for "{{ query }}".
          </div>
          <div v-else class="space-y-2">
            <RouterLink
              v-for="user in results.users"
              :key="user.id"
              :to="`/users/${user.username}`"
              class="flex items-center gap-3 p-3 rounded-lg border hover:border-muted-foreground/30 transition-colors"
            >
              <UserAvatar :src="user.avatar" :username="user.username" size="md" />
              <div>
                <p class="font-medium text-sm">{{ user.username }}</p>
                <p v-if="user.introduction" class="text-xs text-muted-foreground line-clamp-1">
                  {{ user.introduction }}
                </p>
              </div>
            </RouterLink>
          </div>
        </TabsContent>
      </Tabs>
    </template>

    <template v-else>
      <p class="text-center text-muted-foreground py-12">Enter a query to search posts and users.</p>
    </template>
  </div>
</template>

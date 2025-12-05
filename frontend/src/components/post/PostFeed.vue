<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '@/lib/api'
import PostCard, { type Post } from './PostCard.vue'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

const props = defineProps<{
  endpoint?: string
  userId?: string
}>()

const posts = ref<Post[]>([])
const loading = ref(false)
const page = ref(1)
const hasMore = ref(true)
const limit = 20

async function load(reset = false) {
  if (loading.value) return
  loading.value = true
  try {
    const ep = props.endpoint ?? '/posts'
    const params = new URLSearchParams({ page: String(reset ? 1 : page.value), limit: String(limit) })
    if (props.userId) params.set('user_id', props.userId)
    const data = await api.get<Post[]>(`${ep}?${params}`)
    if (reset) {
      posts.value = data
      page.value = 1
    } else {
      posts.value.push(...data)
    }
    hasMore.value = data.length === limit
    page.value++
  } finally {
    loading.value = false
  }
}

function onDeleted(id: string) {
  posts.value = posts.value.filter(p => p.id !== id)
}

onMounted(() => load(true))
watch(() => props.endpoint, () => load(true))
watch(() => props.userId, () => load(true))
</script>

<template>
  <div class="space-y-3">
    <template v-if="loading && posts.length === 0">
      <div v-for="i in 5" :key="i" class="border rounded-lg p-4 space-y-2">
        <Skeleton class="h-4 w-3/4" />
        <Skeleton class="h-3 w-full" />
        <Skeleton class="h-3 w-1/2" />
      </div>
    </template>

    <template v-else-if="posts.length === 0">
      <p class="text-center text-muted-foreground py-12">No posts yet.</p>
    </template>

    <template v-else>
      <PostCard
        v-for="post in posts"
        :key="post.id"
        :post="post"
        @deleted="onDeleted"
      />
      <div class="text-center pt-2">
        <Button
          v-if="hasMore"
          variant="outline"
          size="sm"
          :disabled="loading"
          @click="load()"
        >
          {{ loading ? 'Loading…' : 'Load more' }}
        </Button>
        <p v-else class="text-xs text-muted-foreground">All posts loaded</p>
      </div>
    </template>
  </div>
</template>

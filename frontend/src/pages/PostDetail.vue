<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import UserAvatar from '@/components/user/UserAvatar.vue'
import CommentList from '@/components/comment/CommentList.vue'
import { Lock, Pin, Pencil, Trash2, BookmarkPlus, BookmarkMinus } from 'lucide-vue-next'
import type { Post } from '@/components/post/PostCard.vue'

interface PostDetail extends Post {
  subscribed?: boolean
}

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { toast } = useToast()

const post = ref<PostDetail | null>(null)
const loading = ref(true)
const subscribed = ref(false)

const canEdit = computed(() => auth.user?.id === post.value?.author.id || auth.isAdmin)

onMounted(async () => {
  try {
    const data = await api.get<PostDetail>(`/posts/${route.params.id}`)
    post.value = data
    subscribed.value = !!data.subscribed
  } catch {
    toast({ variant: 'destructive', title: 'Failed to load post' })
  } finally {
    loading.value = false
  }
})

async function deletePost() {
  if (!confirm('Delete this post?')) return
  try {
    await api.delete(`/posts/${post.value!.id}`)
    router.push('/')
  } catch {
    toast({ variant: 'destructive', title: 'Failed to delete post' })
  }
}

async function toggleSubscription() {
  if (!post.value) return
  const id = post.value.id
  try {
    if (subscribed.value) {
      await api.delete(`/posts/${id}/subscribe`)
    } else {
      await api.post(`/posts/${id}/subscribe`)
    }
    subscribed.value = !subscribed.value
  } catch {
    toast({ variant: 'destructive', title: 'Failed to update subscription' })
  }
}

function formatDate(d: string) {
  return new Intl.DateTimeFormat('en', { dateStyle: 'long' }).format(new Date(d))
}
</script>

<template>
  <div>
    <template v-if="loading">
      <div class="space-y-3">
        <Skeleton class="h-8 w-3/4" />
        <Skeleton class="h-4 w-40" />
        <Skeleton class="h-4 w-full" />
        <Skeleton class="h-4 w-full" />
        <Skeleton class="h-4 w-2/3" />
      </div>
    </template>

    <template v-else-if="post">
      <!-- Header -->
      <div class="mb-6">
        <div class="flex items-start gap-2 mb-2">
          <h1 class="text-2xl font-bold flex-1">{{ post.title }}</h1>
          <div class="flex items-center gap-1 shrink-0 mt-1">
            <Badge v-if="post.pinned_at" variant="secondary" class="gap-1">
              <Pin class="h-3 w-3" /> Pinned
            </Badge>
            <Badge v-if="post.status === 'closed'" variant="outline" class="gap-1">
              <Lock class="h-3 w-3" /> Closed
            </Badge>
          </div>
        </div>

        <div class="flex items-center gap-2 text-sm text-muted-foreground flex-wrap">
          <RouterLink :to="`/users/${post.author.username}`" class="flex items-center gap-1.5 hover:underline">
            <UserAvatar :src="post.author.avatar" :username="post.author.username" size="sm" />
            <span class="font-medium text-foreground">{{ post.author.username }}</span>
          </RouterLink>
          <span>&middot;</span>
          <span>{{ formatDate(post.created_at) }}</span>
          <span>&middot;</span>
          <span>{{ post.views }} views</span>

          <div class="ml-auto flex items-center gap-1">
            <Button
              v-if="auth.isLoggedIn"
              variant="ghost"
              size="sm"
              @click="toggleSubscription"
            >
              <BookmarkMinus v-if="subscribed" class="h-4 w-4 mr-1" />
              <BookmarkPlus v-else class="h-4 w-4 mr-1" />
              {{ subscribed ? 'Unsubscribe' : 'Subscribe' }}
            </Button>
            <Button v-if="canEdit" variant="ghost" size="icon" as-child>
              <RouterLink :to="`/posts/${post.id}/edit`">
                <Pencil class="h-4 w-4" />
              </RouterLink>
            </Button>
            <Button v-if="canEdit" variant="ghost" size="icon" class="text-destructive" @click="deletePost">
              <Trash2 class="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      <!-- Body -->
      <div class="prose prose-sm max-w-none mb-8 whitespace-pre-wrap break-words border rounded-lg p-4">
        {{ post.content }}
      </div>

      <!-- Comments -->
      <CommentList :post-id="post.id" :closed="post.status === 'closed'" />
    </template>

    <template v-else>
      <p class="text-muted-foreground text-center py-12">Post not found.</p>
    </template>
  </div>
</template>

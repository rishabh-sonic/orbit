<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import CommentItem, { type Comment } from './CommentItem.vue'
import CommentEditor from './CommentEditor.vue'
import { Skeleton } from '@/components/ui/skeleton'

const props = defineProps<{ postId: string; closed?: boolean }>()

const auth = useAuthStore()
const { toast } = useToast()
const comments = ref<Comment[]>([])
const loading = ref(false)
const postLoading = ref(false)

onMounted(load)

async function load() {
  loading.value = true
  try {
    const data = await api.get<Comment[]>(`/posts/${props.postId}/comments`)
    comments.value = data
  } finally {
    loading.value = false
  }
}

async function postComment(content: string) {
  postLoading.value = true
  try {
    const data = await api.post<Comment>(`/posts/${props.postId}/comments`, { content })
    comments.value.push(data)
  } catch {
    toast({ variant: 'destructive', title: 'Failed to post comment' })
  } finally {
    postLoading.value = false
  }
}

function onDeleted(id: string) {
  function removeById(list: Comment[]): Comment[] {
    return list
      .filter(c => c.id !== id)
      .map(c => ({ ...c, replies: c.replies ? removeById(c.replies) : undefined }))
  }
  comments.value = removeById(comments.value)
}

function onReplied(_comment: Comment) {
  load()
}
</script>

<template>
  <div class="space-y-4">
    <h2 class="font-semibold text-lg">Comments ({{ comments.length }})</h2>

    <template v-if="loading">
      <div v-for="i in 3" :key="i" class="flex gap-2">
        <Skeleton class="h-8 w-8 rounded-full shrink-0" />
        <div class="flex-1 space-y-1">
          <Skeleton class="h-3 w-24" />
          <Skeleton class="h-3 w-full" />
          <Skeleton class="h-3 w-3/4" />
        </div>
      </div>
    </template>

    <template v-else>
      <CommentItem
        v-for="comment in comments"
        :key="comment.id"
        :comment="comment"
        :post-id="postId"
        :depth="0"
        @deleted="onDeleted"
        @replied="onReplied"
      />
      <p v-if="comments.length === 0" class="text-sm text-muted-foreground py-4 text-center">
        No comments yet. Be the first!
      </p>
    </template>

    <div v-if="auth.isLoggedIn && !closed" class="pt-2 border-t">
      <p class="text-sm font-medium mb-2">Add a comment</p>
      <CommentEditor :loading="postLoading" @submit="postComment" @cancel="() => {}" />
    </div>
    <p v-else-if="!auth.isLoggedIn" class="text-sm text-muted-foreground text-center py-2">
      <RouterLink to="/login" class="underline">Sign in</RouterLink> to comment.
    </p>
    <p v-else-if="closed" class="text-sm text-muted-foreground text-center py-2">
      This post is closed for comments.
    </p>
  </div>
</template>

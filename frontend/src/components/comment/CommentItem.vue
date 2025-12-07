<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import UserAvatar from '@/components/user/UserAvatar.vue'
import CommentEditor from './CommentEditor.vue'
import { MessageSquare, Trash2, Pin } from 'lucide-vue-next'

export interface Comment {
  id: string
  content: string
  pinned_at: string | null
  created_at: string
  author: {
    id: string
    username: string
    avatar: string | null
  }
  replies?: Comment[]
}

const props = defineProps<{
  comment: Comment
  postId: string
  depth?: number
}>()

const emit = defineEmits<{
  deleted: [id: string]
  replied: [comment: Comment]
}>()

const auth = useAuthStore()
const { toast } = useToast()
const showReply = ref(false)
const replyLoading = ref(false)

const canDelete = computed(() => auth.user?.id === props.comment.author.id || auth.isAdmin)

async function postReply(content: string) {
  replyLoading.value = true
  try {
    const data = await api.post<Comment>(`/posts/${props.postId}/comments`, {
      content,
      parent_id: props.comment.id,
    })
    emit('replied', data)
    showReply.value = false
  } catch {
    toast({ variant: 'destructive', title: 'Failed to post reply' })
  } finally {
    replyLoading.value = false
  }
}

async function deleteComment() {
  if (!confirm('Delete this comment?')) return
  try {
    await api.delete(`/comments/${props.comment.id}`)
    emit('deleted', props.comment.id)
  } catch {
    toast({ variant: 'destructive', title: 'Failed to delete comment' })
  }
}

function formatDate(d: string) {
  return new Intl.DateTimeFormat('en', { month: 'short', day: 'numeric', year: 'numeric' }).format(new Date(d))
}
</script>

<template>
  <div :class="['space-y-2', depth && depth > 0 ? 'pl-6 border-l ml-2' : '']">
    <div class="flex gap-2">
      <RouterLink :to="`/users/${comment.author.username}`" class="shrink-0">
        <UserAvatar :src="comment.author.avatar" :username="comment.author.username" size="sm" class="mt-0.5" />
      </RouterLink>
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 flex-wrap">
          <RouterLink :to="`/users/${comment.author.username}`" class="text-sm font-medium hover:underline">
            {{ comment.author.username }}
          </RouterLink>
          <span class="text-xs text-muted-foreground">{{ formatDate(comment.created_at) }}</span>
          <Badge v-if="comment.pinned_at" variant="secondary" class="h-4 px-1.5 text-[10px]">
            <Pin class="h-2.5 w-2.5 mr-0.5" /> Pinned
          </Badge>
        </div>
        <p class="text-sm mt-1 whitespace-pre-wrap break-words">{{ comment.content }}</p>
        <div class="flex items-center gap-2 mt-1">
          <Button
            v-if="auth.isLoggedIn && (depth ?? 0) < 3"
            variant="ghost"
            size="sm"
            class="h-6 px-2 text-xs text-muted-foreground"
            @click="showReply = !showReply"
          >
            <MessageSquare class="h-3 w-3 mr-1" /> Reply
          </Button>
          <Button
            v-if="canDelete"
            variant="ghost"
            size="sm"
            class="h-6 px-2 text-xs text-destructive hover:text-destructive"
            @click="deleteComment"
          >
            <Trash2 class="h-3 w-3" />
          </Button>
        </div>

        <div v-if="showReply" class="mt-2">
          <CommentEditor
            :loading="replyLoading"
            :autofocus="true"
            placeholder="Write a reply…"
            @submit="postReply"
            @cancel="showReply = false"
          />
        </div>
      </div>
    </div>

    <!-- Nested replies -->
    <template v-if="comment.replies?.length">
      <CommentItem
        v-for="reply in comment.replies"
        :key="reply.id"
        :comment="reply"
        :post-id="postId"
        :depth="(depth ?? 0) + 1"
        @deleted="emit('deleted', $event)"
        @replied="emit('replied', $event)"
      />
    </template>
  </div>
</template>

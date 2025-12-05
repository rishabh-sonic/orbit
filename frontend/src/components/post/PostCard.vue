<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import UserAvatar from '@/components/user/UserAvatar.vue'
import { MessageSquare, MoreHorizontal, Pin, Lock, Trash2, Pencil } from 'lucide-vue-next'

export interface Post {
  id: string
  title: string
  content: string
  status: 'open' | 'closed'
  pinned_at: string | null
  views: number
  comment_count: number
  created_at: string
  author: {
    id: string
    username: string
    avatar: string | null
  }
}

const props = defineProps<{ post: Post }>()
const emit = defineEmits<{ deleted: [id: string] }>()

const auth = useAuthStore()
const { toast } = useToast()

const canModerate = computed(() => auth.isAdmin)
const isAuthor = computed(() => auth.user?.id === props.post.author.id)
const canEdit = computed(() => isAuthor.value || canModerate.value)

const excerpt = computed(() => {
  const text = props.post.content.replace(/[#*`_~\[\]]/g, '')
  return text.length > 160 ? text.slice(0, 160) + '…' : text
})

async function deletePost() {
  if (!confirm('Delete this post?')) return
  try {
    await api.delete(`/posts/${props.post.id}`)
    emit('deleted', props.post.id)
  } catch {
    toast({ variant: 'destructive', title: 'Failed to delete post' })
  }
}

function formatDate(d: string) {
  return new Intl.DateTimeFormat('en', { month: 'short', day: 'numeric' }).format(new Date(d))
}
</script>

<template>
  <div class="border rounded-lg p-4 hover:border-muted-foreground/30 transition-colors bg-card">
    <div class="flex items-start gap-3">
      <RouterLink :to="`/users/${post.author.username}`" class="shrink-0 mt-0.5">
        <UserAvatar :src="post.author.avatar" :username="post.author.username" size="md" />
      </RouterLink>

      <div class="flex-1 min-w-0">
        <div class="flex items-start gap-2 flex-wrap">
          <RouterLink :to="`/posts/${post.id}`" class="font-semibold hover:underline line-clamp-2 flex-1">
            {{ post.title }}
          </RouterLink>
          <div class="flex items-center gap-1 shrink-0">
            <Badge v-if="post.pinned_at" variant="secondary" class="h-5 px-1.5 text-[10px]">
              <Pin class="h-2.5 w-2.5 mr-0.5" /> Pinned
            </Badge>
            <Badge v-if="post.status === 'closed'" variant="outline" class="h-5 px-1.5 text-[10px]">
              <Lock class="h-2.5 w-2.5 mr-0.5" /> Closed
            </Badge>
          </div>
        </div>

        <p class="text-sm text-muted-foreground mt-1 line-clamp-2">{{ excerpt }}</p>

        <div class="flex items-center gap-3 mt-2 text-xs text-muted-foreground">
          <RouterLink :to="`/users/${post.author.username}`" class="hover:underline font-medium text-foreground/70">
            {{ post.author.username }}
          </RouterLink>
          <span>{{ formatDate(post.created_at) }}</span>
          <span class="flex items-center gap-0.5">
            <MessageSquare class="h-3 w-3" /> {{ post.comment_count }}
          </span>
          <span>{{ post.views }} views</span>

          <DropdownMenu v-if="canEdit">
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="icon" class="h-5 w-5 ml-auto">
                <MoreHorizontal class="h-3.5 w-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem as-child>
                <RouterLink :to="`/posts/${post.id}/edit`" class="flex items-center gap-2">
                  <Pencil class="h-3.5 w-3.5" /> Edit
                </RouterLink>
              </DropdownMenuItem>
              <DropdownMenuItem
                class="text-destructive cursor-pointer flex items-center gap-2"
                @click="deletePost"
              >
                <Trash2 class="h-3.5 w-3.5" /> Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  </div>
</template>

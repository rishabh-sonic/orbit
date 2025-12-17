<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import UserAvatar from '@/components/user/UserAvatar.vue'
import { MessageSquare, Plus } from 'lucide-vue-next'

interface Conversation {
  id: string
  other_user: {
    id: string
    username: string
    avatar: string | null
  }
  last_message: string | null
  last_message_at: string | null
  unread_count: number
}

const { toast } = useToast()
const conversations = ref<Conversation[]>([])
const loading = ref(true)
const newUsername = ref('')
const newLoading = ref(false)
const showNew = ref(false)

onMounted(load)

async function load() {
  try {
    const data = await api.get<Conversation[]>('/messages/conversations')
    conversations.value = data
  } catch {
    toast({ variant: 'destructive', title: 'Failed to load conversations' })
  } finally {
    loading.value = false
  }
}

async function startConversation() {
  if (!newUsername.value.trim()) return
  newLoading.value = true
  try {
    const data = await api.post<Conversation>('/messages/conversations', {
      username: newUsername.value.trim(),
    })
    conversations.value.unshift(data)
    newUsername.value = ''
    showNew.value = false
  } catch {
    toast({ variant: 'destructive', title: 'User not found or error occurred' })
  } finally {
    newLoading.value = false
  }
}

function formatTime(d: string | null) {
  if (!d) return ''
  const date = new Date(d)
  const now = new Date()
  if (date.toDateString() === now.toDateString()) {
    return new Intl.DateTimeFormat('en', { hour: 'numeric', minute: '2-digit' }).format(date)
  }
  return new Intl.DateTimeFormat('en', { month: 'short', day: 'numeric' }).format(date)
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Messages</h1>
      <Button size="sm" @click="showNew = !showNew">
        <Plus class="h-4 w-4 mr-1" /> New
      </Button>
    </div>

    <div v-if="showNew" class="mb-4 flex gap-2">
      <Input
        v-model="newUsername"
        placeholder="Username to message"
        @keydown.enter="startConversation"
      />
      <Button :disabled="newLoading" @click="startConversation">
        {{ newLoading ? '…' : 'Start' }}
      </Button>
    </div>

    <template v-if="loading">
      <div v-for="i in 4" :key="i" class="flex gap-3 p-3 border-b">
        <Skeleton class="h-10 w-10 rounded-full shrink-0" />
        <div class="flex-1 space-y-1.5">
          <Skeleton class="h-4 w-24" />
          <Skeleton class="h-3 w-48" />
        </div>
      </div>
    </template>

    <template v-else-if="conversations.length === 0">
      <div class="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
        <MessageSquare class="h-10 w-10 opacity-30" />
        <p>No conversations yet. Start one above.</p>
      </div>
    </template>

    <div v-else class="border rounded-lg divide-y overflow-hidden">
      <RouterLink
        v-for="conv in conversations"
        :key="conv.id"
        :to="`/messages/${conv.id}`"
        class="flex items-center gap-3 p-3 hover:bg-accent transition-colors"
      >
        <UserAvatar :src="conv.other_user.avatar" :username="conv.other_user.username" size="md" />
        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between">
            <span class="font-medium text-sm">{{ conv.other_user.username }}</span>
            <span class="text-xs text-muted-foreground">{{ formatTime(conv.last_message_at) }}</span>
          </div>
          <p class="text-xs text-muted-foreground truncate">
            {{ conv.last_message ?? 'Start a conversation' }}
          </p>
        </div>
        <div
          v-if="conv.unread_count > 0"
          class="h-5 min-w-5 rounded-full bg-primary text-primary-foreground text-[10px] flex items-center justify-center px-1"
        >
          {{ conv.unread_count }}
        </div>
      </RouterLink>
    </div>
  </div>
</template>

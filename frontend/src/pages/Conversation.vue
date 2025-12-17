<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useMessageStore } from '@/stores/messages'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import UserAvatar from '@/components/user/UserAvatar.vue'
import { Send, ArrowLeft } from 'lucide-vue-next'

interface Message {
  id: string
  content: string
  sender_id: string
  created_at: string
}

interface ConvMeta {
  id: string
  other_user: { id: string; username: string; avatar: string | null }
}

const route = useRoute()
const auth = useAuthStore()
const msgStore = useMessageStore()
const { toast } = useToast()

const meta = ref<ConvMeta | null>(null)
const messages = ref<Message[]>([])
const loading = ref(true)
const sending = ref(false)
const newMsg = ref('')
const scrollContainer = ref<HTMLElement | null>(null)

const convId = route.params.id as string

onMounted(async () => {
  try {
    const [convData, msgData] = await Promise.all([
      api.get<ConvMeta>(`/messages/conversations/${convId}`),
      api.get<Message[]>(`/messages/conversations/${convId}/messages`),
    ])
    meta.value = convData
    messages.value = msgData
    msgStore.clearUnread()
  } catch {
    toast({ variant: 'destructive', title: 'Failed to load conversation' })
  } finally {
    loading.value = false
    await nextTick()
    scrollToBottom()
  }
})

async function sendMessage() {
  if (!newMsg.value.trim() || sending.value) return
  const content = newMsg.value.trim()
  newMsg.value = ''
  sending.value = true
  try {
    const msg = await api.post<Message>(`/messages/conversations/${convId}/messages`, { content })
    messages.value.push(msg)
    await nextTick()
    scrollToBottom()
  } catch {
    toast({ variant: 'destructive', title: 'Failed to send message' })
    newMsg.value = content
  } finally {
    sending.value = false
  }
}

function scrollToBottom() {
  if (scrollContainer.value) {
    scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
  }
}

watch(messages, async () => {
  await nextTick()
  scrollToBottom()
})

function formatTime(d: string) {
  return new Intl.DateTimeFormat('en', { hour: 'numeric', minute: '2-digit' }).format(new Date(d))
}
</script>

<template>
  <div class="flex flex-col h-[calc(100vh-10rem)]">
    <!-- Header -->
    <div class="flex items-center gap-3 pb-3 border-b mb-3">
      <Button variant="ghost" size="icon" as-child class="-ml-1">
        <RouterLink to="/messages"><ArrowLeft class="h-4 w-4" /></RouterLink>
      </Button>
      <template v-if="meta">
        <UserAvatar :src="meta.other_user.avatar" :username="meta.other_user.username" size="md" />
        <RouterLink :to="`/users/${meta.other_user.username}`" class="font-semibold hover:underline">
          {{ meta.other_user.username }}
        </RouterLink>
      </template>
      <template v-else>
        <Skeleton class="h-8 w-8 rounded-full" />
        <Skeleton class="h-4 w-24" />
      </template>
    </div>

    <!-- Messages -->
    <div
      ref="scrollContainer"
      class="flex-1 overflow-y-auto space-y-3 pr-2"
    >
      <template v-if="loading">
        <div v-for="i in 5" :key="i" class="flex gap-2" :class="i % 2 === 0 ? 'flex-row-reverse' : ''">
          <Skeleton class="h-7 w-7 rounded-full shrink-0" />
          <Skeleton class="h-10 w-48 rounded-2xl" />
        </div>
      </template>

      <div
        v-for="msg in messages"
        :key="msg.id"
        class="flex gap-2 items-end"
        :class="msg.sender_id === auth.user?.id ? 'flex-row-reverse' : ''"
      >
        <div
          class="max-w-xs lg:max-w-md px-3 py-2 rounded-2xl text-sm"
          :class="msg.sender_id === auth.user?.id
            ? 'bg-primary text-primary-foreground rounded-br-none'
            : 'bg-muted rounded-bl-none'"
        >
          <p class="break-words">{{ msg.content }}</p>
          <p class="text-[10px] mt-0.5 opacity-60 text-right">{{ formatTime(msg.created_at) }}</p>
        </div>
      </div>
    </div>

    <!-- Input -->
    <div class="flex gap-2 pt-3 border-t mt-3">
      <Input
        v-model="newMsg"
        placeholder="Message…"
        @keydown.enter.prevent="sendMessage"
        :disabled="sending"
        class="flex-1"
      />
      <Button size="icon" :disabled="sending || !newMsg.trim()" @click="sendMessage">
        <Send class="h-4 w-4" />
      </Button>
    </div>
  </div>
</template>

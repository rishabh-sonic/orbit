<script setup lang="ts">
import { onMounted } from 'vue'
import { useNotificationStore } from '@/stores/notifications'
import { Button } from '@/components/ui/button'
import type { Notification } from '@/stores/notifications'
import {
  MessageSquare, UserPlus, Bell, Bookmark, Activity,
  Trash2, AtSign, CheckCheck,
} from 'lucide-vue-next'

const notifStore = useNotificationStore()

const typeIcon: Record<string, unknown> = {
  COMMENT_REPLY: MessageSquare,
  USER_FOLLOWED: UserPlus,
  POST_UPDATED: Bookmark,
  FOLLOWED_POST: Bookmark,
  USER_ACTIVITY: Activity,
  POST_DELETED: Trash2,
  MENTION: AtSign,
}

const typeLabel: Record<string, string> = {
  COMMENT_REPLY: 'replied to your comment',
  USER_FOLLOWED: 'followed you',
  POST_UPDATED: 'updated a post you follow',
  FOLLOWED_POST: 'commented on a post you follow',
  USER_ACTIVITY: 'new activity',
  POST_DELETED: 'deleted a post',
  MENTION: 'mentioned you',
}

onMounted(async () => {
  await notifStore.fetchNotifications()
})

async function markAllRead() {
  await notifStore.markAllRead()
}

function formatDate(d: string) {
  return new Intl.DateTimeFormat('en', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' }).format(new Date(d))
}

function linkFor(n: Notification): string {
  if (n.post_id) return `/posts/${n.post_id}`
  if (n.from_user_id) return `/users/${n.from_user_id}`
  return '#'
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Notifications</h1>
      <Button
        v-if="notifStore.unreadCount > 0"
        variant="outline"
        size="sm"
        @click="markAllRead"
      >
        <CheckCheck class="h-4 w-4 mr-1" /> Mark all read
      </Button>
    </div>

    <template v-if="notifStore.notifications.length === 0">
      <div class="flex flex-col items-center py-16 gap-3 text-muted-foreground">
        <Bell class="h-10 w-10 opacity-30" />
        <p>No notifications yet.</p>
      </div>
    </template>

    <div v-else class="space-y-1">
      <RouterLink
        v-for="n in notifStore.notifications"
        :key="n.id"
        :to="linkFor(n)"
        class="flex items-start gap-3 p-3 rounded-lg hover:bg-accent transition-colors"
        :class="!n.read ? 'bg-accent/50' : ''"
      >
        <div class="h-8 w-8 rounded-full bg-muted flex items-center justify-center shrink-0">
          <component :is="typeIcon[n.type] ?? Bell" class="h-4 w-4 text-muted-foreground" />
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm">
            <span v-if="n.content">{{ n.content }}</span>
            <span v-else>{{ typeLabel[n.type] ?? n.type }}</span>
          </p>
          <p class="text-xs text-muted-foreground mt-0.5">{{ formatDate(n.created_at) }}</p>
        </div>
        <div v-if="!n.read" class="h-2 w-2 rounded-full bg-primary mt-2 shrink-0" />
      </RouterLink>
    </div>
  </div>
</template>

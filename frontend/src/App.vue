<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'
import { useMessageStore } from '@/stores/messages'
import { useWebSocket } from '@/lib/ws'
import AppLayout from '@/components/layout/AppLayout.vue'
import { Toaster } from '@/components/ui/toast'

const auth = useAuthStore()
const notifStore = useNotificationStore()
const msgStore = useMessageStore()
const { connect } = useWebSocket()

onMounted(async () => {
  if (auth.isLoggedIn) {
    await auth.fetchMe()
    await Promise.all([notifStore.fetchUnreadCount(), msgStore.fetchUnreadCount()])
    connect()
  }
})
</script>

<template>
  <AppLayout />
  <Toaster />
</template>

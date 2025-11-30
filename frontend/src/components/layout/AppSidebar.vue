<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useMessageStore } from '@/stores/messages'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Home, Search, MessageSquare, Bell, Settings, Shield, Users, BarChart2 } from 'lucide-vue-next'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const auth = useAuthStore()
const msgStore = useMessageStore()

const navItems = [
  { to: '/', icon: Home, label: 'Feed' },
  { to: '/search', icon: Search, label: 'Search' },
]

const authItems = [
  { to: '/messages', icon: MessageSquare, label: 'Messages', badge: () => msgStore.unreadCount },
  { to: '/notifications', icon: Bell, label: 'Notifications' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

const adminItems = [
  { to: '/admin', icon: BarChart2, label: 'Stats' },
  { to: '/admin/users', icon: Users, label: 'Users' },
]
</script>

<template>
  <aside
    :class="[
      'fixed top-14 left-0 z-20 h-[calc(100vh-3.5rem)] w-56 bg-background border-r flex flex-col transition-transform duration-200',
      'lg:translate-x-0',
      open ? 'translate-x-0' : '-translate-x-full'
    ]"
  >
    <nav class="flex-1 overflow-y-auto p-3 space-y-1">
      <template v-for="item in navItems" :key="item.to">
        <RouterLink
          :to="item.to"
          @click="emit('close')"
          class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium hover:bg-accent hover:text-accent-foreground transition-colors"
          active-class="bg-accent text-accent-foreground"
        >
          <component :is="item.icon" class="h-4 w-4 shrink-0" />
          {{ item.label }}
        </RouterLink>
      </template>

      <template v-if="auth.isLoggedIn">
        <Separator class="my-2" />
        <template v-for="item in authItems" :key="item.to">
          <RouterLink
            :to="item.to"
            @click="emit('close')"
            class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium hover:bg-accent hover:text-accent-foreground transition-colors"
            active-class="bg-accent text-accent-foreground"
          >
            <component :is="item.icon" class="h-4 w-4 shrink-0" />
            {{ item.label }}
            <Badge
              v-if="item.badge && item.badge() > 0"
              class="ml-auto h-5 min-w-5 px-1 text-[10px]"
            >
              {{ item.badge() }}
            </Badge>
          </RouterLink>
        </template>
      </template>

      <template v-if="auth.isAdmin">
        <Separator class="my-2" />
        <p class="px-3 py-1 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          Admin
        </p>
        <template v-for="item in adminItems" :key="item.to">
          <RouterLink
            :to="item.to"
            @click="emit('close')"
            class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium hover:bg-accent hover:text-accent-foreground transition-colors"
            active-class="bg-accent text-accent-foreground"
          >
            <component :is="item.icon" class="h-4 w-4 shrink-0" />
            {{ item.label }}
          </RouterLink>
        </template>
      </template>
    </nav>

    <div class="p-3 border-t text-xs text-muted-foreground">
      <RouterLink to="/admin" class="flex items-center gap-1 text-muted-foreground hover:text-foreground" v-if="auth.isAdmin">
        <Shield class="h-3 w-3" /> Admin Panel
      </RouterLink>
      <span v-else>Orbit &copy; {{ new Date().getFullYear() }}</span>
    </div>
  </aside>
</template>

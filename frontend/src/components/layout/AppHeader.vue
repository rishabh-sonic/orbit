<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'
import { useMessageStore } from '@/stores/messages'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Menu, Search, Bell, MessageSquare, PenSquare, LogOut, User, Settings, Shield } from 'lucide-vue-next'

const emit = defineEmits<{ 'toggle-sidebar': [] }>()

const router = useRouter()
const auth = useAuthStore()
const notifStore = useNotificationStore()
const msgStore = useMessageStore()

const searchQuery = ref('')

function onSearch() {
  if (searchQuery.value.trim()) {
    router.push({ path: '/search', query: { q: searchQuery.value.trim() } })
  }
}

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <header class="fixed top-0 left-0 right-0 z-30 h-14 border-b bg-background/95 backdrop-blur">
    <div class="flex items-center gap-2 h-full px-4">
      <!-- Hamburger -->
      <Button variant="ghost" size="icon" @click="emit('toggle-sidebar')" class="lg:hidden">
        <Menu class="h-5 w-5" />
      </Button>

      <!-- Logo -->
      <RouterLink to="/" class="font-bold text-lg tracking-tight mr-2 hidden sm:block">
        Orbit
      </RouterLink>

      <!-- Search -->
      <form @submit.prevent="onSearch" class="flex-1 max-w-sm">
        <div class="relative">
          <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Search..."
            class="pl-8 h-9"
          />
        </div>
      </form>

      <div class="flex items-center gap-1 ml-auto">
        <!-- New Post -->
        <Button v-if="auth.isLoggedIn" variant="default" size="sm" as-child>
          <RouterLink to="/posts/new">
            <PenSquare class="h-4 w-4 mr-1" />
            <span class="hidden sm:inline">New Post</span>
          </RouterLink>
        </Button>

        <!-- Notifications -->
        <Button v-if="auth.isLoggedIn" variant="ghost" size="icon" as-child class="relative">
          <RouterLink to="/notifications">
            <Bell class="h-5 w-5" />
            <Badge
              v-if="notifStore.unreadCount > 0"
              class="absolute -top-1 -right-1 h-4 min-w-4 px-1 text-[10px] flex items-center justify-center"
            >
              {{ notifStore.unreadCount > 99 ? '99+' : notifStore.unreadCount }}
            </Badge>
          </RouterLink>
        </Button>

        <!-- Messages -->
        <Button v-if="auth.isLoggedIn" variant="ghost" size="icon" as-child class="relative">
          <RouterLink to="/messages">
            <MessageSquare class="h-5 w-5" />
            <Badge
              v-if="msgStore.unreadCount > 0"
              class="absolute -top-1 -right-1 h-4 min-w-4 px-1 text-[10px] flex items-center justify-center"
            >
              {{ msgStore.unreadCount > 99 ? '99+' : msgStore.unreadCount }}
            </Badge>
          </RouterLink>
        </Button>

        <!-- User menu -->
        <DropdownMenu v-if="auth.isLoggedIn">
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" size="icon" class="rounded-full">
              <Avatar class="h-8 w-8">
                <AvatarImage v-if="auth.user?.avatar" :src="auth.user.avatar" />
                <AvatarFallback>{{ auth.user?.username?.[0]?.toUpperCase() ?? 'U' }}</AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="w-48">
            <DropdownMenuItem as-child>
              <RouterLink :to="`/users/${auth.user?.username}`" class="flex items-center gap-2">
                <User class="h-4 w-4" /> Profile
              </RouterLink>
            </DropdownMenuItem>
            <DropdownMenuItem as-child>
              <RouterLink to="/settings" class="flex items-center gap-2">
                <Settings class="h-4 w-4" /> Settings
              </RouterLink>
            </DropdownMenuItem>
            <template v-if="auth.isAdmin">
              <DropdownMenuSeparator />
              <DropdownMenuItem as-child>
                <RouterLink to="/admin" class="flex items-center gap-2">
                  <Shield class="h-4 w-4" /> Admin
                </RouterLink>
              </DropdownMenuItem>
            </template>
            <DropdownMenuSeparator />
            <DropdownMenuItem class="text-destructive cursor-pointer" @click="logout">
              <LogOut class="h-4 w-4 mr-2" /> Log out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <!-- Guest buttons -->
        <template v-else>
          <Button variant="ghost" size="sm" as-child>
            <RouterLink to="/login">Sign in</RouterLink>
          </Button>
          <Button size="sm" as-child>
            <RouterLink to="/register">Sign up</RouterLink>
          </Button>
        </template>
      </div>
    </div>
  </header>
</template>

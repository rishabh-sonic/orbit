<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import UserAvatar from '@/components/user/UserAvatar.vue'
import { Search, Ban, CheckCircle } from 'lucide-vue-next'

interface AdminUser {
  id: string
  username: string
  email: string
  avatar: string | null
  role: 'USER' | 'ADMIN'
  banned: boolean
  created_at: string
}

const { toast } = useToast()
const users = ref<AdminUser[]>([])
const loading = ref(true)
const searchQ = ref('')
const actionLoading = ref<string | null>(null)

onMounted(load)

async function load() {
  loading.value = true
  try {
    const data = await api.get<AdminUser[]>('/admin/users')
    users.value = data
  } catch {
    toast({ variant: 'destructive', title: 'Failed to load users' })
  } finally {
    loading.value = false
  }
}

const filtered = () =>
  users.value.filter(u =>
    u.username.toLowerCase().includes(searchQ.value.toLowerCase()) ||
    u.email.toLowerCase().includes(searchQ.value.toLowerCase()),
  )

async function toggleBan(user: AdminUser) {
  actionLoading.value = user.id
  try {
    if (user.banned) {
      await api.post(`/admin/users/${user.id}/unban`)
      user.banned = false
    } else {
      await api.post(`/admin/users/${user.id}/ban`)
      user.banned = true
    }
    toast({ title: user.banned ? 'User banned' : 'User unbanned' })
  } catch {
    toast({ variant: 'destructive', title: 'Action failed' })
  } finally {
    actionLoading.value = null
  }
}

function formatDate(d: string) {
  return new Intl.DateTimeFormat('en', { dateStyle: 'medium' }).format(new Date(d))
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Users</h1>
      <RouterLink to="/admin" class="text-sm text-muted-foreground hover:underline">← Stats</RouterLink>
    </div>

    <div class="relative mb-4">
      <Search class="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
      <Input v-model="searchQ" placeholder="Search by username or email…" class="pl-9" />
    </div>

    <template v-if="loading">
      <div v-for="i in 5" :key="i" class="flex items-center gap-3 p-3 border-b">
        <Skeleton class="h-8 w-8 rounded-full" />
        <div class="flex-1 space-y-1.5">
          <Skeleton class="h-4 w-28" />
          <Skeleton class="h-3 w-40" />
        </div>
        <Skeleton class="h-8 w-16 rounded" />
      </div>
    </template>

    <div v-else class="border rounded-lg divide-y overflow-hidden">
      <div
        v-for="user in filtered()"
        :key="user.id"
        class="flex items-center gap-3 p-3 hover:bg-muted/50"
      >
        <RouterLink :to="`/users/${user.username}`">
          <UserAvatar :src="user.avatar" :username="user.username" size="sm" />
        </RouterLink>
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 flex-wrap">
            <RouterLink :to="`/users/${user.username}`" class="text-sm font-medium hover:underline">
              {{ user.username }}
            </RouterLink>
            <Badge v-if="user.role === 'ADMIN'" variant="default" class="h-4 px-1.5 text-[10px]">
              Admin
            </Badge>
            <Badge v-if="user.banned" variant="destructive" class="h-4 px-1.5 text-[10px]">
              Banned
            </Badge>
          </div>
          <p class="text-xs text-muted-foreground">{{ user.email }} · joined {{ formatDate(user.created_at) }}</p>
        </div>
        <Button
          v-if="user.role !== 'ADMIN'"
          :variant="user.banned ? 'outline' : 'destructive'"
          size="sm"
          :disabled="actionLoading === user.id"
          @click="toggleBan(user)"
        >
          <CheckCircle v-if="user.banned" class="h-3.5 w-3.5 mr-1" />
          <Ban v-else class="h-3.5 w-3.5 mr-1" />
          {{ user.banned ? 'Unban' : 'Ban' }}
        </Button>
      </div>

      <div v-if="filtered().length === 0" class="p-8 text-center text-muted-foreground text-sm">
        No users found.
      </div>
    </div>
  </div>
</template>

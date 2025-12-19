<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Users, FileText, Eye } from 'lucide-vue-next'

interface DauEntry { date: string; count: number }
interface Stats {
  dau: DauEntry[]
  new_users: DauEntry[]
  posts: DauEntry[]
}

const { toast } = useToast()
const stats = ref<Stats | null>(null)
const loading = ref(true)

// Simple 30-day range defaults
const now = new Date()
const end = now.toISOString().split('T')[0]
const start = new Date(now.setDate(now.getDate() - 30)).toISOString().split('T')[0]

onMounted(async () => {
  try {
    const [dau, newUsers, posts] = await Promise.all([
      api.get<DauEntry[]>(`/admin/stats/dau-range?start=${start}&end=${end}`),
      api.get<DauEntry[]>(`/admin/stats/new-users-range?start=${start}&end=${end}`),
      api.get<DauEntry[]>(`/admin/stats/posts-range?start=${start}&end=${end}`),
    ])
    stats.value = { dau, new_users: newUsers, posts }
  } catch {
    toast({ variant: 'destructive', title: 'Failed to load stats' })
  } finally {
    loading.value = false
  }
})

function total(data: DauEntry[]) {
  return data.reduce((sum, d) => sum + d.count, 0)
}

function sparkBars(data: DauEntry[]) {
  const max = Math.max(...data.map(d => d.count), 1)
  return data.map(d => Math.round((d.count / max) * 100))
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Admin — Stats</h1>
      <RouterLink to="/admin/users" class="text-sm text-muted-foreground hover:underline">Users →</RouterLink>
    </div>

    <template v-if="loading">
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <Skeleton v-for="i in 3" :key="i" class="h-28 rounded-lg" />
      </div>
    </template>

    <template v-else-if="stats">
      <!-- Summary cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium text-muted-foreground">Daily Active Users</CardTitle>
            <Eye class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p class="text-2xl font-bold">{{ total(stats.dau).toLocaleString() }}</p>
            <p class="text-xs text-muted-foreground">last 30 days</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium text-muted-foreground">New Users</CardTitle>
            <Users class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p class="text-2xl font-bold">{{ total(stats.new_users).toLocaleString() }}</p>
            <p class="text-xs text-muted-foreground">last 30 days</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium text-muted-foreground">Posts Created</CardTitle>
            <FileText class="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p class="text-2xl font-bold">{{ total(stats.posts).toLocaleString() }}</p>
            <p class="text-xs text-muted-foreground">last 30 days</p>
          </CardContent>
        </Card>
      </div>

      <!-- Sparkline charts (CSS bars) -->
      <div class="space-y-6">
        <div v-for="(item, key) in [
          { label: 'Daily Active Users', icon: Eye, data: stats.dau },
          { label: 'New Registrations', icon: Users, data: stats.new_users },
          { label: 'Posts Published', icon: FileText, data: stats.posts },
        ]" :key="key">
          <div class="flex items-center gap-2 mb-2">
            <component :is="item.icon" class="h-4 w-4 text-muted-foreground" />
            <h2 class="text-sm font-semibold">{{ item.label }}</h2>
          </div>
          <div class="border rounded-lg p-4">
            <div class="flex items-end gap-0.5 h-24">
              <div
                v-for="(pct, i) in sparkBars(item.data)"
                :key="i"
                class="flex-1 bg-primary/80 rounded-t"
                :style="{ height: `${Math.max(pct, 2)}%` }"
                :title="`${item.data[i]?.date}: ${item.data[i]?.count}`"
              />
            </div>
            <div class="flex justify-between text-[10px] text-muted-foreground mt-1">
              <span>{{ item.data[0]?.date }}</span>
              <span>{{ item.data[item.data.length - 1]?.date }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

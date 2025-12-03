<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/lib/api'

const props = defineProps<{ provider: string }>()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const error = ref<string | null>(null)

onMounted(async () => {
  const code = route.query.code as string | undefined
  const state = route.query.state as string | undefined

  if (!code) {
    error.value = 'No authorization code received.'
    return
  }

  try {
    const data = await api.post<{ token: string }>(`/auth/${props.provider}`, { code, state })
    auth.setToken(data.token)
    await auth.fetchMe()
    router.replace('/')
  } catch {
    error.value = 'OAuth sign-in failed. Please try again.'
  }
})
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-[calc(100vh-8rem)] gap-4">
    <template v-if="!error">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      <p class="text-muted-foreground">Completing sign-in…</p>
    </template>
    <template v-else>
      <p class="text-destructive">{{ error }}</p>
      <RouterLink to="/login" class="underline text-sm">Back to sign in</RouterLink>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api, ApiError } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useToast } from '@/components/ui/toast'
import { Github, Chrome } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const { toast } = useToast()

const identifier = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    const data = await api.post<{ token: string }>('/auth/login', {
      identifier: identifier.value,
      password: password.value,
    })
    auth.setToken(data.token)
    await auth.fetchMe()
    const redirect = route.query.redirect as string | undefined
    router.push(redirect ?? '/')
  } catch (e) {
    toast({
      variant: 'destructive',
      title: 'Login failed',
      description: e instanceof ApiError ? e.message : 'Please try again.',
    })
  } finally {
    loading.value = false
  }
}

function oauthRedirect(provider: string) {
  const base = import.meta.env.VITE_API_URL ?? ''
  window.location.href = `${base}/oauth/${provider}`
}
</script>

<template>
  <div class="flex items-center justify-center min-h-[calc(100vh-8rem)]">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">Sign in</CardTitle>
        <CardDescription>Enter your email or username</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <form @submit.prevent="submit" class="space-y-4">
          <div class="space-y-1">
            <Label>Email or username</Label>
            <Input v-model="identifier" autocomplete="username" required />
          </div>
          <div class="space-y-1">
            <div class="flex justify-between items-center">
              <Label>Password</Label>
              <RouterLink to="/forgot-password" class="text-xs text-muted-foreground hover:underline">
                Forgot password?
              </RouterLink>
            </div>
            <Input v-model="password" type="password" autocomplete="current-password" required />
          </div>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Signing in…' : 'Sign in' }}
          </Button>
        </form>

        <div class="relative">
          <div class="absolute inset-0 flex items-center"><span class="w-full border-t" /></div>
          <div class="relative flex justify-center text-xs uppercase">
            <span class="bg-card px-2 text-muted-foreground">or continue with</span>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-2">
          <Button variant="outline" @click="oauthRedirect('github')" type="button">
            <Github class="h-4 w-4 mr-2" /> GitHub
          </Button>
          <Button variant="outline" @click="oauthRedirect('google')" type="button">
            <Chrome class="h-4 w-4 mr-2" /> Google
          </Button>
        </div>

        <p class="text-center text-sm text-muted-foreground">
          No account?
          <RouterLink to="/register" class="underline underline-offset-2">Sign up</RouterLink>
        </p>
      </CardContent>
    </Card>
  </div>
</template>

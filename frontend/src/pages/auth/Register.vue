<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useToast } from '@/components/ui/toast'

const router = useRouter()
const auth = useAuthStore()
const { toast } = useToast()

const username = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    const data = await api.post<{ token: string }>('/auth/register', {
      username: username.value,
      email: email.value,
      password: password.value,
    })
    auth.setToken(data.token)
    await auth.fetchMe()
    router.push('/')
  } catch (e) {
    toast({
      variant: 'destructive',
      title: 'Registration failed',
      description: e instanceof ApiError ? e.message : 'Please try again.',
    })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-[calc(100vh-8rem)]">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">Create account</CardTitle>
        <CardDescription>Pick a username and you're in</CardDescription>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="submit" class="space-y-4">
          <div class="space-y-1">
            <Label>Username</Label>
            <Input v-model="username" autocomplete="username" required />
          </div>
          <div class="space-y-1">
            <Label>Email</Label>
            <Input v-model="email" type="email" autocomplete="email" required />
          </div>
          <div class="space-y-1">
            <Label>Password</Label>
            <Input v-model="password" type="password" autocomplete="new-password" required minlength="8" />
          </div>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Creating account…' : 'Create account' }}
          </Button>
        </form>
        <p class="text-center text-sm text-muted-foreground mt-4">
          Already have an account?
          <RouterLink to="/login" class="underline underline-offset-2">Sign in</RouterLink>
        </p>
      </CardContent>
    </Card>
  </div>
</template>

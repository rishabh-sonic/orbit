<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useToast } from '@/components/ui/toast'

const router = useRouter()
const { toast } = useToast()

type Step = 'send' | 'verify' | 'reset'
const step = ref<Step>('send')
const email = ref('')
const code = ref('')
const resetToken = ref('')
const newPassword = ref('')
const loading = ref(false)

async function sendCode() {
  loading.value = true
  try {
    await api.post('/auth/forgot/send', { email: email.value })
    step.value = 'verify'
  } catch (e) {
    toast({ variant: 'destructive', title: 'Error', description: e instanceof ApiError ? e.message : 'Try again.' })
  } finally {
    loading.value = false
  }
}

async function verifyCode() {
  loading.value = true
  try {
    const data = await api.post<{ reset_token: string }>('/auth/forgot/verify', {
      email: email.value,
      code: code.value,
    })
    resetToken.value = data.reset_token
    step.value = 'reset'
  } catch (e) {
    toast({ variant: 'destructive', title: 'Invalid code', description: e instanceof ApiError ? e.message : 'Try again.' })
  } finally {
    loading.value = false
  }
}

async function resetPassword() {
  loading.value = true
  try {
    await api.post('/auth/forgot/reset', {
      reset_token: resetToken.value,
      new_password: newPassword.value,
    })
    toast({ title: 'Password updated', description: 'You can now sign in.' })
    router.push('/login')
  } catch (e) {
    toast({ variant: 'destructive', title: 'Error', description: e instanceof ApiError ? e.message : 'Try again.' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-[calc(100vh-8rem)]">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <CardTitle>Reset password</CardTitle>
        <CardDescription>
          <span v-if="step === 'send'">Enter your email to receive a reset code</span>
          <span v-else-if="step === 'verify'">Enter the code sent to {{ email }}</span>
          <span v-else>Choose a new password</span>
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form v-if="step === 'send'" @submit.prevent="sendCode" class="space-y-4">
          <div class="space-y-1">
            <Label>Email</Label>
            <Input v-model="email" type="email" required />
          </div>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Sending…' : 'Send code' }}
          </Button>
        </form>

        <form v-else-if="step === 'verify'" @submit.prevent="verifyCode" class="space-y-4">
          <div class="space-y-1">
            <Label>Reset code</Label>
            <Input v-model="code" maxlength="6" placeholder="123456" required />
          </div>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Verifying…' : 'Verify code' }}
          </Button>
        </form>

        <form v-else @submit.prevent="resetPassword" class="space-y-4">
          <div class="space-y-1">
            <Label>New password</Label>
            <Input v-model="newPassword" type="password" minlength="8" required />
          </div>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? 'Saving…' : 'Set new password' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

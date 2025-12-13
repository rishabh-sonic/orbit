<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api, ApiError } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import UserAvatar from '@/components/user/UserAvatar.vue'

const auth = useAuthStore()
const { toast } = useToast()

const username = ref('')
const introduction = ref('')
const avatarUrl = ref<string | null>(null)
const avatarFile = ref<File | null>(null)
const profileLoading = ref(false)
const avatarLoading = ref(false)

const emailNotif = ref(true)
const pushNotif = ref(false)
const notifLoading = ref(false)

onMounted(() => {
  if (auth.user) {
    username.value = auth.user.username
    introduction.value = auth.user.introduction ?? ''
    avatarUrl.value = auth.user.avatar
  }
})

function onAvatarChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  avatarFile.value = file
  avatarUrl.value = URL.createObjectURL(file)
}

async function saveProfile() {
  profileLoading.value = true
  try {
    if (avatarFile.value) {
      avatarLoading.value = true
      const fd = new FormData()
      fd.append('file', avatarFile.value)
      const uploaded = await api.upload<{ url: string }>('/upload', fd)
      await api.post('/users/me/avatar', { avatar_url: uploaded.url })
      avatarFile.value = null
      avatarLoading.value = false
    }

    const updated = await api.put<typeof auth.user>('/users/me', {
      username: username.value,
      introduction: introduction.value,
    })
    if (updated) auth.setUser(updated)
    await auth.fetchMe()
    toast({ title: 'Profile updated' })
  } catch (e) {
    toast({
      variant: 'destructive',
      title: 'Update failed',
      description: e instanceof ApiError ? e.message : 'Please try again.',
    })
  } finally {
    profileLoading.value = false
    avatarLoading.value = false
  }
}

async function saveNotifications() {
  notifLoading.value = true
  try {
    await api.put('/users/me/notification-preferences', {
      email_notifications: emailNotif.value,
      push_notifications: pushNotif.value,
    })
    toast({ title: 'Notification preferences saved' })
  } catch {
    toast({ variant: 'destructive', title: 'Failed to save preferences' })
  } finally {
    notifLoading.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold">Settings</h1>

    <!-- Profile -->
    <Card>
      <CardHeader>
        <CardTitle>Profile</CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="flex items-center gap-4">
          <UserAvatar :src="avatarUrl" :username="username" size="lg" class="h-16 w-16 text-xl" />
          <div>
            <Label class="cursor-pointer inline-flex">
              <span class="text-sm text-primary underline underline-offset-2">Change avatar</span>
              <input type="file" accept="image/*" class="sr-only" @change="onAvatarChange" />
            </Label>
            <p class="text-xs text-muted-foreground mt-1">JPG, PNG, WebP — max 5 MB</p>
          </div>
        </div>

        <div class="space-y-1">
          <Label>Username</Label>
          <Input v-model="username" required />
        </div>
        <div class="space-y-1">
          <Label>Bio</Label>
          <Textarea v-model="introduction" rows="3" placeholder="Tell us about yourself…" maxlength="500" />
        </div>
        <Button :disabled="profileLoading || avatarLoading" @click="saveProfile">
          {{ profileLoading ? 'Saving…' : 'Save profile' }}
        </Button>
      </CardContent>
    </Card>

    <!-- Notification preferences -->
    <Card>
      <CardHeader>
        <CardTitle>Notifications</CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium">Email notifications</p>
            <p class="text-xs text-muted-foreground">Receive emails for replies and mentions</p>
          </div>
          <input type="checkbox" v-model="emailNotif" class="h-4 w-4 cursor-pointer" />
        </div>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium">Push notifications</p>
            <p class="text-xs text-muted-foreground">Browser push for real-time alerts</p>
          </div>
          <input type="checkbox" v-model="pushNotif" class="h-4 w-4 cursor-pointer" />
        </div>
        <Button :disabled="notifLoading" @click="saveNotifications">
          {{ notifLoading ? 'Saving…' : 'Save preferences' }}
        </Button>
      </CardContent>
    </Card>
  </div>
</template>

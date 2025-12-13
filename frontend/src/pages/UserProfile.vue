<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import UserAvatar from '@/components/user/UserAvatar.vue'
import PostFeed from '@/components/post/PostFeed.vue'

interface UserInfo {
  id: string
  username: string
  avatar: string | null
  introduction: string | null
  follower_count: number
  following_count: number
  post_count: number
  is_following?: boolean
  created_at: string
}

interface FollowUser {
  id: string
  username: string
  avatar: string | null
}

const route = useRoute()
const auth = useAuthStore()
const { toast } = useToast()

const user = ref<UserInfo | null>(null)
const followers = ref<FollowUser[]>([])
const following = ref<FollowUser[]>([])
const loading = ref(true)
const followLoading = ref(false)

const isOwnProfile = computed(() => auth.user?.username === route.params.id)

async function loadProfile(id: string | string[] | undefined) {
  if (!id) return
  loading.value = true
  user.value = null
  followers.value = []
  following.value = []
  try {
    const [userData, followersData, followingData] = await Promise.all([
      api.get<UserInfo>(`/users/${id}`),
      api.get<FollowUser[]>(`/users/${id}/followers`),
      api.get<FollowUser[]>(`/users/${id}/following`),
    ])
    user.value = userData
    followers.value = followersData
    following.value = followingData
  } catch {
    toast({ variant: 'destructive', title: 'Failed to load profile' })
  } finally {
    loading.value = false
  }
}

onMounted(() => loadProfile(route.params.id))

// Re-load when navigating between different user profiles
watch(() => route.params.id, (id) => { if (id) loadProfile(id) })

async function toggleFollow() {
  if (!user.value || isOwnProfile.value || !auth.isLoggedIn) return
  followLoading.value = true
  try {
    if (user.value.is_following) {
      await api.delete(`/subscriptions/users/${user.value.username}`)
      user.value.is_following = false
      user.value.follower_count--
    } else {
      await api.post(`/subscriptions/users/${user.value.username}`)
      user.value.is_following = true
      user.value.follower_count++
    }
  } catch {
    toast({ variant: 'destructive', title: 'Failed to update follow' })
  } finally {
    followLoading.value = false
  }
}

function formatDate(d: string) {
  return new Intl.DateTimeFormat('en', { year: 'numeric', month: 'long' }).format(new Date(d))
}
</script>

<template>
  <div>
    <template v-if="loading">
      <div class="flex gap-4 mb-6">
        <Skeleton class="h-16 w-16 rounded-full" />
        <div class="flex-1 space-y-2">
          <Skeleton class="h-5 w-32" />
          <Skeleton class="h-4 w-48" />
          <Skeleton class="h-3 w-24" />
        </div>
      </div>
    </template>

    <template v-else-if="user">
      <!-- Profile header -->
      <div class="flex items-start gap-4 mb-6">
        <UserAvatar :src="user.avatar" :username="user.username" size="lg" class="h-16 w-16 text-xl" />
        <div class="flex-1">
          <div class="flex items-center gap-3 flex-wrap">
            <h1 class="text-xl font-bold">{{ user.username }}</h1>
            <Button
              v-if="auth.isLoggedIn && !isOwnProfile"
              :variant="user.is_following ? 'outline' : 'default'"
              size="sm"
              :disabled="followLoading"
              @click="toggleFollow"
            >
              {{ user.is_following ? 'Unfollow' : 'Follow' }}
            </Button>
            <Button v-if="isOwnProfile" variant="outline" size="sm" as-child>
              <RouterLink to="/settings">Edit profile</RouterLink>
            </Button>
          </div>
          <p v-if="user.introduction" class="text-sm text-muted-foreground mt-1">
            {{ user.introduction }}
          </p>
          <div class="flex gap-4 text-sm mt-2">
            <span><strong>{{ user.post_count }}</strong> posts</span>
            <span><strong>{{ user.follower_count }}</strong> followers</span>
            <span><strong>{{ user.following_count }}</strong> following</span>
          </div>
          <p class="text-xs text-muted-foreground mt-1">Joined {{ formatDate(user.created_at) }}</p>
        </div>
      </div>

      <!-- Tabs -->
      <Tabs default-value="posts">
        <TabsList class="mb-4">
          <TabsTrigger value="posts">Posts</TabsTrigger>
          <TabsTrigger value="followers">Followers</TabsTrigger>
          <TabsTrigger value="following">Following</TabsTrigger>
        </TabsList>

        <TabsContent value="posts">
          <PostFeed :user-id="user.id" />
        </TabsContent>

        <TabsContent value="followers">
          <div v-if="followers.length === 0" class="text-muted-foreground text-sm py-8 text-center">
            No followers yet.
          </div>
          <div class="space-y-2">
            <RouterLink
              v-for="f in followers"
              :key="f.id"
              :to="`/users/${f.username}`"
              class="flex items-center gap-2 p-2 rounded-md hover:bg-accent"
            >
              <UserAvatar :src="f.avatar" :username="f.username" size="sm" />
              <span class="text-sm font-medium">{{ f.username }}</span>
            </RouterLink>
          </div>
        </TabsContent>

        <TabsContent value="following">
          <div v-if="following.length === 0" class="text-muted-foreground text-sm py-8 text-center">
            Not following anyone yet.
          </div>
          <div class="space-y-2">
            <RouterLink
              v-for="f in following"
              :key="f.id"
              :to="`/users/${f.username}`"
              class="flex items-center gap-2 p-2 rounded-md hover:bg-accent"
            >
              <UserAvatar :src="f.avatar" :username="f.username" size="sm" />
              <span class="text-sm font-medium">{{ f.username }}</span>
            </RouterLink>
          </div>
        </TabsContent>
      </Tabs>
    </template>

    <template v-else>
      <p class="text-muted-foreground text-center py-12">User not found.</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import PostFeed from '@/components/post/PostFeed.vue'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import { PenSquare } from 'lucide-vue-next'

const auth = useAuthStore()
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Feed</h1>
      <Button v-if="auth.isLoggedIn" as-child size="sm">
        <RouterLink to="/posts/new">
          <PenSquare class="h-4 w-4 mr-1" /> New Post
        </RouterLink>
      </Button>
    </div>
    <PostFeed />
  </div>
</template>

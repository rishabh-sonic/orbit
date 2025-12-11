<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import PostEditor from '@/components/post/PostEditor.vue'
import { Skeleton } from '@/components/ui/skeleton'

const route = useRoute()
const router = useRouter()
const { toast } = useToast()

const title = ref('')
const content = ref('')
const loading = ref(false)
const fetching = ref(true)

onMounted(async () => {
  try {
    const data = await api.get<{ title: string; content: string }>(`/posts/${route.params.id}`)
    title.value = data.title
    content.value = data.content
  } catch {
    toast({ variant: 'destructive', title: 'Failed to load post' })
    router.back()
  } finally {
    fetching.value = false
  }
})

async function submit() {
  loading.value = true
  try {
    await api.put(`/posts/${route.params.id}`, { title: title.value, content: content.value })
    router.push(`/posts/${route.params.id}`)
  } catch (e) {
    toast({
      variant: 'destructive',
      title: 'Failed to update post',
      description: e instanceof ApiError ? e.message : 'Please try again.',
    })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">Edit Post</h1>
    <div v-if="fetching" class="space-y-3">
      <Skeleton class="h-9 w-full" />
      <Skeleton class="h-40 w-full" />
    </div>
    <PostEditor
      v-else
      v-model:title="title"
      v-model:content="content"
      :loading="loading"
      submit-label="Save changes"
      @submit="submit"
      @cancel="router.back()"
    />
  </div>
</template>

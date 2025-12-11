<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '@/lib/api'
import { useToast } from '@/components/ui/toast'
import PostEditor from '@/components/post/PostEditor.vue'

const router = useRouter()
const { toast } = useToast()

const title = ref('')
const content = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    const data = await api.post<{ id: string }>('/posts', { title: title.value, content: content.value })
    router.push(`/posts/${data.id}`)
  } catch (e) {
    toast({
      variant: 'destructive',
      title: 'Failed to create post',
      description: e instanceof ApiError ? e.message : 'Please try again.',
    })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">New Post</h1>
    <PostEditor
      v-model:title="title"
      v-model:content="content"
      :loading="loading"
      submit-label="Publish"
      @submit="submit"
      @cancel="router.back()"
    />
  </div>
</template>

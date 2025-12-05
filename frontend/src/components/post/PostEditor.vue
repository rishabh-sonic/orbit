<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'

const props = defineProps<{
  title: string
  content: string
  loading?: boolean
  submitLabel?: string
}>()

const emit = defineEmits<{
  'update:title': [v: string]
  'update:content': [v: string]
  submit: []
  cancel: []
}>()
</script>

<template>
  <form @submit.prevent="emit('submit')" class="space-y-4">
    <div class="space-y-1">
      <Label>Title</Label>
      <Input
        :model-value="title"
        @update:model-value="emit('update:title', $event as string)"
        placeholder="Post title"
        required
        maxlength="300"
      />
    </div>
    <div class="space-y-1">
      <Label>Content</Label>
      <Textarea
        :model-value="content"
        @update:model-value="emit('update:content', $event as string)"
        placeholder="Write your post…"
        rows="12"
        required
        class="resize-y min-h-[200px]"
      />
    </div>
    <div class="flex gap-2 justify-end">
      <Button type="button" variant="outline" @click="emit('cancel')">Cancel</Button>
      <Button type="submit" :disabled="loading">
        {{ loading ? 'Saving…' : (submitLabel ?? 'Submit') }}
      </Button>
    </div>
  </form>
</template>
